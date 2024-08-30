package main

import (
	"flag"
	"log"
	"net/http"
	"strconv"

	"github.com/mcadenas-bjss/go-do-it/logging"
	"github.com/mcadenas-bjss/go-do-it/server"
	"github.com/mcadenas-bjss/go-do-it/store"
)

var logger = logging.NewLogger(0)

func main() {
	var port, logLevel int
	var db string

	// Get the command line arguments
	flag.IntVar(&port, "port", 8000, "Port number")
	flag.IntVar(&logLevel, "logLevel", 0, "Log level")
	flag.StringVar(&db, "db", "todo.db", "Database file path")

	flag.Parse()

	logger.SetLevel(logLevel)

	dataStore, err := store.NewDbTodoStore(db)

	if err != nil {
		panic(err)
	}

	logger.Info("Starting server on port " + strconv.Itoa(port))
	server := server.NewTodoServer(dataStore)
	log.Fatal(http.ListenAndServe("localhost:"+strconv.Itoa(port), server))
}
