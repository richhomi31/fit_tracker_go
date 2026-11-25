package main

import (
	"database/sql"
	"errors"
	"fmt"
	"math"
	"net/http"
	"slices"
	"time"

	"github.com/richhomi31/fit_tracker_go.git/internal/app/model"
)

type workoutRequest struct {
	UserID     int               `json:"user_id"`
	StartedAt  time.Time         `json:"started_at"`
	FinishedAt *time.Time        `json:"finished_at"`
	Notes      *string           `json:"notes"`
	Exercises  []workoutExercise `json:"exercises"`
}

type workoutExercise struct {
	ExerciseID int    `json:"exercise_id"`
	Type       string `json:"type"`
	Sets       []struct {
		Number int      `json:"set_number"`
		Weight *float64 `json:"weight_kg"`
		Reps   int      `json:"reps"`
	} `json:"sets"`
	Cardio *struct {
		Duration int               `json:"duration_seconds"`
		Points   []gpsPointRequest `json:"gps_points"`
	} `json:"cardio"`
}

type gpsPointRequest struct {
	Lat        *float64  `json:"lat"`
	Lon        *float64  `json:"lon"`
	RecordedAt time.Time `json:"recorded_at"`
}

func (a api) workoutList(w http.ResponseWriter, r *http.Request, user model.User) {
	if r.Method == http.MethodPost {
		a.saveWorkout(w, r, user.ID, 0)
		return
	}
	workouts, err := a.workouts.List(r.Context(), user.ID)
	if err != nil {
		databaseError(w, err)
		return
	}
	sendJSON(w, http.StatusOK, workouts)
}

func (a api) workout(w http.ResponseWriter, r *http.Request, user model.User) {
	id, ok := routeID(w, r)
	if !ok {
		return
	}
	switch r.Method {
	case http.MethodGet:
		workout, err := a.workouts.Get(r.Context(), user.ID, id)
		if err != nil {
			databaseError(w, err)
			return
		}
		sendJSON(w, http.StatusOK, workout)
	case http.MethodPut:
		a.saveWorkout(w, r, user.ID, id)
	case http.MethodDelete:
		if err := a.workouts.Delete(r.Context(), user.ID, id); err != nil {
			databaseError(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func (a api) saveWorkout(w http.ResponseWriter, r *http.Request, userID, id int) {
	var req workoutRequest
	if !decode(w, r, &req) {
		return
	}
	if req.UserID != 0 && req.UserID != userID {
		fail(w, http.StatusForbidden, "Тренировка должна принадлежать текущему пользователю")
		return
	}
	if id == 0 && req.StartedAt.IsZero() {
		req.StartedAt = time.Now().UTC()
	}
	// Пустой запрос создаёт черновик. Когда клиент присылает упражнения,
	// тренировка завершается, а время окончания можно указать явно.
	if req.FinishedAt == nil && len(req.Exercises) > 0 {
		now := time.Now().UTC()
		req.FinishedAt = &now
	}
	if req.FinishedAt != nil && (req.FinishedAt.IsZero() || len(req.Exercises) == 0) {
		fail(w, http.StatusBadRequest, "Для завершения тренировки нужны упражнения и корректное время окончания")
		return
	}

	var sets []model.WorkoutSet
	for _, input := range req.Exercises {
		if input.ExerciseID <= 0 || input.ExerciseID > 2147483647 {
			fail(w, http.StatusBadRequest, "Неверный ID упражнения")
			return
		}
		// По справочнику определяем, доступно ли упражнение пользователю
		// и какие данные оно принимает: подходы, таймер или таймер с GPS.
		exercise, err := a.exercises.Get(r.Context(), userID, input.ExerciseID)
		if errors.Is(err, sql.ErrNoRows) {
			fail(w, http.StatusBadRequest, "Упражнение отсутствует в справочнике")
			return
		}
		if err != nil {
			databaseError(w, err)
			return
		}
		exerciseSets, err := input.workoutSets(exercise.Type)
		if err != nil {
			fail(w, http.StatusBadRequest, err.Error())
			return
		}
		sets = append(sets, exerciseSets...)
	}

	workout := model.Workout{ID: id, UserID: userID, StartedAt: req.StartedAt, FinishedAt: req.FinishedAt, Notes: req.Notes}
	saved, err := a.workouts.Save(r.Context(), workout, sets)
	if err != nil {
		databaseError(w, err)
		return
	}
	status := http.StatusOK
	if id == 0 {
		status = http.StatusCreated
	}
	sendJSON(w, status, saved)
}

func (input workoutExercise) workoutSets(exerciseType string) ([]model.WorkoutSet, error) {
	if input.Type != "" && input.Type != exerciseType {
		return nil, fmt.Errorf("Тип упражнения %d не совпадает со справочником", input.ExerciseID)
	}
	switch exerciseType {
	case "strength":
		if len(input.Sets) == 0 || input.Cardio != nil {
			return nil, errors.New("Для силового упражнения требуются подходы")
		}
		sets := make([]model.WorkoutSet, 0, len(input.Sets))
		numbers := make(map[int]bool)
		for _, set := range input.Sets {
			if set.Number <= 0 || numbers[set.Number] || set.Weight == nil || *set.Weight < 0 || set.Reps <= 0 {
				return nil, errors.New("Укажите уникальный номер подхода, вес и положительное число повторений")
			}
			numbers[set.Number] = true
			sets = append(sets, model.WorkoutSet{ExerciseID: input.ExerciseID, SetNumber: &set.Number, WeightKg: set.Weight, Reps: &set.Reps})
		}
		return sets, nil
	case "cardio", "cardio_gps":
		if input.Cardio == nil || input.Cardio.Duration <= 0 || len(input.Sets) > 0 {
			return nil, errors.New("Для кардио укажите положительную длительность")
		}
		set := model.WorkoutSet{ExerciseID: input.ExerciseID, DurationSeconds: &input.Cardio.Duration}
		points := input.Cardio.Points
		if exerciseType == "cardio" && len(points) != 0 {
			return nil, errors.New("GPS-точки допустимы только для cardio_gps")
		}
		if exerciseType == "cardio_gps" {
			track := make([]model.GPSPoint, 0, len(points))
			for _, point := range points {
				if point.Lat == nil || point.Lon == nil {
					return nil, errors.New("Укажите широту и долготу каждой GPS-точки")
				}
				track = append(track, model.GPSPoint{Lat: *point.Lat, Lon: *point.Lon, RecordedAt: point.RecordedAt})
			}
			distance, err := trackDistance(track)
			if err != nil {
				return nil, err
			}
			set.DistanceM = &distance
			set.GPSPoints = track
		}
		return []model.WorkoutSet{set}, nil
	default:
		return nil, errors.New("Неизвестный тип упражнения")
	}
}

func trackDistance(points []model.GPSPoint) (float64, error) {
	if len(points) < 2 {
		return 0, errors.New("Для GPS-трека нужны минимум две точки")
	}
	for i := range points {
		point := &points[i]
		if point.RecordedAt.IsZero() || math.Abs(point.Lat) > 90 || math.Abs(point.Lon) > 180 {
			return 0, errors.New("Укажите корректные координаты и время каждой GPS-точки")
		}
		// Расстояние рассчитывается по координатам с точностью хранения в БД.
		point.Lat = math.Round(point.Lat*1e6) / 1e6
		point.Lon = math.Round(point.Lon*1e6) / 1e6
	}
	// Точки могут прийти вразнобой. Складываем расстояния между соседними
	// точками по времени, используя формулу гаверсинусов для каждого отрезка.
	slices.SortStableFunc(points, func(a, b model.GPSPoint) int { return a.RecordedAt.Compare(b.RecordedAt) })
	const radians = math.Pi / 180
	const earthRadiusM = 6371000
	var distance float64
	for i := 1; i < len(points); i++ {
		a, b := points[i-1], points[i]
		dLat, dLon := (b.Lat-a.Lat)*radians, (b.Lon-a.Lon)*radians
		h := math.Pow(math.Sin(dLat/2), 2) + math.Cos(a.Lat*radians)*math.Cos(b.Lat*radians)*math.Pow(math.Sin(dLon/2), 2)
		distance += 2 * earthRadiusM * math.Asin(math.Sqrt(min(1, h)))
	}
	return math.Round(distance*100) / 100, nil
}
