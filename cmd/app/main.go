package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func sendJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// --- USERS ---

func getUsers(w http.ResponseWriter, r *http.Request) {
	// Массив пользователей
	users := []string{"Иван Иванов", "Петр Петров", "Анна Смирнова"}
	sendJSON(w, http.StatusOK, users)
}

func getUserByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	// Имитация получения пользователя по ID
	sendJSON(w, http.StatusOK, map[string]string{
		"message": "Пользователь с ID: " + id,
		"name":    "Иван Иванов",
		"email":   "ivan@example.com",
	})
}

// --- EXERCISES ---

func getExercises(w http.ResponseWriter, r *http.Request) {
	// Массив упражнений
	exercises := []string{"Жим лежа", "Приседания", "Становая тяга", "Подтягивания"}
	sendJSON(w, http.StatusOK, exercises)
}

// --- WORKOUTS ---

func getWorkouts(w http.ResponseWriter, r *http.Request) {
	// Массив тренировок
	workouts := []string{
		"Тренировка №1: Силовая, сожжено 300 ккал",
		"Тренировка №2: Кардио, сожжено 500 ккал",
	}
	sendJSON(w, http.StatusOK, workouts)
}

func createWorkout(w http.ResponseWriter, r *http.Request) {
	sendJSON(w, http.StatusCreated, map[string]string{
		"message": "Тренировка успешно создана",
	})
}

func getWorkoutByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	sendJSON(w, http.StatusOK, map[string]string{
		"message": "Детали тренировки с ID: " + id,
		"type":    "gym",
		"notes":   "Хорошая тренировка",
	})
}

// --- GYM LOG ENTRIES ---

func getGymLogEntries(w http.ResponseWriter, r *http.Request) {
	// Массив записей о подходах в зале
	logs := []string{
		"Подход 1: Жим лежа, 100кг x 10 раз",
		"Подход 2: Жим лежа, 105кг x 8 раз",
		"Подход 3: Приседания, 120кг x 5 раз",
	}
	sendJSON(w, http.StatusOK, logs)
}

func createGymLogEntry(w http.ResponseWriter, r *http.Request) {
	sendJSON(w, http.StatusCreated, map[string]string{
		"message": "Запись о подходе добавлена",
	})
}

// --- CARDIO DETAILS ---

func getCardioDetails(w http.ResponseWriter, r *http.Request) {
	// Массив данных кардио
	cardio := []string{
		"Кардио №1: Бег, 5км, 6000 шагов, средняя скорость 10 км/ч",
		"Кардио №2: Ходьба, 3км, 4000 шагов, средняя скорость 5 км/ч",
	}
	sendJSON(w, http.StatusOK, cardio)
}

func createCardioDetails(w http.ResponseWriter, r *http.Request) {
	sendJSON(w, http.StatusCreated, map[string]string{
		"message": "Данные кардио добавлены",
	})
}

// --- GPS POINTS ---

func getGPSPoints(w http.ResponseWriter, r *http.Request) {
	// Массив GPS точек
	points := []string{
		"Точка 1: lat 55.7558, lon 37.6173",
		"Точка 2: lat 55.7560, lon 37.6180",
	}
	sendJSON(w, http.StatusOK, points)
}

func createGPSPoint(w http.ResponseWriter, r *http.Request) {
	sendJSON(w, http.StatusCreated, map[string]string{
		"message": "GPS точка записана",
	})
}

// --- MAIN ---

func main() {

	r := mux.NewRouter()

	// Users
	r.HandleFunc("/users", getUsers).Methods("GET")
	r.HandleFunc("/users/{id}", getUserByID).Methods("GET")

	// Exercises
	r.HandleFunc("/exercises", getExercises).Methods("GET")

	// Workouts
	r.HandleFunc("/workouts", getWorkouts).Methods("GET")
	r.HandleFunc("/workouts", createWorkout).Methods("POST")
	r.HandleFunc("/workouts/{id}", getWorkoutByID).Methods("GET")

	// Gym Log Entries
	r.HandleFunc("/logs", getGymLogEntries).Methods("GET")
	r.HandleFunc("/logs", createGymLogEntry).Methods("POST")

	// Cardio Details
	r.HandleFunc("/cardio", getCardioDetails).Methods("GET")
	r.HandleFunc("/cardio", createCardioDetails).Methods("POST")

	// GPS Points
	r.HandleFunc("/gps", getGPSPoints).Methods("GET")
	r.HandleFunc("/gps", createGPSPoint).Methods("POST")

	fmt.Println("Сервер запущен на http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
