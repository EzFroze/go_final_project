package db

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"time"
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

func Tasks(limit int, search string) ([]*Task, error) {
	if db == nil {
		return nil, errors.New("database is not initialized")
	}

	date, err := time.Parse("02.01.2006", search)
	isDate := err == nil

	var tasks []*Task

	var query string
	if isDate {
		query = `SELECT id, date, title, comment, repeat
                 FROM scheduler
                 WHERE date = ?
                 ORDER BY date ASC, id ASC
                 LIMIT ?`
	} else {
		query = `SELECT id, date, title, comment, repeat
                 FROM scheduler
                 WHERE title LIKE ? OR comment LIKE ?
                 ORDER BY date ASC, id ASC
                 LIMIT ?`
	}

	stmt, err := db.Prepare(query)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare query: %v", err)
	}
	defer stmt.Close()

	var rows *sql.Rows
	if isDate {
		rows, err = stmt.Query(date.Format("20060102"), limit)
	} else {
		searchParam := "%" + search + "%"
		rows, err = stmt.Query(searchParam, searchParam, limit)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var (
			id int64
			t  Task
		)
		err = rows.Scan(&id, &t.Date, &t.Title, &t.Comment, &t.Repeat)
		if err != nil {
			return nil, err
		}
		t.ID = strconv.FormatInt(id, 10)
		tasks = append(tasks, &t)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	if tasks == nil {
		return []*Task{}, err
	}

	return tasks, nil
}

func GetTask(id string) (*Task, error) {
	var task Task
	if db == nil {
		return nil, errors.New("database is not initialized")
	}

	err := db.QueryRow("SELECT id, date, title, comment, repeat FROM scheduler WHERE id = ?", id).
		Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat)

	if err != nil {
		return nil, err
	}

	return &task, nil
}

func UpdateTask(task *Task) error {
	if db == nil {
		return errors.New("database is not initialized")
	}

	if task.ID == "" {
		return errors.New("task id is empty")
	}

	query :=
		`UPDATE scheduler 
		SET title = :title,
		    comment = :comment,
		    repeat = :repeat,
		    date = :date
		WHERE id = :id`

	res, err := db.Exec(query,
		sql.Named("title", &task.Title),
		sql.Named("comment", &task.Comment),
		sql.Named("repeat", &task.Repeat),
		sql.Named("date", &task.Date),
		sql.Named("id", &task.ID),
	)
	if err != nil {
		return err
	}
	// метод RowsAffected() возвращает количество записей к которым
	// был применена SQL команда
	count, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if count == 0 {
		return fmt.Errorf(`incorrect id for updating task`)
	}
	return nil
}
