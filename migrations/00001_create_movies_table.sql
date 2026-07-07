-- +goose Up
CREATE TABLE IF NOT EXISTS movies (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    title TEXT NOT NULL,
    year INT NOT NULL,
    runtime INT NOT NULL,
    genres TEXT[] NOT NULL,
    version INT NOT NULL DEFAULT 1
);

-- +goose Down
DROP TABLE IF EXISTS movies;
