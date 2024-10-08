package server_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mcadenas-bjss/go-do-it/server"
	"github.com/mcadenas-bjss/go-do-it/store"
)

type StubStore struct {
	todos map[int]store.Todo
}

func (s *StubStore) StartManager() chan<- store.Command {
	cmds := make(chan store.Command)

	go func() {
		for cmd := range cmds {
			switch cmd.Cmd {
			case store.GetCommand:
				if todo, err := s.get(cmd.Ctx, cmd.Payload.(int)); err != nil {
					cmd.Err <- err
				} else {
					cmd.Reply <- todo
				}
			case store.GetAllCommand:
				if todos, err := s.all(cmd.Ctx); err != nil {
					cmd.Err <- err
				} else {
					cmd.Reply <- todos
				}
			case store.InsertCommand:
				if id, err := s.insert(cmd.Ctx, cmd.Payload.(store.Todo)); err != nil {
					cmd.Err <- err
				} else {
					cmd.Reply <- id
				}
			case store.UpdateCommand:
				if ok, err := s.update(cmd.Ctx, cmd.Payload.(store.Todo)); err != nil {
					cmd.Err <- err
				} else {
					cmd.Reply <- ok
				}
			case store.DeleteCommand:
				if ok, err := s.delete(cmd.Ctx, cmd.Payload.(int)); err != nil {
					cmd.Err <- err
				} else {
					cmd.Reply <- ok
				}
			case store.ToggleCommand:
				if ok, err := s.toggle(cmd.Ctx, cmd.Payload.(int)); err != nil {
					cmd.Err <- err
				} else {
					cmd.Reply <- ok
				}
			default:
				log.Fatal("unknown command type", cmd.Cmd)
			}
		}
	}()
	return cmds
}

func (s *StubStore) get(ctx context.Context, id int) (store.Todo, error) {
	data := make(chan store.Todo, 1)
	e := make(chan error, 1)

	go func() {
		var result store.Todo
		select {
		case <-ctx.Done():
			return
		default:
			if t, ok := s.todos[id]; !ok {
				e <- errors.New("todo not found")
				return
			} else {
				result = t
			}
		}
		data <- result
	}()

	select {
	case <-ctx.Done():
		return store.Todo{}, ctx.Err()
	case todo := <-data:
		return todo, nil
	case err := <-e:
		return store.Todo{}, err
	}
}

func (s *StubStore) all(ctx context.Context) ([]store.Todo, error) {
	v := make([]store.Todo, 0, len(s.todos))

	for _, value := range s.todos {
		v = append(v, value)
	}
	return v, nil
}

func (s *StubStore) insert(ctx context.Context, todo store.Todo) (int, error) {
	newId := len(s.todos) + 1
	s.todos[newId] = todo
	return newId, nil
}

func (s *StubStore) update(ctx context.Context, todo store.Todo) (bool, error) {
	return true, nil
}

func (s *StubStore) delete(ctx context.Context, id int) (bool, error) {
	return true, nil
}

func (s *StubStore) toggle(ctx context.Context, id int) (bool, error) {
	return true, nil
}

func TestHealth(t *testing.T) {
	server := server.NewTodoServer(&StubStore{})

	t.Run("it returns 200 on /health", func(t *testing.T) {
		request, _ := http.NewRequest("GET", "/api/health", nil)
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		var got struct{ status string }

		err := json.NewDecoder(response.Body).Decode(&got)
		if err != nil {
			t.Fatalf("Unable to parse response from server %q into health check body, '%v'", response.Body, err)
		}

		assertStatus(t, response.Code, http.StatusOK)
	})
}

func TestCRUD(t *testing.T) {
	stubStore := StubStore{
		map[int]store.Todo{
			1: {Id: 1, Time: "2024-01-01T00:00:00.000Z", Description: "Buy milk", Completed: false},
			2: {Id: 2, Time: "2024-01-01T00:00:00.000Z", Description: "Buy bread", Completed: true},
		},
	}

	todoServer := server.NewTodoServer(&stubStore)

	t.Run("Get existing todo", func(t *testing.T) {
		request := NewGetTodoRequest(1)
		response := httptest.NewRecorder()

		todoServer.ServeHTTP(response, request)

		var got store.Todo
		assertJson(t, response.Body, &got)
		assertStatus(t, response.Code, http.StatusOK)
	})

	t.Run("Get non existing todo", func(t *testing.T) {
		request := NewGetTodoRequest(3)
		response := httptest.NewRecorder()

		todoServer.ServeHTTP(response, request)

		assertStatus(t, response.Code, http.StatusNotFound)
	})

	t.Run("Get all returns array with items", func(t *testing.T) {
		request, _ := http.NewRequest("GET", "/api/todos", nil)
		response := httptest.NewRecorder()

		todoServer.ServeHTTP(response, request)

		var got []store.Todo
		assertJson(t, response.Body, &got)
		assertStatus(t, response.Code, http.StatusOK)

		if len(got) != 2 {
			t.Errorf("Expected 2 todos, got %d", len(got))
		}
	})

	t.Run("Get all returns empty array", func(t *testing.T) {
		emptyServer := server.NewTodoServer(&StubStore{
			make(map[int]store.Todo),
		})
		request, _ := http.NewRequest("GET", "/api/todos", nil)
		response := httptest.NewRecorder()

		emptyServer.ServeHTTP(response, request)

		var got []store.Todo
		assertJson(t, response.Body, &got)
		assertStatus(t, response.Code, http.StatusOK)

		if len(got) != 0 {
			t.Errorf("Expected 0 todos, got %d", len(got))
		}
	})

	t.Run("Insert", func(t *testing.T) {
		request := NewPostTodoRequest(store.Todo{Id: 3, Time: "2024-01-01T00:00:00.000Z", Description: "Buy butter", Completed: false})
		response := httptest.NewRecorder()

		todoServer.ServeHTTP(response, request)

		assertStatus(t, response.Code, http.StatusOK)

		if len(stubStore.todos) != 3 {
			t.Errorf("Expected 3 todos, got %d", len(stubStore.todos))
		}
	})
}

func NewPostTodoRequest(todo store.Todo) *http.Request {
	buff := bytes.Buffer{}
	json.NewEncoder(&buff).Encode(todo)
	req, _ := http.NewRequest(http.MethodPost, "/api/todo", &buff)
	return req
}

func NewGetTodoRequest(id int) *http.Request {
	req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/api/todo/%d", id), nil)
	return req
}

func assertStatus(t testing.TB, got, want int) {
	t.Helper()
	if got != want {
		t.Errorf("did not get correct status, got %d, want %d", got, want)
	}
}

func assertJson[T any](t testing.TB, j *bytes.Buffer, got T) {
	t.Helper()

	dec := json.NewDecoder(j)
	dec.DisallowUnknownFields()
	err := dec.Decode(&got)
	if err != nil {
		t.Fatalf("Unable to parse response from server %q into Todo, '%v'", j, err)
	}
}

func BenchmarkGet(b *testing.B) {
	stubStore := StubStore{
		map[int]store.Todo{
			1: {Id: 1, Time: "2024-01-01T00:00:00.000Z", Description: "Buy milk", Completed: false},
			2: {Id: 2, Time: "2024-01-01T00:00:00.000Z", Description: "Buy bread", Completed: true},
		},
	}

	todoServer := server.NewTodoServer(&stubStore)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		request, _ := http.NewRequest("GET", fmt.Sprintf("/api/todo/%d", 1), nil)
		response := httptest.NewRecorder()

		todoServer.ServeHTTP(response, request)
	}
}
