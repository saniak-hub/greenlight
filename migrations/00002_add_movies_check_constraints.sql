-- +goose Up
ALTER TABLE movies
    ADD CONSTRAINT chk_movies_runtime CHECK (runtime >= 0),
    ADD CONSTRAINT chk_movies_year CHECK (year >= 1888 AND year <= EXTRACT(YEAR FROM NOW())),
    ADD CONSTRAINT chk_movies_genres CHECK (array_length(genres, 1) BETWEEN 1 AND 5);

-- +goose Down
ALTER TABLE movies
    DROP CONSTRAINT IF EXISTS chk_movies_runtime,
    DROP CONSTRAINT IF EXISTS chk_movies_year,
    DROP CONSTRAINT IF EXISTS chk_movies_genres;
