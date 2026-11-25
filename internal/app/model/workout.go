package model

import "time"

type Workout struct {
	ID         int        `db:"id" json:"id"`
	UserID     int        `db:"user_id" json:"user_id"`
	StartedAt  time.Time  `db:"started_at" json:"started_at"`
	FinishedAt *time.Time `db:"finished_at" json:"finished_at"`
	Notes      *string    `db:"notes" json:"notes"`
}
type WorkoutDetails struct {
	Workout
	Sets []WorkoutSet `json:"sets"`
}
