package main

import (
	"fmt"
	"syscall/js"
	"time"
)

func main() {
	c := make(chan interface{})
	fmt.Println("Hello, WebAssembly!")

	js.Global().Set("sayHello", js.FuncOf(sayHello))

	<-c
}

func sayHello(this js.Value, args []js.Value) any {
	ch := make(chan interface{})
	go func() {
		name := args[0].String()
		time.Sleep(5 * time.Second)
		ch <- fmt.Sprintf("Hello, %s!", name)
	}()

	return <-ch
}
