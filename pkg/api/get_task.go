package api

import (
	"errors"
	"net/http"

	"github.com/ezfroze/go_final_project/pkg/db"
)

func getTask(w http.ResponseWriter, r *http.Request) {
	id := r.FormValue("id")

	if id == "" {
		writeJSONError(w, http.StatusBadRequest, errors.New("id is required"))
		return
	}

	task, err := db.GetTask(id)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err)
		return
	}

	writeJSON(w, http.StatusOK, task)
}
