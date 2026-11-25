package model

import "time"

type User struct {
	ID           int       `db:"id" json:"id"`
	Email        string    `db:"email" json:"email"`
	PasswordHash string    `db:"password_hash" json:"-"`
	Name         string    `db:"name" json:"name"`
	Gender       *string   `db:"gender" json:"gender"`
	RegisteredAt time.Time `db:"registered_at" json:"registered_at"`
}
