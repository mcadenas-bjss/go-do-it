package main

import (
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/mcadenas-bjss/go-do-it/api/server"
	"github.com/mcadenas-bjss/go-do-it/api/store"
)

func main() {

	store, err := store.NewDbTodoStore()

	if err != nil {
		panic(err)
	}

	portnum := 8000
	if len(os.Args) > 1 {
		portnum, _ = strconv.Atoi(os.Args[1])
	}
	log.Printf("Going to listen on port %d\n", portnum)

	server := server.NewTodoServer(store)
	log.Fatal(http.ListenAndServe("localhost:"+strconv.Itoa(portnum), server))
}
