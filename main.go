package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/ezfroze/go_final_project/pkg/api"
	"github.com/ezfroze/go_final_project/pkg/db"
)

var dbFileDefault = "scheduler.db"

func main() {
	http.Handle("/", http.FileServer(http.Dir("./web")))
	api.Init()

	port := os.Getenv("TODO_PORT")
	dbfile := os.Getenv("TODO_DBFILE")

	if port == "" {
		port = "7540"
	}

	if dbfile == "" {
		dbfile = dbFileDefault
	}

	database, err := db.Init(dbfile)

	if err != nil {
		log.Fatal(err)
	}
	defer database.Close()

	err = http.ListenAndServe(fmt.Sprintf(":%s", port), nil)
	if err != nil {
		log.Fatal(err)
	}
}
