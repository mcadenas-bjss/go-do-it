package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/mcadenas-bjss/go-do-it/api/store"
	"github.com/pkg/errors"
)

type TodoStore interface {
	Get(ctx context.Context, id int) (store.Todo, error)
	All() ([]store.Todo, error)
	Insert(todo store.Todo) (int, error)
	Update(id int, todo store.Todo) (bool, error)
	Delete(id int) error
}

type TodoServer struct {
	store TodoStore
	http.Handler
}

const jsonContentType = "application/json"

func NewTodoServer(store TodoStore) *TodoServer {
	t := new(TodoServer)

	t.store = store

	router := http.NewServeMux()
	router.Handle("GET /health", http.HandlerFunc(t.healthHandler))
	router.Handle("GET /todo/{id}", http.HandlerFunc(t.handleGetTodo))
	router.Handle("POST /todo", http.HandlerFunc(t.handlePostTodo))
	router.Handle("GET /todos", http.HandlerFunc(t.handleGetAllTodo))

	t.Handler = router

	return t
}

func (t *TodoServer) healthHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("health check")
	w.Header().Set("content-type", jsonContentType)
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, `{"status": "ok"}`)
}

// func (t *TodoServer) todoHandler(w http.ResponseWriter, r *http.Request) {
// 	log.Printf("%s %s", r.Method, r.URL.Path)

// 	switch r.Method {
// 	case http.MethodPost:
// 		t.Insert(w, r)
// 	case http.MethodGet:
// 		t.fetchOne(w, r)
// 	}
// }

// func (t *TodoServer) todosHandler(w http.ResponseWriter, r *http.Request) {
// 	log.Printf("%s %s", r.Method, r.URL.Path)

// 	switch r.Method {
// 	case http.MethodGet:
// 		t.fetchAll(w)
// 	}
// }

func (t *TodoServer) handleGetTodo(w http.ResponseWriter, r *http.Request) {
	log.Printf("%s %s", r.Method, r.URL.Path)
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		log.Println(errors.Wrap(err, "failed to get id from path"))
		w.WriteHeader(http.StatusBadRequest)
	}

	todo, err := t.store.Get(r.Context(), id)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(todo)
}

func (t *TodoServer) handleGetAllTodo(w http.ResponseWriter, r *http.Request) {
	log.Printf("%s %s", r.Method, r.URL.Path)

	todos, err := t.store.All()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(todos)
}

func (t *TodoServer) handlePostTodo(w http.ResponseWriter, r *http.Request) {
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

	id, err := t.store.Insert(todo)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("content-type", jsonContentType)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(struct{ id int }{id: id})
}

func getId(path string) (int, error) {
	return strconv.Atoi(strings.TrimPrefix(path, "/todo/"))
}
