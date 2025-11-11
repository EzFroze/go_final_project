package db

import (
	"errors"
	"strconv"
)

type Task struct {
	ID      string `json:"id"`
	Date    string `json:"date"`
	Title   string `json:"title"`
	Comment string `json:"comment"`
	Repeat  string `json:"repeat"`
}

func AddTask(task *Task) (int64, error) {
	if db == nil {
		return 0, errors.New("database is not initialized")
	}

	res, err := db.Exec(`INSERT INTO scheduler (date, title, comment, repeat) VALUES (?, ?, ?, ?)`,
		task.Date, task.Title, task.Comment, task.Repeat)
	if err != nil {
		return 0, err
	}

	return res.LastInsertId()
}

func Tasks(limit int) ([]*Task, error) {
	if db == nil {
		return nil, errors.New("database is not initialized")
	}

	var tasks []*Task

	res, err := db.Query(`
				SELECT id, date, title, comment, repeat 
				FROM scheduler 
				ORDER BY date ASC, id ASC 
				LIMIT ?
			`, limit)

	if err != nil {
		return nil, err
	}
	defer res.Close()

	for res.Next() {
		var (
			id int64
			t  Task
		)
		err = res.Scan(&id, &t.Date, &t.Title, &t.Comment, &t.Repeat)
		if err != nil {
			return nil, err
		}
		t.ID = strconv.FormatInt(id, 10)
		tasks = append(tasks, &t)
	}

	if err = res.Err(); err != nil {
		return nil, err
	}

	if tasks == nil {
		return []*Task{}, err
	}

	return tasks, nil
}
