package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/ezfroze/go_final_project/pkg/db"
)

func updateTask(w http.ResponseWriter, r *http.Request) {
	if r.Body == nil {
		writeJSONError(w, http.StatusBadRequest, errors.New("empty request body"))
		return
	}

	var task db.Task
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		writeJSONError(w, http.StatusBadRequest, err)
		return
	}

	task.Title = strings.TrimSpace(task.Title)
	if task.Title == "" {
		writeJSONError(w, http.StatusBadRequest, errors.New("title is required"))
		return
	}

	task.Date = strings.TrimSpace(task.Date)
	task.Repeat = strings.TrimSpace(task.Repeat)

	if err := checkDate(&task); err != nil {
		writeJSONError(w, http.StatusBadRequest, err)
		return
	}

	err := db.UpdateTask(&task)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"id": task.ID})
}
