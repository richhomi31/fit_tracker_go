package model

type WorkoutSet struct {
	ID              int        `db:"id" json:"id"`
	WorkoutID       int        `db:"workout_id" json:"workout_id"`
	ExerciseID      int        `db:"exercise_id" json:"exercise_id"`
	SetNumber       *int       `db:"set_number" json:"set_number"`             // NULL для кардио
	WeightKg        *float64   `db:"weight_kg" json:"weight_kg"`               // NULL для кардио
	Reps            *int       `db:"reps" json:"reps"`                         // NULL для кардио
	DurationSeconds *int       `db:"duration_seconds" json:"duration_seconds"` // NULL для силовых
	DistanceM       *float64   `db:"distance_m" json:"distance_m"`             // NULL, если не GPS
	GPSPoints       []GPSPoint `db:"-" json:"gps_points,omitempty"`
}
