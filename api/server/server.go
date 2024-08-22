package server

import (
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
	Get(id int) (store.Todo, error)
	All() ([]store.Todo, error)
	Insert(todo store.Todo) error
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
	router.Handle("/health", http.HandlerFunc(t.healthHandler))
	router.Handle("/todo/", http.HandlerFunc(t.todoHandler))
	router.Handle("/todos/", http.HandlerFunc(t.todosHandler))

	t.Handler = router

	return t
}

func (t *TodoServer) healthHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("health check")
	w.Header().Set("content-type", jsonContentType)
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, `{"status": "ok"}`)
}

func (t *TodoServer) todoHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("%s %s", r.Method, r.URL.Path)

	switch r.Method {
	case http.MethodPost:
		t.Insert(w, r)
	case http.MethodGet:
		t.fetchOne(w, r)
	}
}

func (t *TodoServer) todosHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("%s %s", r.Method, r.URL.Path)

	switch r.Method {
	case http.MethodGet:
		t.fetchAll(w)
	}
}

func (t *TodoServer) fetchOne(w http.ResponseWriter, r *http.Request) {
	id, err := getId(r.URL.Path)
	if err != nil {
		log.Println(errors.Wrap(err, "failed to get id from path"))
		w.WriteHeader(http.StatusBadRequest)
	}

	todo, err := t.store.Get(id)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
	}

	json.NewEncoder(w).Encode(todo)
	w.WriteHeader(http.StatusOK)
}

func (t *TodoServer) fetchAll(w http.ResponseWriter) {
	log.Println("fetch all")

	todos, err := t.store.All()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	log.Printf("Found %d items", len(todos))
	okWithData(w, todos)
}

func (t *TodoServer) Insert(w http.ResponseWriter, r *http.Request) {
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

	err = t.store.Insert(todo)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func getId(path string) (int, error) {
	return strconv.Atoi(strings.TrimPrefix(path, "/todo/"))
}

func okWithData[T any](w http.ResponseWriter, data T) {
	json.NewEncoder(w).Encode(data)
	w.WriteHeader(http.StatusOK)
}
