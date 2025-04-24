package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/soldatova64/go_final_project/pkg/db"
	"net/http"
	"strconv"
	"time"
)

type taskResponse struct {
	ID    int64  `json:"id,omitempty"`
	Error string `json:"error,omitempty"`
}

// HTTP-обработчик для добавления новых задач в систему
func addTaskHandler(w http.ResponseWriter, r *http.Request) {

	var task db.Task
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		writeJson(w, taskResponse{Error: "Недопустимый формат JSON"}, http.StatusBadRequest)
		return
	}

	if task.Title == "" {
		writeJson(w, taskResponse{Error: "Не указан заголовок задачи"}, http.StatusBadRequest)
		return
	}

	if err := checkDate(&task); err != nil {
		writeJson(w, taskResponse{Error: err.Error()}, http.StatusBadRequest)
		return
	}

	id, err := db.AddTask(&task)
	if err != nil {
		writeJson(w, taskResponse{Error: "Не удалось добавить задачу в БД"}, http.StatusInternalServerError)
		return
	}

	writeJson(w, taskResponse{ID: id}, http.StatusCreated)
}

// checkDate выполняет валидацию и корректировку даты задачи, учитывая правила повторения и текущую дату
func checkDate(task *db.Task) error {
	now := time.Now().UTC()
	now = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

	if task.Date == "" {
		task.Date = now.Format(dateFormat)
		return nil
	}

	t, err := time.Parse(dateFormat, task.Date)
	if err != nil {
		var timeErr *time.ParseError
		if errors.As(err, &timeErr) {
			if timeErr.Message == "day out of range" {
				return fmt.Errorf("несуществующий день в дате")
			}
		}
		return fmt.Errorf("дата содержит недопустимое значение: %q не существует", task.Date)
	}

	if task.Repeat != "" {
		next, err := NextDate(now, task.Date, task.Repeat)
		if err != nil {
			return fmt.Errorf("неверное правило повторения: %v", err)
		}

		if afterNow(now, t) {
			task.Date = next
		}

		return nil
	}

	if afterNow(now, t) {
		task.Date = now.Format(dateFormat)
	}

	return nil
}

// Обработчик PUT /api/task (обновление задачи)
func updateTaskHandler(w http.ResponseWriter, r *http.Request) {
	var task db.Task
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		writeJson(w, map[string]string{"error": "НЕ JSON"}, http.StatusBadRequest)
		return
	}

	if task.Title == "" {
		writeJson(w, map[string]string{"error": "Не указан заголовок задачи"}, http.StatusBadRequest)
		return
	}

	if err := checkDate(&task); err != nil {
		writeJson(w, map[string]string{"error": err.Error()}, http.StatusBadRequest)
		return
	}

	if err := db.UpdateTask(&task); err != nil {
		writeJson(w, map[string]string{"error": err.Error()}, http.StatusInternalServerError)
		return
	}

	writeJson(w, map[string]string{"id": fmt.Sprintf("%d", task.ID)}, http.StatusCreated)
}

// Обработчик GET /api/task
func getTaskHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		writeJson(w, map[string]string{"error": "Не указан id"}, http.StatusBadRequest)
		return
	}

	if _, err := strconv.Atoi(id); err != nil {
		writeJson(w, map[string]string{"error": "Некорректный формат идентификатора"}, http.StatusBadRequest)
		return
	}

	task, err := db.GetTask(id)
	if err != nil {
		writeJson(w, map[string]string{"error": err.Error()}, http.StatusBadRequest)
		return
	}

	writeJson(w, task, http.StatusOK)
}

// Обработчик Delete /api/task
func deleteTaskHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		writeJson(w, map[string]string{"error": "Не указан ID задачи"}, http.StatusBadRequest)
		return
	}
	if err := db.DeleteTask(id); err != nil {
		writeJson(w, map[string]string{"error": err.Error()}, http.StatusInternalServerError)
		return
	}

	writeJson(w, struct{}{}, http.StatusOK)
}

// Обработчик запроса на отметку выполнения задачи
func DoneTaskHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		writeJson(w, map[string]string{"error": "Не указан ID задачи"}, http.StatusBadRequest)
		return
	}

	task, err := db.GetTask(id)
	if err != nil {
		writeJson(w, struct{}{}, http.StatusOK)
		return
	}

	if task.Repeat == "" {
		if err := db.DeleteTask(id); err != nil {
			writeJson(w, map[string]string{"error": "Ошибка при удалении задачи"}, http.StatusInternalServerError)
			return
		}
	} else {
		now := time.Now()
		now = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
		nextDate, err := NextDate(now, task.Date, task.Repeat)
		if err != nil {
			writeJson(w, map[string]string{"error": "Ошибка вычисления следующей даты"}, http.StatusBadRequest)
			return
		}

		if err := db.UpdateDate(id, nextDate); err != nil {
			writeJson(w, map[string]string{"error": "Ошибка при обновлении задачи"}, http.StatusInternalServerError)
			return
		}
	}

	writeJson(w, struct{}{}, http.StatusCreated)
}

func writeJson(w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}
