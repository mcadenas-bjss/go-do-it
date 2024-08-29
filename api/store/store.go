package store

import (
	"context"
	"database/sql"
	"fmt"
	"log"
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
  time TEXT,
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
	db             *sql.DB
	CommandChannel chan Command
	lock           sync.RWMutex
}

type Todo struct {
	Id          int
	Time        string
	Description string
	Completed   bool
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
	Cmd     CommandType
	Ctx     context.Context
	Payload interface{}
	Reply   chan interface{}
	Err     chan error
}

func (dts *DbTodoStore) StartManager() chan<- Command {
	cmds := make(chan Command)

	go func() {
		for cmd := range cmds {
			switch cmd.Cmd {
			case GetCommand:
				if todo, err := dts.get(cmd.Ctx, cmd.Payload.(int)); err != nil {
					cmd.Err <- err
				} else {
					cmd.Reply <- todo
				}
			case GetAllCommand:
				if todos, err := dts.all(cmd.Ctx); err != nil {
					cmd.Err <- err
				} else {
					cmd.Reply <- todos
				}
			case InsertCommand:
				if id, err := dts.insert(cmd.Ctx, cmd.Payload.(Todo)); err != nil {
					cmd.Err <- err
				} else {
					cmd.Reply <- id
				}
			case UpdateCommand:
				if ok, err := dts.update(cmd.Ctx, cmd.Payload.(Todo)); err != nil {
					cmd.Err <- err
				} else {
					cmd.Reply <- ok
				}
			case DeleteCommand:
				if ok, err := dts.delete(cmd.Ctx, cmd.Payload.(int)); err != nil {
					cmd.Err <- err
				} else {
					cmd.Reply <- ok
				}
			case ToggleCommand:
				if ok, err := dts.toggle(cmd.Ctx, cmd.Payload.(int)); err != nil {
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

func withContext[T any](ctx context.Context, def func() (T, error)) (T, error) {
	data := make(chan T)
	e := make(chan error)

	go func() {
		var result T
		select {
		case <-ctx.Done():
			return
		default:
			if res, err := def(); err != nil {
				e <- err
				close(e)
				return
			} else {
				result = res
			}
		}
		data <- result
		close(data)
	}()

	var t T
	select {
	case <-ctx.Done():
		logger.Error("Connection closed")
		return t, ctx.Err()
	case result := <-data:
		logger.Info(fmt.Sprintf("%+v", result))
		return result, nil
	case err := <-e:
		logger.Info(fmt.Sprintf("%v", err))
		return t, err
	}
}

func (dts *DbTodoStore) get(ctx context.Context, id int) (Todo, error) {
	logger.Info(fmt.Sprintf("Getting todo item: %d", id))

	return withContext(ctx, func() (Todo, error) {
		row := dts.db.QueryRowContext(ctx, "SELECT * FROM todo WHERE id=?", id)
		todo := Todo{}
		if err := row.Scan(&todo.Id, &todo.Time, &todo.Description, &todo.Completed); err != nil {
			return Todo{}, errors.Wrap(err, "Id not found")
		}

		return todo, nil
	})

}

func (dts *DbTodoStore) all(ctx context.Context) ([]Todo, error) {
	logger.Info("Getting all todos")

	return withContext(ctx, func() ([]Todo, error) {
		rows, err := dts.db.QueryContext(ctx, "SELECT * FROM todo")

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
	})
}

func (t *DbTodoStore) insert(ctx context.Context, todo Todo) (int, error) {
	logger.Info("Inserting todo", todo)

	return withContext(ctx, func() (int, error) {
		t.lock.Lock()
		defer t.lock.Unlock()
		res, err := t.db.ExecContext(ctx, "INSERT INTO todo VALUES(NULL,?,?,?);", todo.Time, todo.Description, todo.Completed)
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
	})

}

func (d *DbTodoStore) update(ctx context.Context, todo Todo) (bool, error) {
	logger.Info(fmt.Sprintf("Updating todo %+v", todo))

	return withContext(ctx, func() (bool, error) {
		res, err := d.db.ExecContext(ctx, "UPDATE todo SET time=?, description=? WHERE id=?", todo.Time, todo.Description, todo.Id)
		if err != nil {
			logger.Info("Error: %s", err)
			return false, errors.Wrap(err, "Update failed")
		}
		if n, e := res.RowsAffected(); n != 1 {
			logger.Info("Error: %s", e)
			return false, e
		}
		return true, nil
	})

}

func (d *DbTodoStore) delete(ctx context.Context, id int) (bool, error) {
	logger.Info(fmt.Sprintf("Deleting todo %d", id))
	return withContext(ctx, func() (bool, error) {

		res, err := d.db.ExecContext(ctx, "DELETE FROM todo WHERE id=?", id)
		if err != nil {
			logger.Error("Error: %s", err)
			return false, err
		}
		if n, e := res.RowsAffected(); n > 1 {
			logger.Error("Error: %s", e)
			return false, e
		}
		return true, nil
	})
}

func (d *DbTodoStore) toggle(ctx context.Context, id int) (bool, error) {
	logger.Info(fmt.Sprintf("Toggling complete status for todo %d", id))
	return withContext(ctx, func() (bool, error) {
		todo, e := d.get(ctx, id)
		if e != nil {
			logger.Error("Error: %s", e)
			return false, e
		}
		res, err := d.db.ExecContext(ctx, "UPDATE todo SET completed=? WHERE id=?", !todo.Completed, id)
		if err != nil {
			logger.Error("Error: %s", err)
			return false, err
		}
		if n, e := res.RowsAffected(); n != 1 {
			logger.Error("Error: %s", e)
			return false, e
		}
		return true, nil
	})
}

func prepareGet(db *sql.DB) (*sql.Stmt, error) {
	return db.Prepare("SELECT * FROM todo WHERE id=?")
}
