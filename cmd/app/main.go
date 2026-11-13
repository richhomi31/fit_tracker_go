package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

// --- СТРУКТУРЫ ДЛЯ JSON ---

type StrengthSet struct {
	SetNumber int     `json:"set_number"`
	WeightKg  float64 `json:"weight_kg"`
	Reps      int     `json:"reps"`
}

type CardioSet struct {
	DurationSeconds int     `json:"duration_seconds"`
	DistanceM       float64 `json:"distance_m,omitempty"`
}

type WorkoutExercise struct {
	ExerciseID int           `json:"exercise_id"`
	Type       string        `json:"type"`                 // "strength", "cardio", "cardio_gps"
	Sets       []StrengthSet `json:"sets,omitempty"`       // Для силовых
	Cardio     *CardioSet    `json:"cardio,omitempty"`     // Для кардио
	GPSPoints  []string      `json:"gps_points,omitempty"` // Для cardio_gps (в реале тут struct с lat/lon)
}

type CreateWorkoutRequest struct {
	UserID     int               `json:"user_id"`
	StartedAt  string            `json:"started_at"`
	FinishedAt string            `json:"finished_at"`
	Notes      string            `json:"notes"`
	Exercises  []WorkoutExercise `json:"exercises"` // Тот самый массив упражнений
}

// --- УТИЛИТЫ ---

func sendJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// --- USERS (Бизнес-процесс 4.1) ---

func registerUser(w http.ResponseWriter, r *http.Request) {
	// В реале тут: парсинг JSON, хэширование пароля, INSERT INTO users
	sendJSON(w, http.StatusCreated, map[string]string{
		"message": "Пользователь успешно зарегистрирован",
		"token":   "jwt_token_placeholder_12345", // Выдача токена (JWT)
	})
}

func loginUser(w http.ResponseWriter, r *http.Request) {
	// В реале: проверка хэша, выдача JWT
	sendJSON(w, http.StatusOK, map[string]string{
		"message": "Успешный вход",
		"token":   "jwt_token_placeholder_12345",
	})
}

// --- EXERCISES (Справочник) ---

func getExercises(w http.ResponseWriter, r *http.Request) {
	// Отдаем упражнения с типами (согласно нашей БД)
	exercises := []map[string]interface{}{
		{"id": 1, "name": "Жим лежа", "muscle_group": "Грудь", "type": "strength"},
		{"id": 2, "name": "Приседания", "muscle_group": "Ноги", "type": "strength"},
		{"id": 3, "name": "Бег на улице", "muscle_group": nil, "type": "cardio_gps"},
		{"id": 4, "name": "Велотренажер", "muscle_group": nil, "type": "cardio"},
	}
	sendJSON(w, http.StatusOK, exercises)
}

// --- WORKOUTS (Бизнес-процесс 4.2) ---

func getWorkouts(w http.ResponseWriter, r *http.Request) {
	workouts := []map[string]interface{}{
		{"id": 101, "started_at": "2026-11-13T10:00:00Z", "notes": "Утренняя тренировка", "status": "finished"},
		{"id": 102, "started_at": "2026-11-12T18:00:00Z", "notes": "Вечернее кардио", "status": "finished"},
	}
	sendJSON(w, http.StatusOK, workouts)
}

func createWorkout(w http.ResponseWriter, r *http.Request) {
	var req CreateWorkoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendJSON(w, http.StatusBadRequest, map[string]string{"error": "Неверный формат JSON"})
		return
	}

	for _, ex := range req.Exercises {
		if ex.Type == "strength" && len(ex.Sets) == 0 {
			sendJSON(w, http.StatusBadRequest, map[string]string{"error": "Для силового упражнения нужны подходы"})
			return
		}
		if ex.Type == "cardio_gps" && ex.Cardio == nil {
			sendJSON(w, http.StatusBadRequest, map[string]string{"error": "Для GPS кардио нужны данные таймера"})
			return
		}
	}

	sendJSON(w, http.StatusCreated, map[string]string{
		"message":    "Тренировка и все подходы успешно сохранены",
		"workout_id": "103",
	})
}

func getWorkoutByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	workoutDetail := map[string]interface{}{
		"id":         id,
		"notes":      "Хорошая тренировка",
		"started_at": "2026-11-13T10:00:00Z",
		"sets": []map[string]interface{}{
			{"exercise": "Жим лежа", "set_number": 1, "weight_kg": 80.0, "reps": 10},
			{"exercise": "Жим лежа", "set_number": 2, "weight_kg": 85.0, "reps": 8},
			{"exercise": "Бег", "duration_seconds": 1800, "distance_m": 5000.0},
		},
	}
	sendJSON(w, http.StatusOK, workoutDetail)
}

// --- ОТЧЕТЫ (Бизнес-процесс 4.3) ---

func getReports(w http.ResponseWriter, r *http.Request) {
	// В реале: SQL-агрегации (SUM, COUNT, AVG)
	// Эмуляция отчета со встроенной математикой
	report := map[string]interface{}{
		"period":         "week",
		"total_tonnage":  12500.5, // SUM(weight_kg * reps)
		"total_distance": 15000.0, // SUM(distance_m)
		"workouts_count": 3,
	}
	sendJSON(w, http.StatusOK, report)
}

// --- MAIN ---

func main() {
	r := mux.NewRouter()

	// 1. Авторизация (Бизнес-процесс 4.1)
	r.HandleFunc("/api/register", registerUser).Methods("POST")
	r.HandleFunc("/api/login", loginUser).Methods("POST")

	// 2. Справочник
	r.HandleFunc("/api/exercises", getExercises).Methods("GET")

	// 3. Тренировки (Бизнес-процесс 4.2)
	r.HandleFunc("/api/workouts", getWorkouts).Methods("GET")         // История
	r.HandleFunc("/api/workouts", createWorkout).Methods("POST")      // Создание (отправка JSON)
	r.HandleFunc("/api/workouts/{id}", getWorkoutByID).Methods("GET") // Детали

	// 4. Отчеты (Бизнес-процесс 4.3)
	r.HandleFunc("/api/reports", getReports).Methods("GET")

	fmt.Println("Сервер запущен на http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
