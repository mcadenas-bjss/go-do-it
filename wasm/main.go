package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"syscall/js"
	"time"
)

type Orchestrator struct {
	store Store
	cmds  chan<- Command
}

func main() {
	c := make(chan interface{})
	fmt.Println("Hello, WebAssembly!")

	orc := new(Orchestrator)
	orc.store = *NewStore()

	if ok := orc.store.AddTodo(Todo{ID: 1, Title: "Buy milk", Completed: false}); ok {
		fmt.Println("initial todo added")
	}
	orc.cmds = orc.store.StoreManager()

	if t, o := orc.store.GetTodo(1); !o {
		fmt.Println("todo not found")
	} else {
		encoder := json.NewEncoder(os.Stdout)
		fmt.Println(encoder.Encode(t))
	}

	js.Global().Set("getTodos", asyncFunc(orc.all))
	js.Global().Set("getTodo", asyncFunc(orc.get))
	js.Global().Set("addTodo", asyncFunc(orc.add))
	js.Global().Set("sayHelloInFive", asyncFunc(sayHelloInFive))
	js.Global().Set("sayHelloInTwo", asyncFunc(sayHelloInTwo))

	<-c
}

type fn func(this js.Value, args []js.Value) (any, error)

var (
	jsErr     js.Value = js.Global().Get("Error")
	jsPromise js.Value = js.Global().Get("Promise")
)

func asyncFunc(innerFunc fn) js.Func {
	return js.FuncOf(func(this js.Value, args []js.Value) any {
		handler := js.FuncOf(func(_ js.Value, promFn []js.Value) any {
			resolve, reject := promFn[0], promFn[1]

			go func() {
				defer func() {
					if r := recover(); r != nil {
						reject.Invoke(jsErr.New(fmt.Sprint("panic:", r)))
					}
				}()

				res, err := innerFunc(this, args)
				if err != nil {
					reject.Invoke(jsErr.New(err.Error()))
				} else {
					resolve.Invoke(res)
				}
			}()

			return nil
		})

		return jsPromise.New(handler)
	})
}

func sayHelloInTwo(this js.Value, args []js.Value) (any, error) {
	ch := make(chan interface{})
	go func() {
		name := args[0].String()
		time.Sleep(2 * time.Second)
		ch <- fmt.Sprintf("Hello, %s!", name)
	}()

	return <-ch, nil
}

func sayHelloInFive(this js.Value, args []js.Value) (any, error) {
	ch := make(chan interface{})
	go func() {
		name := args[0].String()
		time.Sleep(5 * time.Second)
		ch <- fmt.Sprintf("Hello, %s!", name)
	}()

	return <-ch, nil
}

func (o *Orchestrator) get(this js.Value, args []js.Value) (any, error) {
	if todo, ok := o.store.GetTodo(args[0].Int()); !ok {
		return nil, fmt.Errorf("todo with id %d not found", args[0].Int())
	} else {
		return todo.toJsObject(), nil
	}
}

func (o *Orchestrator) add(this js.Value, args []js.Value) (any, error) {
	if ok := o.store.AddTodo(Todo{ID: args[0].Int(), Title: args[1].String(), Completed: false}); !ok {
		return nil, fmt.Errorf("todo with id %d already exists", args[0].Int())
	} else {
		return true, nil
	}
}

func (o *Orchestrator) all(this js.Value, args []js.Value) (any, error) {
	todos := o.store.GetTodos()
	todoArr := make([]interface{}, len(todos))
	for i, todo := range todos {
		todoArr[i] = todo.toJsObject()
	}

	return todoArr, nil
}

// Store

type Store struct {
	db   map[int]Todo
	lock sync.RWMutex
}

type Todo struct {
	ID        int    `json:"id"`
	Title     string `json:"title"`
	Completed bool   `json:"completed"`
}

func (t *Todo) toJsObject() map[string]interface{} {
	return map[string]interface{}{
		"id":        t.ID,
		"title":     t.Title,
		"completed": t.Completed,
	}
}

func (s *Store) GetTodos() []Todo {
	s.lock.RLock()
	defer s.lock.RUnlock()
	todos := make([]Todo, 0, len(s.db))
	for _, todo := range s.db {
		todos = append(todos, todo)
	}
	return todos
}

func (s *Store) GetTodo(id int) (Todo, bool) {
	s.lock.RLock()
	defer s.lock.RUnlock()
	todo, ok := s.db[id]
	return todo, ok
}

func (s *Store) AddTodo(todo Todo) bool {
	s.lock.Lock()
	defer s.lock.Unlock()

	if ok := s.todoExists(todo.ID); !ok {
		s.db[todo.ID] = todo
		return true
	}

	return false
}

func (s *Store) UpdateTodo(todo Todo) bool {
	// s.lock.Lock()
	// defer s.lock.Unlock()

	if ok := s.todoExists(todo.ID); !ok {
		return ok
	}
	s.db[todo.ID] = todo
	return true
}

func (s *Store) DeleteTodo(id int) bool {
	// s.lock.Lock()
	// defer s.lock.Unlock()
	if ok := s.todoExists(id); !ok {
		return ok
	}

	delete(s.db, id)

	return true
}

func (s *Store) ClearTodos() {
	// s.lock.Lock()
	// defer s.lock.Unlock()
	s.db = make(map[int]Todo)
}

func (s *Store) ToggleTodo(id int) bool {
	// s.lock.Lock()
	// defer s.lock.Unlock()
	if todo, ok := s.db[id]; !ok {
		return ok
	} else {
		todo.Completed = !todo.Completed
		s.db[id] = todo
	}
	return true
}

func (s *Store) CountTodos() int {
	// s.lock.RLock()
	// defer s.lock.RUnlock()
	return len(s.db)
}

func (s *Store) todoExists(id int) bool {
	_, ok := s.db[id]
	return ok
}

func NewStore() *Store {
	return &Store{
		map[int]Todo{},
		sync.RWMutex{},
	}
}

type CommandType int

const (
	GetCommand = iota
	GetAllCommand
	InsertCommand
	UpdateCommand
	DeleteCommand
	ToggleCommand
)

type Command struct {
	Action  CommandType
	Payload interface{}
	Reply   chan interface{}
	Err     chan interface{}
}

func (s *Store) StoreManager() chan<- Command {
	ch := make(chan Command)
	go func() {
		for command := range ch {
			switch command.Action {
			case GetCommand:
				if todo, ok := s.GetTodo(command.Payload.(int)); !ok {
					command.Err <- ok
				} else {
					command.Reply <- todo
				}
			case GetAllCommand:
				todos := s.GetTodos()
				command.Reply <- todos
			case InsertCommand:
				if ok := s.AddTodo(command.Payload.(Todo)); !ok {
					command.Err <- ok
				} else {
					command.Reply <- true
				}
			case UpdateCommand:
				if ok := s.UpdateTodo(command.Payload.(Todo)); !ok {
					command.Err <- ok
				} else {
					command.Reply <- true
				}
			case DeleteCommand:
				if ok := s.DeleteTodo(command.Payload.(int)); !ok {
					command.Err <- ok
				} else {
					command.Reply <- true
				}
			case ToggleCommand:
				if ok := s.ToggleTodo(command.Payload.(int)); !ok {
					command.Err <- ok
				} else {
					command.Reply <- true
				}
			}
		}
	}()
	return ch
}
