package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

type UserCredentials struct {
	Email    string `json:"email"` // используем email как логин
	Password string `json:"password"`
}

type Exercise struct {
	Name        string `json:"name"`
	MuscleGroup string `json:"muscle_group"`
	Type        string `json:"type"` // "strength", "cardio", "cardio_gps"
}

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
	Type       string        `json:"type"`
	Sets       []StrengthSet `json:"sets,omitempty"`
	Cardio     *CardioSet    `json:"cardio,omitempty"`
}

type CreateWorkoutRequest struct {
	UserID     int               `json:"user_id"`
	StartedAt  string            `json:"started_at"`
	FinishedAt string            `json:"finished_at"`
	Notes      string            `json:"notes"`
	Exercises  []WorkoutExercise `json:"exercises"`
}

func sendJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// USERS
// Регистрация
func registerUser(w http.ResponseWriter, r *http.Request) {
	var creds UserCredentials
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		sendJSON(w, http.StatusBadRequest, map[string]string{"error": "Неверный формат данных"})
		return
	}
	// В реале: проверка уникальности, хэширование, INSERT INTO users
	sendJSON(w, http.StatusCreated, map[string]string{
		"message": "Пользователь зарегистрирован",
		"user_id": "1",
	})
}

// Вход
func loginUser(w http.ResponseWriter, r *http.Request) {
	var creds UserCredentials
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		sendJSON(w, http.StatusBadRequest, map[string]string{"error": "Неверный формат данных"})
		return
	}

	sendJSON(w, http.StatusOK, map[string]string{
		"message": "Успешный вход",
		"user_id": "1",
		"name":    "Иван",
	})
}

// удаление
func deleteUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["id"]
	// В реале: DELETE FROM users WHERE id = ? (с ON DELETE CASCADE удалятся тренировки)
	sendJSON(w, http.StatusOK, map[string]string{
		"message": "Пользователь " + userID + " удален",
	})
}

// EXERCISES
// Просмотр
func getExercises(w http.ResponseWriter, r *http.Request) {
	exercises := []map[string]interface{}{
		{"id": 1, "name": "Жим лежа", "type": "strength"},
		{"id": 2, "name": "Бег на улице", "type": "cardio_gps"},
	}
	sendJSON(w, http.StatusOK, exercises)
}

// Создание
func createExercise(w http.ResponseWriter, r *http.Request) {
	var ex Exercise
	if err := json.NewDecoder(r.Body).Decode(&ex); err != nil {
		sendJSON(w, http.StatusBadRequest, map[string]string{"error": "Неверный формат"})
		return
	}
	sendJSON(w, http.StatusCreated, map[string]string{
		"message": "Упражнение '" + ex.Name + "' добавлено",
	})
}

// Удаление
func deleteExercise(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	exID := vars["id"]
	sendJSON(w, http.StatusOK, map[string]string{
		"message": "Упражнение " + exID + " удалено",
	})
}

// WORKOUTS
// Создание
func createWorkout(w http.ResponseWriter, r *http.Request) {
	var req CreateWorkoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendJSON(w, http.StatusBadRequest, map[string]string{"error": "Неверный JSON"})
		return
	}
	// Валидация и INSERT в workouts, workout_sets, gps_points
	sendJSON(w, http.StatusCreated, map[string]string{
		"message":    "Тренировка сохранена",
		"workout_id": "100",
	})
}

// Обновление, т.е. полная перезапись
func updateWorkout(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	workoutID := vars["id"]

	var req CreateWorkoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendJSON(w, http.StatusBadRequest, map[string]string{"error": "Неверный JSON"})
		return
	}

	sendJSON(w, http.StatusOK, map[string]string{
		"message": "Тренировка " + workoutID + " успешно обновлена",
	})
}

// Удаление
func deleteWorkout(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	workoutID := vars["id"]
	sendJSON(w, http.StatusOK, map[string]string{
		"message": "Тренировка " + workoutID + " удалена",
	})
}

// ОТЧЕТЫ

func getReport1(w http.ResponseWriter, r *http.Request) {
	// Отчет 1 Прогресс по объему работы
	// Тоннаж = Сумма (вес * повторения) за каждый месяц

	// типо данные для графика
	reportData := map[string]interface{}{
		"report_type": "volume_progress",
		"exercise":    "Жим лежа",
		"period":      "month",
		"labels":      []string{"Июль", "Август", "Сентябрь", "Октябрь"}, // Ось X
		"data":        []float64{1500.5, 1800.0, 2100.5, 2500.0},         // Ось Y (тоннаж)
		"unit":        "кг",
	}
	sendJSON(w, http.StatusOK, reportData)
}

func getReport2(w http.ResponseWriter, r *http.Request) {
	// Отчет 2 Прогресс максимального веса по упражнениям
	// Список упражнений, по месяцам макс вес и приросn %

	reportData := map[string]interface{}{
		"report_type": "max_weight_progress",
		"period":      "month",
		"exercise_progress": []map[string]interface{}{
			{
				"exercise_name": "Жим лежа",
				"data": []map[string]interface{}{
					{"month": "Сентябрь", "max_weight": 100.0, "change_percent": 0.0},
					{"month": "Октябрь", "max_weight": 105.0, "change_percent": 5.0}, // +5%
				},
			},
			{
				"exercise_name": "Приседания",
				"data": []map[string]interface{}{
					{"month": "Сентябрь", "max_weight": 120.0, "change_percent": 0.0},
					{"month": "Октябрь", "max_weight": 115.0, "change_percent": -4.16}, // -4.16%
				},
			},
		},
	}
	sendJSON(w, http.StatusOK, reportData)
}

func main() {
	r := mux.NewRouter()

	// Пользователи
	r.HandleFunc("/api/users", registerUser).Methods("POST") // Регистрация
	r.HandleFunc("/api/login", loginUser).Methods("POST")    // Вход
	r.HandleFunc("/api/users/{id}", deleteUser).Methods("DELETE")

	// Упражнения
	r.HandleFunc("/api/exercises", getExercises).Methods("GET")
	r.HandleFunc("/api/exercises", createExercise).Methods("POST")
	r.HandleFunc("/api/exercises/{id}", deleteExercise).Methods("DELETE")

	// Тренировки
	r.HandleFunc("/api/workouts", createWorkout).Methods("POST")
	r.HandleFunc("/api/workouts/{id}", deleteWorkout).Methods("DELETE")
	r.HandleFunc("/api/workouts/{id}", updateWorkout).Methods("PUT")

	// Отчеты
	r.HandleFunc("/api/reports/volume", getReport1).Methods("GET")
	r.HandleFunc("/api/reports/max-weight", getReport2).Methods("GET")

	fmt.Println("Сервер запущен на http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
