package store

import (
	"database/sql"
	"log"
	"sync"

	_ "github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"
)

const file string = "../todo.db"

const create string = `
  CREATE TABLE IF NOT EXISTS todo (
  id INTEGER NOT NULL PRIMARY KEY,
  time DATETIME NOT NULL,
  description TEXT
  );`

func NewDbTodoStore() (*DbTodoStore, error) {
	db, err := sql.Open("sqlite3", file)
	if err != nil {
		return nil, err
	}
	if _, err := db.Exec(create); err != nil {
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

func (dts *DbTodoStore) Get(id int) (Todo, error) {
	log.Printf("Getting todo item: %d", id)

	// dts.lock.RLock()
	// defer dts.lock.RUnlock()

	row := dts.db.QueryRow("SELECT id, time, description FROM todo WHERE id=?", id)

	todo := Todo{}
	var err error
	if err = row.Scan(&todo.Id, &todo.Time, &todo.Description); err == sql.ErrNoRows {
		log.Printf("ID not found")
		return Todo{}, errors.Wrap(err, "Row not found")
	}

	return todo, err
}

func (t *DbTodoStore) All() ([]Todo, error) {
	log.Printf("Getting all todos")

	rows, err := t.db.Query("SELECT id, time, description FROM todo")

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	todos := []Todo{}

	for rows.Next() {
		todo := Todo{}
		if err := rows.Scan(&todo.Id, &todo.Time, &todo.Description); err != nil {
			return nil, errors.Wrap(err, "Error scanning row")
		}
		todos = append(todos, todo)
	}

	log.Printf("Found %d items", len(todos))
	return todos, nil
}

func (t *DbTodoStore) Insert(todo Todo) error {
	log.Printf("Inserting todo: %+v", todo)
	res, err := t.db.Exec("INSERT INTO todo VALUES(NULL,?,?);", todo.Time, todo.Description)
	if err != nil {
		log.Printf("Error: %s", err)
		return err
	}

	if _, err := res.LastInsertId(); err != nil {
		log.Printf("Error: %s", err)
		return err
	}
	return nil
}

func (d *DbTodoStore) Update(id int, todo Todo) (bool, error) {
	log.Printf("Updating todo %d with %+v", id, todo)

	_, err := d.db.Exec("UPDATE todo SET time=?, description=? WHERE id=?", todo.Time, todo.Description, id)

	if err != nil {
		log.Printf("Error: %s", err)
		return false, errors.Wrap(err, "Update failed")
	}

	return true, nil
}

func (d *DbTodoStore) Delete(id int) error {
	_, err := d.db.Exec("DELETE FROM todo WHERE id=?", id)
	return err
}
