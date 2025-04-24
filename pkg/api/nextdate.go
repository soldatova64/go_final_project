package api

import (
	"errors"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	maxDays    = 400
	dateFormat = "20060102"
)

var (
	ErrEmptyRepeat     = errors.New("пустое правило повторения")
	ErrInvalidDate     = errors.New("неверный формат даты")
	ErrInvalidRepeat   = errors.New("неверный формат правила повторения")
	ErrUnsupportedRule = errors.New("неподдерживаемый формат правила")
	ErrInvalidDay      = errors.New("недопустимый день")
	ErrInvalidMonth    = errors.New("недопустимый месяц")
	ErrInvalidWeekday  = errors.New("недопустимый день недели")
	ErrMaxDaysExceeded = errors.New("превышено максимальное количество дней")
)

// NextDate - возвращает следующую дату и ошибку.
func NextDate(now time.Time, dstart string, repeat string) (string, error) {
	if repeat == "" {
		return "", ErrEmptyRepeat
	}

	date, err := time.Parse(dateFormat, dstart)
	if err != nil {
		return "", ErrInvalidDate
	}

	parts := strings.Fields(repeat)
	if len(parts) == 0 {
		return "", ErrInvalidRepeat
	}

	now = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

	switch parts[0] {
	case "d":
		return handleDailyRepeat(now, date, parts)
	case "y":
		return handleYearlyRepeat(now, date)
	case "w":
		return handleWeeklyRepeat(now, date, parts)
	case "m":
		return handleMonthlyRepeat(now, date, parts)
	default:
		return "", ErrUnsupportedRule
	}
}

// handleDailyRepeat - обрабатывает правило ежедневного повторения задач (d) и вычисляет следующую дату выполнения
func handleDailyRepeat(now, date time.Time, parts []string) (string, error) {
	if len(parts) != 2 {
		return "", ErrInvalidRepeat
	}

	days, err := strconv.Atoi(parts[1])
	if err != nil {
		return "", ErrInvalidRepeat
	}

	if days == 1 && now.Format(dateFormat) == time.Now().Format(dateFormat) {
		return date.Format(dateFormat), nil
	}

	if days <= 0 || days > maxDays {
		return "", ErrMaxDaysExceeded
	}

	dateNew := date.AddDate(0, 0, days)

	for !afterNow(dateNew, now) {
		dateNew2 := dateNew.AddDate(0, 0, days)
		dateNew = dateNew2
	}

	return dateNew.Format(dateFormat), nil
}

// handleYearlyRepeat - обрабатывает правило ежегодного повторения задач (y) и вычисляет следующую дату выполнения
func handleYearlyRepeat(now, date time.Time) (string, error) {
	newDate := date.AddDate(1, 0, 0)

	if afterNow(newDate, now) {
		return newDate.Format(dateFormat), nil
	}

	return handleYearlyRepeat(now, newDate)
}

// handleWeeklyRepeat - обрабатывает правило еженедельного повторения задач (w) и вычисляет следующую дату выполнения
func handleWeeklyRepeat(now, date time.Time, parts []string) (string, error) {
	if len(parts) != 2 {
		return "", ErrInvalidRepeat
	}

	weekdays, err := parseWeekdays(parts[1])
	if err != nil {
		return "", err
	}
	date = date.AddDate(0, 0, 1)

	current := date

	for {
		if afterNow(current, now) {
			for _, wd := range weekdays {
				if int(current.Weekday()) == wd {
					return current.Format(dateFormat), nil
				}
			}
		}
		current = current.AddDate(0, 0, 1)
	}
}

// handleMonthlyRepeat - обрабатывает правило ежемесячного повторения задач (m) и вычисляет следующую дату выполнения
func handleMonthlyRepeat(now, date time.Time, parts []string) (string, error) {
	if len(parts) < 2 || len(parts) > 3 {
		return "", ErrUnsupportedRule
	}

	days, months, err := parseMonthly(parts[1:])
	if err != nil {
		return "", err
	}

	datesCandidates := []time.Time{}
	var dateInitial time.Time
	var dateCandidate time.Time
	var valueMonth int
	var valueDay int

	dateInitial = date
	daysInMonth := time.Date(dateInitial.Year(), dateInitial.Month()+1, 0, 0, 0, 0, 0, time.UTC).Day()

	if dateInitial.Unix() < now.Unix() {
		dateInitial = now
	}

	if len(months) > 0 && len(days) > 0 {
		for valueMonth = range months {
			for valueDay = range days {
				dateCandidate = time.Date(dateInitial.Year(), time.Month(valueMonth), valueDay, 0, 0, 0, 0, time.UTC)

				if dateCandidate.Unix() <= dateInitial.Unix() {
					dateCandidate = dateCandidate.AddDate(1, 0, 0)
				}

				datesCandidates = append(datesCandidates, dateCandidate)
			}
		}

		sort.Slice(datesCandidates, func(i, j int) bool {
			return datesCandidates[i].Before(datesCandidates[j])
		})

		return datesCandidates[0].Format(dateFormat), nil
	}

	if len(days) > 0 {
		for valueDay = range days {
			if daysInMonth < valueDay {
				dateInitial = time.Date(dateInitial.Year(), dateInitial.Month()+1, 1, 0, 0, 0, 0, time.UTC)
			}

			if valueDay > 0 {
				dateCandidate = time.Date(dateInitial.Year(), dateInitial.Month(), valueDay, 0, 0, 0, 0, time.UTC)
			} else {
				dateCandidate = time.Date(dateInitial.Year(), dateInitial.Month(), daysInMonth+(valueDay+1), 0, 0, 0, 0, time.UTC)
			}

			if dateCandidate.Unix() <= dateInitial.Unix() {
				dateCandidate = dateCandidate.AddDate(0, 1, 0)
			}

			datesCandidates = append(datesCandidates, dateCandidate)
		}

		sort.Slice(datesCandidates, func(i, j int) bool {
			return datesCandidates[i].Before(datesCandidates[j])
		})

		return datesCandidates[0].Format(dateFormat), nil
	}

	return "", ErrMaxDaysExceeded
}

// afterNow - сравнивает две даты без учёта времени и проверяет, находится ли первая дата строго после второй
func afterNow(date, now time.Time) bool {
	date = time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	now = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	return date.After(now)
}

// parseWeekdays - преобразует строку с днями недели в массив чисел и валидирует их
func parseWeekdays(weekdaysStr string) ([]int, error) {
	parts := strings.Split(weekdaysStr, ",")
	days := make([]int, 0, len(parts))
	for _, part := range parts {
		wd, err := strconv.Atoi(strings.TrimSpace(part))
		if err != nil || wd < 1 || wd > 7 {
			return nil, ErrInvalidWeekday
		}
		days = append(days, wd%7)
	}
	return days, nil
}

// parseMonthly - парсит и валидирует правила для ежемесячного повторения задач
func parseMonthly(parts []string) (map[int]bool, map[int]bool, error) {
	days := make(map[int]bool)
	months := make(map[int]bool)

	for _, dayStr := range strings.Split(parts[0], ",") {
		day, err := strconv.Atoi(strings.TrimSpace(dayStr))
		if err != nil {
			return nil, nil, ErrInvalidDay
		}
		if day == -1 || day == -2 {
			days[day] = true
		} else if day < 1 || day > 31 {
			return nil, nil, ErrInvalidDay
		} else {
			days[day] = true
		}
	}

	if len(parts) > 1 {
		for _, monthStr := range strings.Split(parts[1], ",") {
			month, err := strconv.Atoi(strings.TrimSpace(monthStr))
			if err != nil || month < 1 || month > 12 {
				return nil, nil, ErrInvalidMonth
			}
			months[month] = true
		}
	}

	return days, months, nil
}

// HandleNexDate - обработчик HTTP запросов
func HandleNexDate(w http.ResponseWriter, r *http.Request) {

	nowStr := r.FormValue("now")
	date := r.FormValue("date")
	repeat := r.FormValue("repeat")

	var now time.Time
	var err error

	if nowStr == "" {
		now = time.Now().UTC()
	} else {
		now, err = time.Parse(dateFormat, nowStr)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	if date == "" {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if repeat == "" {
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte(""))
		return
	}

	// Вычисляем следующую дату
	result, err := NextDate(now, date, repeat)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Возвращаем результат
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(result))
}
