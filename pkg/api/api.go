package api

import (
	"errors"
	"net/http"
)

const DATEFORMAT = "20060102"

func taskHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		addTaskHandler(w, r)
	default:
		writeJSONError(w, http.StatusMethodNotAllowed, errors.New("method not allowed"))
	}
}

func Init() {
	http.HandleFunc("/api/nextdate", nextDateHandler)
	http.HandleFunc("/api/task", taskHandler)
	http.HandleFunc("/api/tasks", tasksHandler)
}
