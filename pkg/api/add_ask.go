package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/ezfroze/go_final_project/pkg/db"
)

func addTaskHandler(w http.ResponseWriter, r *http.Request) {
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

	id, err := db.AddTask(&task)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"id": strconv.FormatInt(id, 10)})
}

func checkDate(task *db.Task) error {
	now := time.Now()
	today := now.Format(Dateformat)

	if task.Date == "" {
		task.Date = today
	}

	if _, err := time.Parse(Dateformat, task.Date); err != nil {
		return err
	}

	var (
		next string
		err  error
	)
	if task.Repeat != "" {
		next, err = NextDate(now, task.Date, task.Repeat)
		if err != nil {
			return err
		}
	}

	if task.Date < today {
		if task.Repeat == "" {
			task.Date = today
		} else {
			task.Date = next
		}
	}

	return nil
}
