package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"

	"github.com/golang-jwt/jwt/v5"
)

type RequestBody struct {
	Password string `json:"password"`
}

const hmacSampleSecret = "my_secret_key"

func signInHandler(w http.ResponseWriter, r *http.Request) {
	if r.Body == nil {
		writeJSONError(w, http.StatusBadRequest, errors.New("empty request body"))
		return
	}

	var requestBody RequestBody
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		writeJSONError(w, http.StatusBadRequest, err)
		return
	}

	if requestBody.Password == "" {
		writeJSONError(w, http.StatusBadRequest, errors.New("empty password"))
		return
	}

	password := os.Getenv("TODO_PASSWORD")

	if password == "" {
		writeJSONError(w, http.StatusBadRequest, errors.New("not set TODO_PASSWORD"))
		return
	}

	if password != requestBody.Password {
		writeJSONError(w, http.StatusUnauthorized, errors.New("wrong password"))
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"password": requestBody.Password,
	})

	tokenString, err := token.SignedString([]byte(hmacSampleSecret))

	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"token": tokenString,
	})
}

func auth(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// смотрим наличие пароля
		pass := os.Getenv("TODO_PASSWORD")
		if len(pass) > 0 {
			var jwtToken string // JWT-токен из куки
			// получаем куку
			cookie, err := r.Cookie("token")
			if err == nil {
				jwtToken = cookie.Value
			}
			// здесь код для валидации и проверки JWT-токена
			token, err := jwt.Parse(jwtToken, func(token *jwt.Token) (any, error) {
				return []byte(hmacSampleSecret), nil
			}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))

			if err != nil {
				writeJSONError(w, http.StatusUnauthorized, err)
				return
			}

			if !token.Valid {
				// возвращаем ошибку авторизации 401
				writeJSONError(w, http.StatusUnauthorized, errors.New("authentification required"))
				return
			}
		}
		next(w, r)
	})
}
