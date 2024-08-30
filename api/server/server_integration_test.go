package server_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"

	"github.com/mcadenas-bjss/go-do-it/server"
	"github.com/mcadenas-bjss/go-do-it/store"
)

const DBConnection = "file:test1?mode=memory&cache=shared"

const expectedJson = `<li id="todo-1">
  <div class="todo">
    <input id="todo-1-checkbox" type="checkbox" ZgotmplZ />
    <p>test todo</p>
    <button
      hx-delete="/api/todo/1"
      hx-swap="delete"
      hx-target="#todo-1">Delete</button
    >
    <div class="meta">
      <time datetime=2024-01-01T00:00:00.000Z>Mon, 01 Jan 2024 00:00</time>
    </div>
  </div>
</li>`

func TestInsertingTodoItemsAndRetrievingThem(t *testing.T) {
	os.Setenv("env", "test")
	dbStore, err := store.NewDbTodoStore(DBConnection)
	if err != nil {
		t.Error(err)
	}

	defer dbStore.Close()

	srv := *server.NewTodoServer(dbStore)

	newTodo := &store.Todo{
		Id:          1, // This ID is ignored in the insert operation as the db creates them sequentially.
		Time:        "2024-01-01T00:00:00.000Z",
		Description: "test todo",
		Completed:   false,
	}

	// insert
	insertResp := httptest.NewRecorder()
	srv.ServeHTTP(insertResp, NewPostTodoRequest(*newTodo))
	assertHtml(t, insertResp.Body, expectedJson) // assert html response

	// get
	response := httptest.NewRecorder()
	srv.ServeHTTP(response, NewGetTodoRequest(1))
	assertStatus(t, response.Code, http.StatusOK)
	assertTodo(t, response.Body, *newTodo)
}

func assertTodo(t testing.TB, actual *bytes.Buffer, expected store.Todo) {
	t.Helper()

	var td store.Todo
	json.NewDecoder(actual).Decode(&td)

	if !reflect.DeepEqual(td, expected) {
		t.Errorf("got %v want %v", td, expected)
	}
}

func assertHtml(t testing.TB, actual *bytes.Buffer, expected string) {
	t.Helper()

	if actual.String() != expected {
		t.Errorf("got %v want %v", actual, expected)
	}
}
