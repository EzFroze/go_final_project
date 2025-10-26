package main

import (
	"fmt"
	"net/http"
	"os"
)

func main() {
	http.Handle("/", http.FileServer(http.Dir("./web")))

	PORT := os.Getenv("TODO_PORT")

	if PORT == "" {
		PORT = "7540"
	}

	err := http.ListenAndServe(fmt.Sprintf(":%s", PORT), nil)
	if err != nil {
		panic(err)
	}
}
