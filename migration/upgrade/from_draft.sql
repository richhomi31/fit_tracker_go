-- Обновление схемы, созданной ранним черновиком лабораторной.
-- Старые даты без часового пояса интерпретируются как UTC.
BEGIN;

ALTER TABLE users
    ALTER COLUMN gender DROP NOT NULL,
    ALTER COLUMN registered_at TYPE TIMESTAMPTZ USING registered_at AT TIME ZONE 'UTC';

ALTER TABLE workouts
    ALTER COLUMN started_at TYPE TIMESTAMPTZ USING started_at AT TIME ZONE 'UTC',
    ALTER COLUMN finished_at TYPE TIMESTAMPTZ USING finished_at AT TIME ZONE 'UTC',
    ADD CHECK (finished_at >= started_at);

ALTER TABLE workout_sets
    DROP CONSTRAINT chk_sets_shape,
    ADD CHECK (set_number > 0 AND reps > 0 AND weight_kg >= 0),
    ADD CHECK (duration_seconds > 0 AND distance_m >= 0),
    ADD CONSTRAINT chk_sets_shape CHECK (
        (set_number IS NOT NULL AND weight_kg IS NOT NULL AND reps IS NOT NULL AND duration_seconds IS NULL AND distance_m IS NULL)
        OR
        (set_number IS NULL AND weight_kg IS NULL AND reps IS NULL AND duration_seconds IS NOT NULL)
    );

ALTER TABLE gps_points
    ALTER COLUMN recorded_at TYPE TIMESTAMPTZ USING recorded_at AT TIME ZONE 'UTC',
    ADD CHECK (lat BETWEEN -90 AND 90 AND lon BETWEEN -180 AND 180);

INSERT INTO exercises (name, muscle_group, type)
SELECT seed.name, seed.muscle_group, seed.type
FROM (VALUES
    ('Жим лёжа', 'грудь', 'strength'),
    ('Велотренажёр', NULL, 'cardio'),
    ('Бег на улице', NULL, 'cardio_gps')
) AS seed(name, muscle_group, type)
WHERE NOT EXISTS (
    SELECT 1 FROM exercises e WHERE NOT e.is_custom AND e.name = seed.name AND e.type = seed.type
);

COMMIT;
