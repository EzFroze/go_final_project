package api

import (
	"net/http"

	"github.com/ezfroze/go_final_project/pkg/db"
)

type TasksResp struct {
	Tasks []*db.Task `json:"tasks"`
}

const tasksLimit = 50

func tasksHandler(w http.ResponseWriter, r *http.Request) {
	search := r.FormValue("search")

	tasks, err := db.Tasks(tasksLimit, search) // в параметре максимальное количество записей
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err)
		return
	}

	writeJSON(w, http.StatusOK, TasksResp{
		Tasks: tasks,
	})
}
