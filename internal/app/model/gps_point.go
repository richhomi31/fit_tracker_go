package model

import "time"

type GPSPoint struct {
	ID         int       `db:"id" json:"id"`
	SetID      int       `db:"set_id" json:"set_id"`
	Lat        float64   `db:"lat" json:"lat"`
	Lon        float64   `db:"lon" json:"lon"`
	RecordedAt time.Time `db:"recorded_at" json:"recorded_at"`
}
