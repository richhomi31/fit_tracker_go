package main

import (
	"net/http"
	"strconv"

	"github.com/richhomi31/fit_tracker_go.git/internal/app/model"
)

func (a api) report(metric string) func(http.ResponseWriter, *http.Request, model.User) {
	return func(w http.ResponseWriter, r *http.Request, user model.User) {
		query := r.URL.Query()
		period := query.Get("period")
		switch period {
		case "", "day", "week", "month":
		default:
			fail(w, http.StatusBadRequest, "Допустимые периоды: day, week, month")
			return
		}
		var id int
		if value := query.Get("exercise_id"); value != "" {
			parsed, err := strconv.ParseInt(value, 10, 32)
			if err != nil || parsed <= 0 {
				fail(w, http.StatusBadRequest, "Неверный ID упражнения")
				return
			}
			id = int(parsed)
		}
		rows, err := a.workouts.Report(r.Context(), user.ID, metric, period, id)
		if err != nil {
			databaseError(w, err)
			return
		}
		sendJSON(w, http.StatusOK, map[string]any{"report_type": metric, "period": period, "data": rows})
	}
}
