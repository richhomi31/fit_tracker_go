package repository

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/richhomi31/fit_tracker_go.git/internal/app/model"
)

type UserRepository struct {
	db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, user *model.User) error {
	return r.db.GetContext(ctx, user, `INSERT INTO users (email, password_hash, name, gender)
		VALUES ($1, $2, $3, $4) RETURNING *`, user.Email, user.PasswordHash, user.Name, user.Gender)
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (model.User, error) {
	var user model.User
	err := r.db.GetContext(ctx, &user, `SELECT * FROM users WHERE email = $1`, email)
	return user, err
}

func (r *UserRepository) Update(ctx context.Context, id int, name string, gender *string) (model.User, error) {
	var user model.User
	err := r.db.GetContext(ctx, &user, `UPDATE users SET name = $2, gender = $3 WHERE id = $1 RETURNING *`, id, name, gender)
	return user, err
}

func (r *UserRepository) Delete(ctx context.Context, id int) error {
	return r.db.GetContext(ctx, &id, `DELETE FROM users WHERE id = $1 RETURNING id`, id)
}
