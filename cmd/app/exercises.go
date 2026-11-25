package main

import (
	"net/http"
	"strings"

	"github.com/richhomi31/fit_tracker_go.git/internal/app/model"
)

func (a api) exerciseList(w http.ResponseWriter, r *http.Request, user model.User) {
	if r.Method == http.MethodGet {
		exercises, err := a.exercises.List(r.Context(), user.ID)
		if err != nil {
			databaseError(w, err)
			return
		}
		sendJSON(w, http.StatusOK, exercises)
		return
	}

	var req model.Exercise
	if !decode(w, r, &req) {
		return
	}
	if strings.TrimSpace(req.Name) == "" {
		fail(w, http.StatusBadRequest, "Укажите название упражнения")
		return
	}
	exercise, err := a.exercises.Create(r.Context(), user.ID, req)
	if err != nil {
		databaseError(w, err)
		return
	}
	sendJSON(w, http.StatusCreated, exercise)
}

func (a api) exercise(w http.ResponseWriter, r *http.Request, user model.User) {
	id, ok := routeID(w, r)
	if !ok {
		return
	}
	switch r.Method {
	case http.MethodGet:
		exercise, err := a.exercises.Get(r.Context(), user.ID, id)
		if err != nil {
			databaseError(w, err)
			return
		}
		sendJSON(w, http.StatusOK, exercise)
	case http.MethodPut:
		var req model.Exercise
		if !decode(w, r, &req) {
			return
		}
		if strings.TrimSpace(req.Name) == "" {
			fail(w, http.StatusBadRequest, "Укажите название упражнения")
			return
		}
		exercise, err := a.exercises.Update(r.Context(), user.ID, id, req)
		if err != nil {
			databaseError(w, err)
			return
		}
		sendJSON(w, http.StatusOK, exercise)
	case http.MethodDelete:
		if err := a.exercises.Delete(r.Context(), user.ID, id); err != nil {
			databaseError(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}
