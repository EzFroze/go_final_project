package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/ezfroze/go_final_project/pkg/db"
)

var dbFile = "scheduler.db"

func main() {
	http.Handle("/", http.FileServer(http.Dir("./web")))

	PORT := os.Getenv("TODO_PORT")
	TODO_DBFILE := os.Getenv("TODO_DBFILE")

	if PORT == "" {
		PORT = "7540"
	}

	if TODO_DBFILE == "" {
		TODO_DBFILE = dbFile
	}

	err := db.Init(TODO_DBFILE)

	if err != nil {
		panic(err)
	}

	err = http.ListenAndServe(fmt.Sprintf(":%s", PORT), nil)
	if err != nil {
		panic(err)
	}
}
