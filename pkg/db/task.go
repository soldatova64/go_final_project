package db

import (
	"errors"
	"fmt"
	"time"
)

type Task struct {
	ID      string `json:"id"`
	Date    string `json:"date"`
	Title   string `json:"title"`
	Comment string `json:"comment,omitempty"`
	Repeat  string `json:"repeat,omitempty"`
}

// AddTask добавляет задачу в таблицу scheduler и возвращает ID новой записи
func AddTask(task *Task) (int64, error) {
	var id int64

	date, err := time.Parse("20060102", task.Date)
	if err != nil {
		return 0, err
	}

	dbDate := date.Format("20060102")

	query := "insert into scheduler (date, title, comment, repeat) values (?, ?, ?, ?)"
	args := []interface{}{dbDate, task.Title, task.Comment, task.Repeat}

	res, err := db.Exec(query, args...)
	if err != nil {
		return 0, err
	}

	id, err = res.LastInsertId()
	if err != nil {
		return 0, err
	}

	return id, nil
}

// Tasks возвращает список задач с сортировкой по дате
func Tasks(limit int) ([]*Task, error) {
	var tasks []*Task
	query := "select * from scheduler order by date ASC lIMIT ?"

	rows, err := db.Query(query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var task Task
		if err := rows.Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat); err != nil {
			return nil, err
		}
		tasks = append(tasks, &task)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if tasks == nil {
		tasks = make([]*Task, 0)
	}

	return tasks, nil
}

// GetTask возвращает задачу по ID
func GetTask(id string) (*Task, error) {
	if id == "" {
		return nil, errors.New("no tasks found")
	}
	query := "select * from scheduler where id = ?"

	var task Task

	err := db.QueryRow(query, id).Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat)
	if err != nil {
		return nil, err
	}
	return &task, nil
}

// UpdateTask обновляет существующую задачу
func UpdateTask(task *Task) error {
	query := "update scheduler set date=?, title=?, comment=?, repeat=? where id=?"
	res, err := db.Exec(query, task.Date, task.Title, task.Comment, task.Repeat, task.ID)
	if err != nil {
		return err
	}
	count, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if count == 0 {
		return fmt.Errorf(`incorrect id for updating task`)
	}
	return nil
}

func UpdateDate(id string, next string) error {
	if next == "" || id == "" {
		return errors.New("ID и дата не могут быть пустыми")
	}
	query := "update scheduler set date=? where id=?"
	res, err := db.Exec(query, next, id)
	if err != nil {
		return err
	}
	count, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if count == 0 {
		return fmt.Errorf(`incorrect id for updating task`)
	}
	return nil
}

// DeleteTask удаляет существующую задачу
func DeleteTask(id string) error {
	if id == "" {
		return errors.New("no tasks found")
	}
	query := "delete from scheduler where id = ?"
	res, err := db.Exec(query, id)
	if err != nil {
		return err
	}
	count, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if count == 0 {
		return fmt.Errorf(`incorrect id for deleting task`)
	}
	return nil
}
