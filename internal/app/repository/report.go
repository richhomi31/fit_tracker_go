package repository

import (
	"context"
	"errors"
	"time"
)

type ReportRow struct {
	ExerciseID   int        `db:"exercise_id" json:"exercise_id"`
	ExerciseName string     `db:"exercise_name" json:"exercise_name"`
	Bucket       *time.Time `db:"bucket" json:"bucket"`
	Value        float64    `db:"value" json:"value"`
}

func (r *WorkoutRepository) Report(ctx context.Context, userID int, metric, period string, exerciseID int) ([]ReportRow, error) {
	// Метрика выбирает одну из заранее заданных формул. Значения запроса
	// (пользователь, период и упражнение) передаём отдельными параметрами SQL.
	var aggregate, filter string
	switch metric {
	case "volume":
		aggregate, filter = "sum(s.weight_kg * s.reps)", "s.weight_kg IS NOT NULL"
	case "max-weight":
		aggregate, filter = "max(s.weight_kg)", "s.weight_kg IS NOT NULL"
	case "distance":
		aggregate, filter = "sum(s.distance_m)", "s.distance_m IS NOT NULL"
	default:
		return nil, errors.New("unknown report metric")
	}
	// Пустой период превращается в NULL, чтобы получить один итог на
	// упражнение за всю историю. Дневные, недельные и месячные границы идут по UTC.
	rows := []ReportRow{}
	err := r.db.SelectContext(ctx, &rows, `
		SELECT e.id AS exercise_id, e.name AS exercise_name,
		       date_trunc(NULLIF($2, ''), w.started_at, 'UTC') AS bucket, `+aggregate+` AS value
		FROM workouts w
		JOIN workout_sets s ON s.workout_id = w.id
		JOIN exercises e ON e.id = s.exercise_id
		WHERE w.user_id = $1 AND w.finished_at IS NOT NULL AND ($3 = 0 OR e.id = $3) AND `+filter+`
		GROUP BY e.id, e.name, bucket ORDER BY bucket, e.id`, userID, period, exerciseID)
	return rows, err
}
