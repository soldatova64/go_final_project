package api

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
	"os"
	"time"
)

// HTTP-обработчик для аутентификации пользователей
func SigninHandler(w http.ResponseWriter, r *http.Request) {
	var creds struct {
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		writeJson(w, map[string]string{"error": "Недопустимый формат запроса"}, http.StatusBadRequest)
		return
	}

	envPassword := os.Getenv("TODO_PASSWORD")
	if envPassword == "" {
		writeJson(w, map[string]string{"error": "Аутентификация не настроена"}, http.StatusInternalServerError)
		return
	}

	if creds.Password != envPassword {
		writeJson(w, map[string]string{"error": "Неверный пароль"}, http.StatusUnauthorized)
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"pass_hash": hashPassword(envPassword),
		"exp":       time.Now().Add(8 * time.Hour).Unix(),
	})

	tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		writeJson(w, map[string]string{"error": err.Error()}, http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    tokenString,
		Expires:  time.Now().Add(8 * time.Hour),
		HttpOnly: true,
		Path:     "/",
	})

	writeJson(w, map[string]string{"token": tokenString}, http.StatusOK)
}

func hashPassword(password string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(password)))
}
