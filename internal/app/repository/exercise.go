package repository

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/richhomi31/fit_tracker_go.git/internal/app/model"
)

type ExerciseRepository struct {
	db *sqlx.DB
}

func NewExerciseRepository(db *sqlx.DB) *ExerciseRepository {
	return &ExerciseRepository{db: db}
}

func (r *ExerciseRepository) List(ctx context.Context, userID int) ([]model.Exercise, error) {
	exercises := []model.Exercise{}
	err := r.db.SelectContext(ctx, &exercises, `SELECT * FROM exercises
		WHERE NOT is_custom OR created_by = $1 ORDER BY id`, userID)
	return exercises, err
}

func (r *ExerciseRepository) Create(ctx context.Context, userID int, exercise model.Exercise) (model.Exercise, error) {
	err := r.db.GetContext(ctx, &exercise, `INSERT INTO exercises (name, muscle_group, type, is_custom, created_by)
		VALUES ($1, $2, $3, TRUE, $4) RETURNING *`, exercise.Name, exercise.MuscleGroup, exercise.Type, userID)
	return exercise, err
}

func (r *ExerciseRepository) Update(ctx context.Context, userID, id int, exercise model.Exercise) (model.Exercise, error) {
	// Тип упражнения задаётся при создании, потому что от него зависит
	// смысл уже сохранённых данных: вес и повторения или время и маршрут.
	err := r.db.GetContext(ctx, &exercise, `UPDATE exercises SET name = $3, muscle_group = $4
		WHERE id = $1 AND created_by = $2 RETURNING *`, id, userID, exercise.Name, exercise.MuscleGroup)
	return exercise, err
}

func (r *ExerciseRepository) Delete(ctx context.Context, userID, id int) error {
	return r.db.GetContext(ctx, &id, `DELETE FROM exercises WHERE id = $1 AND created_by = $2 RETURNING id`, id, userID)
}

func (r *ExerciseRepository) Get(ctx context.Context, userID, id int) (model.Exercise, error) {
	var exercise model.Exercise
	err := r.db.GetContext(ctx, &exercise, `SELECT * FROM exercises
		WHERE id = $2 AND (NOT is_custom OR created_by = $1)`, userID, id)
	return exercise, err
}
