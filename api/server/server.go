package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/mcadenas-bjss/go-do-it/api/store"
	"github.com/mcadenas-bjss/go-do-it/api/views"
	"github.com/pkg/errors"
)

type TodoStore interface {
	StartManager() chan<- store.Command
}

type TodoServer struct {
	store TodoStore
	http.Handler
	cmds     chan<- store.Command
	renderer views.TodoRenderer
}

const jsonContentType = "application/json"
const htmlContentType = "text/html"
const (
	HEALTH_PATH    = "GET /api/health"
	TODO_ID_PATH   = "/api/todo/{id}"
	POST_TODO_PATH = "POST /api/todo"
	GET_TODOS_PATH = "GET /api/todos"
)

func NewTodoServer(store TodoStore) *TodoServer {
	t := new(TodoServer)

	t.store = store

	renderer, err := views.NewTodoRenderer()
	if err != nil {
		log.Println("failed to create renderer")
		log.Println(err)
		panic(err)
	}
	t.renderer = *renderer

	t.cmds = t.store.StartManager()

	router := http.NewServeMux()

	// API CRUD
	router.Handle(HEALTH_PATH, http.HandlerFunc(t.healthHandler))
	router.Handle(fmt.Sprintf("GET %s", TODO_ID_PATH), http.HandlerFunc(t.handleGetTodo))
	router.Handle(POST_TODO_PATH, http.HandlerFunc(t.handlePostTodo))
	router.Handle(fmt.Sprintf("DELETE %s", TODO_ID_PATH), http.HandlerFunc(t.handleDeleteTodo))
	router.Handle(fmt.Sprintf("PUT %s", TODO_ID_PATH), http.HandlerFunc(t.handlePutTodo))
	router.Handle(GET_TODOS_PATH, http.HandlerFunc(t.handleGetAllTodo))

	// Partials
	router.Handle("POST /api/todo/toggle/{id}", http.HandlerFunc(t.handleToggleCompleteState))

	t.Handler = router

	return t
}

func (t *TodoServer) healthHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("health check")
	w.Header().Set("content-type", jsonContentType)
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, `{"status": "ok"}`)
}

func (t *TodoServer) handleGetTodo(w http.ResponseWriter, r *http.Request) {
	log.Printf("%s %s", r.Method, r.URL.Path)
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		log.Println(errors.Wrap(err, "failed to get id from path"))
		w.WriteHeader(http.StatusBadRequest)
	}

	errChannel := make(chan error)
	replyChan := make(chan interface{})
	t.cmds <- store.Command{Cmd: store.GetCommand, Ctx: r.Context(), Payload: id, Reply: replyChan, Err: errChannel}

	select {
	case err := <-errChannel:
		switch err.Error() {
		case "todo not found":
			w.WriteHeader(http.StatusNotFound)
			return
		default:
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	case reply := <-replyChan:
		json.NewEncoder(w).Encode(reply)
	}
}

func (t *TodoServer) handleGetAllTodo(w http.ResponseWriter, r *http.Request) {
	log.Printf("%s %s", r.Method, r.URL.Path)

	errChan := make(chan error)
	replyChan := make(chan interface{})
	t.cmds <- store.Command{Cmd: store.GetAllCommand, Ctx: r.Context(), Payload: nil, Reply: replyChan, Err: errChan}

	select {
	case err := <-errChan:
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	case reply := <-replyChan:
		json.NewEncoder(w).Encode(reply)
	}
}

func (t *TodoServer) handlePostTodo(w http.ResponseWriter, r *http.Request) {
	log.Printf("%s %s", r.Method, r.URL.Path)

	var todo store.Todo
	err := decodeJSONBody(w, r, &todo)
	if err != nil {
		var mr *malformedRequest
		if errors.As(err, &mr) {
			http.Error(w, mr.msg, mr.status)
		} else {
			log.Print(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
		return
	}

	errChannel := make(chan error)
	replyChan := make(chan interface{})
	t.cmds <- store.Command{Cmd: store.InsertCommand, Ctx: r.Context(), Payload: todo, Reply: replyChan, Err: errChannel}

	select {
	case err := <-errChannel:
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	case reply := <-replyChan:
		newTodo := store.Todo{Id: reply.(int), Time: todo.Time, Description: todo.Description, Completed: todo.Completed}
		if err := t.renderer.RenderTodo(w, newTodo); err != nil {
			log.Printf("failed to render todo: %s", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
}

func (t *TodoServer) handlePutTodo(w http.ResponseWriter, r *http.Request) {
	log.Printf("%s %s", r.Method, r.URL.Path)

	var todo store.Todo
	err := decodeJSONBody(w, r, &todo)
	if err != nil {
		var mr *malformedRequest
		if errors.As(err, &mr) {
			http.Error(w, mr.msg, mr.status)
		} else {
			log.Print(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
		return
	}

	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		log.Println(errors.Wrap(err, "failed to get id from path"))
		w.WriteHeader(http.StatusBadRequest)
	}

	todo.Id = id

	errChannel := make(chan error)
	replyChan := make(chan interface{})
	t.cmds <- store.Command{Cmd: store.UpdateCommand, Ctx: r.Context(), Payload: todo, Reply: replyChan, Err: errChannel}

	select {
	case err := <-errChannel:
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	case reply := <-replyChan:
		json.NewEncoder(w).Encode(reply)
	}
}

func (t *TodoServer) handleDeleteTodo(w http.ResponseWriter, r *http.Request) {
	log.Printf("%s %s", r.Method, r.URL.Path)

	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		log.Println(errors.Wrap(err, "failed to get id from path"))
		w.WriteHeader(http.StatusBadRequest)
	}

	errChannel := make(chan error)
	replyChan := make(chan interface{})
	t.cmds <- store.Command{Cmd: store.DeleteCommand, Ctx: r.Context(), Payload: id, Reply: replyChan, Err: errChannel}

	select {
	case err := <-errChannel:
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	case reply := <-replyChan:
		json.NewEncoder(w).Encode(reply)
	}
}

func (t *TodoServer) handleToggleCompleteState(w http.ResponseWriter, r *http.Request) {
	log.Printf("%s %s", r.Method, r.URL.Path)

	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		log.Println(errors.Wrap(err, "failed to get id from path"))
		w.WriteHeader(http.StatusBadRequest)
	}

	errChannel := make(chan error)
	replyChan := make(chan interface{})
	t.cmds <- store.Command{Cmd: store.ToggleCommand, Ctx: r.Context(), Payload: id, Reply: replyChan, Err: errChannel}

	select {
	case err := <-errChannel:
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	case reply := <-replyChan:
		json.NewEncoder(w).Encode(reply)
	}
}
