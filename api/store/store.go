package store

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"sync"

	_ "github.com/mattn/go-sqlite3"
	"github.com/mcadenas-bjss/go-do-it/api/logging"
	"github.com/pkg/errors"
)

// const file string = "../todo.db"
var logger = logging.NewLogger(0)

const create string = `
  CREATE TABLE IF NOT EXISTS todo (
  id INTEGER NOT NULL PRIMARY KEY,
  time DATETIME NOT NULL,
  description TEXT,
  completed BOOLEAN NOT NULL DEFAULT FALSE
  );`

func NewDbTodoStore(file string) (*DbTodoStore, error) {
	env := os.Getenv("env")
	logger.Info("Running on " + env)
	logger.Info(fmt.Sprintf("Opening sqlite file at %s", file))
	db, err := sql.Open("sqlite3", file)
	if err != nil {
		return nil, err
	}
	if _, err := db.Exec(create); err != nil {
		return nil, err
	}

	insert, err := db.Prepare("INSERT INTO todo VALUES(NULL,?,?,?);")
	if err != nil {
		return nil, err
	}
	defer insert.Close()

	list, err := db.Prepare("SELECT * FROM todo")
	if err != nil {
		return nil, err
	}
	defer list.Close()

	// Seed data if db empty
	if rows, err := list.Query(); err == nil && env != "test" {
		if !rows.Next() {
			logger.Info("Seeding data")
			insert.Exec("2020-01-01T00:00:00Z", "test", false)
		}
		rows.Close()
	} else {
		return nil, err
	}

	return &DbTodoStore{
		db:   db,
		lock: sync.RWMutex{},
	}, nil
}

type DbTodoStore struct {
	db   *sql.DB
	lock sync.RWMutex
}

type Todo struct {
	Id          int
	Time        string
	Description string
	Completed   bool
}

func (dts *DbTodoStore) Get(ctx context.Context, id int) (Todo, error) {
	logger.Info(fmt.Sprintf("Getting todo item: %d", id))

	data := make(chan Todo, 1)
	e := make(chan error, 1)

	go func() {
		var result Todo
		select {
		case <-ctx.Done():
			return
		default:
			row := dts.db.QueryRowContext(ctx, "SELECT * FROM todo WHERE id=?", id)
			todo := Todo{}
			if err := row.Scan(&todo.Id, &todo.Time, &todo.Description, &todo.Completed); err == sql.ErrNoRows {
				e <- errors.Wrap(err, "Id not found")
				return
			} else {
				result = todo
			}
		}
		data <- result
	}()

	select {
	case <-ctx.Done():
		logger.Error("Connection closed")
		return Todo{}, ctx.Err()
	case result := <-data:
		logger.Info(fmt.Sprintf("Found todo item: %+v", result))
		return result, nil
	case err := <-e:
		logger.Error("Id not found", err)
		return Todo{}, err
	}
}

func (dts *DbTodoStore) All() ([]Todo, error) {
	logger.Info("Getting all todos")

	rows, err := dts.db.Query("SELECT * FROM todo")

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	todos := []Todo{}

	for rows.Next() {
		todo := Todo{}
		if err := rows.Scan(&todo.Id, &todo.Time, &todo.Description, &todo.Completed); err != nil {
			return nil, errors.Wrap(err, "Error scanning row")
		}
		todos = append(todos, todo)
	}

	logger.Info(fmt.Sprintf("Found %d items", len(todos)))
	return todos, nil
}

func (t *DbTodoStore) Insert(todo Todo) (int, error) {
	logger.Info("Inserting todo: %+v", todo)
	res, err := t.db.Exec("INSERT INTO todo VALUES(NULL,?,?,?);", todo.Time, todo.Description, todo.Completed)
	if err != nil {
		logger.Error("Error: %s", err)
		return 0, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		logger.Error("Error: %s", err)
		return 0, err
	}

	return int(id), nil
}

func (d *DbTodoStore) Update(id int, todo Todo) (bool, error) {
	logger.Info("Updating todo %d with %+v", id, todo)

	_, err := d.db.Exec("UPDATE todo SET time=?, description=? WHERE id=?", todo.Time, todo.Description, id)

	if err != nil {
		logger.Info("Error: %s", err)
		return false, errors.Wrap(err, "Update failed")
	}

	return true, nil
}

func (d *DbTodoStore) Delete(id int) error {
	_, err := d.db.Exec("DELETE FROM todo WHERE id=?", id)
	return err
}

func prepareGet(db *sql.DB) (*sql.Stmt, error) {
	return db.Prepare("SELECT * FROM todo WHERE id=?")
}
