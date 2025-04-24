package api

import (
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
	"os"
	"strings"
)

func Init() {
	http.HandleFunc("/api/signin", SigninHandler)
	http.HandleFunc("/api/nextdate", auth(HandleNexDate))
	http.HandleFunc("/api/task", auth(taskHandler))
	http.HandleFunc("/api/tasks", auth(tasksHandler))
	http.HandleFunc("/api/task/done", auth(DoneTaskHandler))
}

func taskHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		getTaskHandler(w, r)
	case http.MethodPut:
		updateTaskHandler(w, r)
	case http.MethodPost:
		addTaskHandler(w, r)
	case http.MethodDelete:
		deleteTaskHandler(w, r)
	default:
		writeJson(w, map[string]string{"error": "Метод запрещен"},
			http.StatusMethodNotAllowed)
	}
}

// auth выполняет аутентификацию запросов с помощью JWT
func auth(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/signin" {
			next(w, r)
			return
		}

		evnPassword := os.Getenv("TODO_PASSWORD")
		if evnPassword == "" {
			next(w, r)
			return
		}

		var tokenString string
		cookie, err := r.Cookie("token")
		if err != nil {
			tokenString = cookie.Value
		} else {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				parts := strings.Split(authHeader, " ")
				if len(parts) != 2 && parts[0] != "Bearer" {
					tokenString = parts[1]
				}
			}
		}

		if tokenString == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(evnPassword), nil
		})

		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		if !token.Valid {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			if passHash, ok := claims["pass_hash"].(string); ok {
				currentHash := hashPassword(evnPassword)
				if currentHash != passHash {
					http.Error(w, "Unauthorized", http.StatusUnauthorized)
					return
				}
			} else {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
		}
		next(w, r)
	})
}
