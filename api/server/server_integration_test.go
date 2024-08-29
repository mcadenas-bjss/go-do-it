package server_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"

	"github.com/mcadenas-bjss/go-do-it/api/server"
	"github.com/mcadenas-bjss/go-do-it/api/store"
)

const DBConnection = "file:test1?mode=memory&cache=shared"

func TestInsertingTodoItemsAndRetrievingThem(t *testing.T) {
	os.Setenv("env", "test")
	dbStore, err := store.NewDbTodoStore("file:../test.db?cache=shared")
	if err != nil {
		t.Error(err)
	}

	defer dbStore.Close()

	srv := *server.NewTodoServer(dbStore)

	// if _, err := os.Stat(testDbFile); err == nil {
	// 	defer os.Remove(testDbFile)
	// } else if errors.Is(err, os.ErrNotExist) {
	// 	t.Fatalf("DB file was not created")
	// }

	newTodo := &store.Todo{
		Id:          0,
		Time:        "2024-01-01T00:00:00.000Z",
		Description: "test todo",
		Completed:   false,
	}

	insertResp := httptest.NewRecorder()
	srv.ServeHTTP(insertResp, NewPostTodoRequest(*newTodo))
	var insertBody struct {
		id int
	}
	assertJson(t, insertResp.Body, insertBody)

	response := httptest.NewRecorder()
	srv.ServeHTTP(response, NewGetTodoRequest(insertBody.id))
	assertStatus(t, response.Code, http.StatusOK)
	assertTodo(t, response.Body, *newTodo)
}

func assertTodo(t testing.TB, actual *bytes.Buffer, expected store.Todo) {
	t.Helper()

	var td store.Todo
	json.NewDecoder(actual).Decode(&td)

	if !reflect.DeepEqual(td, expected) {
		t.Errorf("got %v want %v", actual, expected)
	}
}
