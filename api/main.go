package main

import (
	"flag"
	"log"
	"net/http"
	"strconv"

	"github.com/mcadenas-bjss/go-do-it/logger"
	"github.com/mcadenas-bjss/go-do-it/server"
	"github.com/mcadenas-bjss/go-do-it/store"
)

func main() {
	log := logger.NewLogger(nil)
	log.SetLevel(logger.Info)
	var port, logLevel int
	var db string

	// Get the command line arguments
	flag.IntVar(&port, "port", 8000, "Port number")
	flag.IntVar(&logLevel, "logLevel", 0, "Log level")
	flag.StringVar(&db, "db", "todo.db", "Database file path")

	flag.Parse()

	// logger.SetLevel(logLevel)

	dataStore, err := store.NewDbTodoStore(db)

	if err != nil {
		panic(err)
	}

	log.Info("Starting server on port " + strconv.Itoa(port))
	server := server.NewTodoServer(dataStore)
	if err := http.ListenAndServe("localhost:"+strconv.Itoa(port), server); err != nil {
		log.Error(err)
	}
}

func setUpLogging() {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds | log.Llongfile)
}
