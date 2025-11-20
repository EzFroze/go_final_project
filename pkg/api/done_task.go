package api

import (
	"errors"
	"net/http"
	"time"

	"github.com/ezfroze/go_final_project/pkg/db"
)

func doneTaskHandler(w http.ResponseWriter, r *http.Request) {
	id := r.FormValue("id")

	if id == "" {
		writeJSONError(w, http.StatusBadRequest, errors.New("id is required"))
	}

	task, err := db.GetTask(id)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err)
		return
	}

	if task.Repeat == "" {
		err := db.DeleteTask(id)
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, err)
			return
		}

		writeJSON(w, http.StatusOK, map[string]interface{}{})
		return
	}

	nextDate, err := NextDate(time.Now(), task.Date, task.Repeat)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err)
		return
	}

	err = db.UpdateTaskDate(id, nextDate)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{})
}
