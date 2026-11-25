package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/richhomi31/fit_tracker_go.git/internal/app/repository"
)

type api struct {
	users     *repository.UserRepository
	exercises *repository.ExerciseRepository
	workouts  *repository.WorkoutRepository
}

func newRouter(db *sqlx.DB) *mux.Router {
	a := api{
		users:     repository.NewUserRepository(db),
		exercises: repository.NewExerciseRepository(db),
		workouts:  repository.NewWorkoutRepository(db),
	}
	r := mux.NewRouter()
	r.HandleFunc("/api/users", a.register).Methods(http.MethodPost)
	r.HandleFunc("/api/login", a.login).Methods(http.MethodPost)
	r.HandleFunc("/api/users/{id}", a.auth(a.user)).Methods(http.MethodGet, http.MethodPut, http.MethodDelete)
	r.HandleFunc("/api/exercises", a.auth(a.exerciseList)).Methods(http.MethodGet, http.MethodPost)
	r.HandleFunc("/api/exercises/{id}", a.auth(a.exercise)).Methods(http.MethodGet, http.MethodPut, http.MethodDelete)
	r.HandleFunc("/api/workouts", a.auth(a.workoutList)).Methods(http.MethodGet, http.MethodPost)
	r.HandleFunc("/api/workouts/{id}", a.auth(a.workout)).Methods(http.MethodGet, http.MethodPut, http.MethodDelete)
	r.HandleFunc("/api/reports/volume", a.auth(a.report("volume"))).Methods(http.MethodGet)
	r.HandleFunc("/api/reports/max-weight", a.auth(a.report("max-weight"))).Methods(http.MethodGet)
	r.HandleFunc("/api/reports/distance", a.auth(a.report("distance"))).Methods(http.MethodGet)
	return r
}

func sendJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func fail(w http.ResponseWriter, status int, message string) {
	sendJSON(w, status, map[string]string{"error": message})
}

func decode(w http.ResponseWriter, r *http.Request, dst any) bool {
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4<<20))
	if decoder.Decode(dst) != nil || decoder.Decode(new(json.RawMessage)) != io.EOF {
		fail(w, http.StatusBadRequest, "Неверный JSON")
		return false
	}
	return true
}

func routeID(w http.ResponseWriter, r *http.Request) (int, bool) {
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil || id <= 0 || id > 2147483647 {
		fail(w, http.StatusBadRequest, "Неверный ID")
		return 0, false
	}
	return id, true
}

func databaseError(w http.ResponseWriter, err error) {
	var pg *pq.Error
	switch {
	case errors.Is(err, sql.ErrNoRows):
		fail(w, http.StatusNotFound, "Запись не найдена")
	case errors.As(err, &pg) && (pg.Code == "23505" || pg.Code == "23503" || pg.Code == "23001"):
		fail(w, http.StatusConflict, "Нарушена уникальность или связь с другой записью")
	case errors.As(err, &pg) && (pg.Code.Class() == "22" || pg.Code == "23514"):
		fail(w, http.StatusBadRequest, "Данные не соответствуют ограничениям базы")
	default:
		fail(w, http.StatusInternalServerError, "Ошибка базы данных")
	}
}

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("Укажите DATABASE_URL")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	db, err := sqlx.ConnectContext(ctx, "postgres", dsn)
	cancel()
	if err != nil {
		log.Fatal("Ошибка подключения к PostgreSQL: ", err)
	}
	defer db.Close()

	address := os.Getenv("HTTP_ADDR")
	if address == "" {
		address = "localhost:8080"
	}
	server := &http.Server{Addr: address, Handler: newRouter(db), ReadHeaderTimeout: 5 * time.Second}
	log.Printf("Сервер запущен на http://%s", address)
	if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
}
