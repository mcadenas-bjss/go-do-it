package store

import "sync"

type Store struct {
	db   map[int]Todo
	lock sync.RWMutex
}

type Todo struct {
	ID        int    `json:"id"`
	Title     string `json:"title"`
	Completed bool   `json:"completed"`
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
		return ok
	}

	s.db[todo.ID] = todo

	return true
}

func (s *Store) UpdateTodo(todo Todo) bool {
	s.lock.Lock()
	defer s.lock.Unlock()

	if ok := s.todoExists(todo.ID); !ok {
		return ok
	}
	s.db[todo.ID] = todo
	return true
}

func (s *Store) DeleteTodo(id int) bool {
	s.lock.Lock()
	defer s.lock.Unlock()
	if ok := s.todoExists(id); !ok {
		return ok
	}

	delete(s.db, id)

	return true
}

func (s *Store) ClearTodos() {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.db = make(map[int]Todo)
}

func (s *Store) ToggleTodo(id int) bool {
	s.lock.Lock()
	defer s.lock.Unlock()
	if todo, ok := s.db[id]; !ok {
		return ok
	} else {
		todo.Completed = !todo.Completed
		s.db[id] = todo
	}
	return true
}

func (s *Store) CountTodos() int {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return len(s.db)
}

func (s *Store) todoExists(id int) bool {
	_, ok := s.db[id]
	return ok
}

func NewStore() *Store {
	return &Store{
		db: make(map[int]Todo),
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
