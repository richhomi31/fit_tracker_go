package main

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"net/mail"
	"strings"

	"github.com/richhomi31/fit_tracker_go.git/internal/app/model"
	"golang.org/x/crypto/bcrypt"
)

type userRequest struct {
	Email    string  `json:"email"`
	Password string  `json:"password"`
	Name     string  `json:"name"`
	Gender   *string `json:"gender"`
}

func (a api) register(w http.ResponseWriter, r *http.Request) {
	var req userRequest
	if !decode(w, r, &req) {
		return
	}
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
	address, err := mail.ParseAddress(req.Email)
	if err != nil || address.Address != req.Email || strings.TrimSpace(req.Name) == "" || req.Password == "" {
		fail(w, http.StatusBadRequest, "Укажите почту, пароль и имя")
		return
	}
	// В БД сохраняем только bcrypt-хеш. Сам пароль нужен на время запроса.
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		fail(w, http.StatusBadRequest, "Пароль должен содержать не более 72 байт")
		return
	}
	user := model.User{Email: req.Email, PasswordHash: string(hash), Name: req.Name, Gender: req.Gender}
	if err := a.users.Create(r.Context(), &user); err != nil {
		databaseError(w, err)
		return
	}
	sendJSON(w, http.StatusCreated, user)
}

var errCredentials = errors.New("invalid credentials")

func (a api) authenticate(ctx context.Context, email, password string) (model.User, error) {
	user, err := a.users.GetByEmail(ctx, strings.ToLower(strings.TrimSpace(email)))
	if errors.Is(err, sql.ErrNoRows) {
		return user, errCredentials
	}
	if err != nil {
		return user, err
	}
	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)) != nil {
		return user, errCredentials
	}
	return user, nil
}

func (a api) login(w http.ResponseWriter, r *http.Request) {
	var req userRequest
	if !decode(w, r, &req) {
		return
	}
	user, err := a.authenticate(r.Context(), req.Email, req.Password)
	if errors.Is(err, errCredentials) {
		fail(w, http.StatusUnauthorized, "Неверная почта или пароль")
		return
	}
	if err != nil {
		databaseError(w, err)
		return
	}
	sendJSON(w, http.StatusOK, map[string]any{"user_id": user.ID, "name": user.Name})
}

// HTTP Basic передаёт почту и пароль с каждым запросом. После проверки
// обработчик получает пользователя из БД и может ограничить доступ его записями.
func (a api) auth(next func(http.ResponseWriter, *http.Request, model.User)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		email, password, ok := r.BasicAuth()
		if !ok {
			w.Header().Set("WWW-Authenticate", `Basic realm="Fitness Tracker"`)
			fail(w, http.StatusUnauthorized, "Требуется авторизация")
			return
		}
		user, err := a.authenticate(r.Context(), email, password)
		if errors.Is(err, errCredentials) {
			w.Header().Set("WWW-Authenticate", `Basic realm="Fitness Tracker"`)
			fail(w, http.StatusUnauthorized, "Неверная почта или пароль")
			return
		}
		if err != nil {
			databaseError(w, err)
			return
		}
		next(w, r, user)
	}
}

func (a api) user(w http.ResponseWriter, r *http.Request, user model.User) {
	id, ok := routeID(w, r)
	if !ok {
		return
	}
	if id != user.ID {
		fail(w, http.StatusNotFound, "Запись не найдена")
		return
	}

	switch r.Method {
	case http.MethodGet:
		sendJSON(w, http.StatusOK, user)
	case http.MethodPut:
		var req userRequest
		if !decode(w, r, &req) {
			return
		}
		if strings.TrimSpace(req.Name) == "" {
			fail(w, http.StatusBadRequest, "Укажите имя")
			return
		}
		updated, err := a.users.Update(r.Context(), id, req.Name, req.Gender)
		if err != nil {
			databaseError(w, err)
			return
		}
		sendJSON(w, http.StatusOK, updated)
	case http.MethodDelete:
		if err := a.users.Delete(r.Context(), id); err != nil {
			databaseError(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}
