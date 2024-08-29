# go-do-it

My Golang ToDo App

## Server

Start the server by running `go run ./api/main.go`.
The server will create the sqlite file required.
Options:

- `-port` default is 8000
- `-db` default is "todo.db"

## WebApp

The web app is built with [Astro](https://docs.astro.build/en/getting-started/) and [HTMX](https://thevalleyofcode.com/htmx).
Using HTML allows for requests to the server to be made directly from html elements and use the response to swap out content. I use this to insert a new todo item into the DOM directly from the POST request response.

### Running the app

1.  `cd ./web`
2.  `npm install`
3.  `npm start` _**NOTE: API server should be running first because the app is SSR**_

## Local go Lang app

This is a small application made in go using the [fyne.io](https://docs.fyne.io/started/) library. To this point, the functionality is limited to Fetching the todo list and adding new items.
Due to time constraints I ignored a proper file structure and testing.

### Running the app

As with the web app, the API must be running in the background for this app to work as it communicates with the go back-end via http.

`go run ./app/main.go`

## Help Scripts

- `./scrips/insert.sh` inserts a dummy item
- `./scripts/get.sh 1` retrieves item with id 1
