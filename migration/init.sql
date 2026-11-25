BEGIN;

-- 1. Таблица пользователей
CREATE TABLE users (
    id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    email VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    gender CHAR(1),
    registered_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 2. Таблица упражнений
CREATE TABLE exercises (
    id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    muscle_group VARCHAR(50),
    type VARCHAR(20) NOT NULL,
    is_custom BOOLEAN NOT NULL DEFAULT FALSE,
    created_by INT,
    CONSTRAINT fk_exercises_user FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE SET NULL,
    CONSTRAINT chk_exercises_type CHECK (type IN ('strength', 'cardio', 'cardio_gps'))
);

-- 3. Таблица тренировок
CREATE TABLE workouts (
    id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id INT NOT NULL,
    started_at TIMESTAMPTZ NOT NULL,
    finished_at TIMESTAMPTZ,
    notes TEXT,
    CHECK (finished_at >= started_at),
    CONSTRAINT fk_workouts_users FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- 4. Таблица подходов/сессий
CREATE TABLE workout_sets (
    id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    workout_id INT NOT NULL,
    exercise_id INT NOT NULL,
    set_number INT,
    weight_kg DECIMAL(6,2),
    reps INT,
    duration_seconds INT,
    distance_m DECIMAL(10,2),
    CONSTRAINT fk_sets_workouts FOREIGN KEY (workout_id) REFERENCES workouts(id) ON DELETE CASCADE,
    CONSTRAINT fk_sets_exercises FOREIGN KEY (exercise_id) REFERENCES exercises(id) ON DELETE RESTRICT,
    CHECK (set_number > 0 AND reps > 0 AND weight_kg >= 0),
    CHECK (duration_seconds > 0 AND distance_m >= 0),
    CONSTRAINT chk_sets_shape CHECK (
        (set_number IS NOT NULL AND weight_kg IS NOT NULL AND reps IS NOT NULL AND duration_seconds IS NULL AND distance_m IS NULL)
        OR
        (set_number IS NULL AND weight_kg IS NULL AND reps IS NULL AND duration_seconds IS NOT NULL)
    )
);

-- 5. Таблица GPS точек
CREATE TABLE gps_points (
    id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    set_id INT NOT NULL,
    lat DECIMAL(9,6) NOT NULL,
    lon DECIMAL(9,6) NOT NULL,
    recorded_at TIMESTAMPTZ NOT NULL,
    CHECK (lat BETWEEN -90 AND 90 AND lon BETWEEN -180 AND 180),
    CONSTRAINT fk_gps_sets FOREIGN KEY (set_id) REFERENCES workout_sets(id) ON DELETE CASCADE
);

-- ИНДЕКСЫ для ускорения отчетов (Бизнес-процесс 4.3)
CREATE INDEX idx_workouts_user_date ON workouts(user_id, started_at);
CREATE INDEX idx_sets_workout ON workout_sets(workout_id);
CREATE INDEX idx_sets_exercise ON workout_sets(exercise_id);
CREATE INDEX idx_gps_set ON gps_points(set_id, recorded_at);

INSERT INTO exercises (name, muscle_group, type) VALUES
    ('Жим лёжа', 'грудь', 'strength'),
    ('Велотренажёр', NULL, 'cardio'),
    ('Бег на улице', NULL, 'cardio_gps');

COMMIT;
