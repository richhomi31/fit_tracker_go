package repository

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/richhomi31/fit_tracker_go.git/internal/app/model"
)

type WorkoutRepository struct {
	db *sqlx.DB
}

func NewWorkoutRepository(db *sqlx.DB) *WorkoutRepository {
	return &WorkoutRepository{db: db}
}

func (r *WorkoutRepository) List(ctx context.Context, userID int) ([]model.Workout, error) {
	workouts := []model.Workout{}
	err := r.db.SelectContext(ctx, &workouts, `SELECT * FROM workouts WHERE user_id = $1 ORDER BY started_at DESC, id DESC`, userID)
	return workouts, err
}

func (r *WorkoutRepository) Get(ctx context.Context, userID, id int) (model.WorkoutDetails, error) {
	result := model.WorkoutDetails{Sets: []model.WorkoutSet{}}
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return result, err
	}
	defer tx.Rollback()
	// FOR SHARE удерживает запись тренировки до конца чтения. Иначе
	// параллельный PUT мог бы заменить подходы между запросами к разным таблицам.
	if err := tx.GetContext(ctx, &result.Workout, `SELECT * FROM workouts WHERE id = $1 AND user_id = $2 FOR SHARE`, id, userID); err != nil {
		return result, err
	}
	if err := tx.SelectContext(ctx, &result.Sets, `SELECT * FROM workout_sets WHERE workout_id = $1 ORDER BY id`, id); err != nil {
		return result, err
	}
	var points []model.GPSPoint
	if err := tx.SelectContext(ctx, &points, `SELECT p.* FROM gps_points p
		JOIN workout_sets s ON s.id = p.set_id WHERE s.workout_id = $1 ORDER BY p.recorded_at, p.id`, id); err != nil {
		return result, err
	}
	bySet := make(map[int][]model.GPSPoint)
	for _, point := range points {
		bySet[point.SetID] = append(bySet[point.SetID], point)
	}
	for i := range result.Sets {
		result.Sets[i].GPSPoints = bySet[result.Sets[i].ID]
	}
	return result, tx.Commit()
}

func (r *WorkoutRepository) Save(ctx context.Context, workout model.Workout, sets []model.WorkoutSet) (model.Workout, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return workout, err
	}
	// При ошибке Rollback отменит все записи этой операции. После Commit
	// отложенный Rollback уже ничего не меняет.
	defer tx.Rollback()

	if workout.ID == 0 {
		err = tx.GetContext(ctx, &workout, `INSERT INTO workouts (user_id, started_at, finished_at, notes)
			VALUES ($1, $2, $3, $4) RETURNING *`, workout.UserID, workout.StartedAt, workout.FinishedAt, workout.Notes)
	} else {
		err = tx.GetContext(ctx, &workout, `UPDATE workouts
			SET started_at = COALESCE($3, started_at), finished_at = $4, notes = $5
			WHERE id = $1 AND user_id = $2 RETURNING *`, workout.ID, workout.UserID, optionalStart(workout), workout.FinishedAt, workout.Notes)
		// PUT присылает тренировку целиком. Удаляем прежние подходы,
		// а PostgreSQL каскадно удаляет их GPS-точки.
		if err == nil {
			_, err = tx.ExecContext(ctx, `DELETE FROM workout_sets WHERE workout_id = $1`, workout.ID)
		}
	}
	if err != nil {
		return workout, err
	}
	for _, set := range sets {
		// RETURNING id даёт ключ новой записи, по которому привязываем GPS-точки.
		var setID int
		err := tx.GetContext(ctx, &setID, `INSERT INTO workout_sets
			(workout_id, exercise_id, set_number, weight_kg, reps, duration_seconds, distance_m)
			VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`,
			workout.ID, set.ExerciseID, set.SetNumber, set.WeightKg, set.Reps, set.DurationSeconds, set.DistanceM)
		if err != nil {
			return workout, err
		}
		for _, point := range set.GPSPoints {
			_, err := tx.ExecContext(ctx, `INSERT INTO gps_points (set_id, lat, lon, recorded_at)
				VALUES ($1, $2, $3, $4)`, setID, point.Lat, point.Lon, point.RecordedAt)
			if err != nil {
				return workout, err
			}
		}
	}
	return workout, tx.Commit()
}

func optionalStart(workout model.Workout) any {
	// Без started_at SQL получает NULL, и COALESCE оставляет прежнюю дату.
	if workout.StartedAt.IsZero() {
		return nil
	}
	return workout.StartedAt
}

func (r *WorkoutRepository) Delete(ctx context.Context, userID, id int) error {
	return r.db.GetContext(ctx, &id, `DELETE FROM workouts WHERE id = $1 AND user_id = $2 RETURNING id`, id, userID)
}
