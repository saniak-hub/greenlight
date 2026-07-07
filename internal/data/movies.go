package data

import (
	"context"
	"database/sql"
	"time"

	"github.com/lib/pq"
	"github.com/saniak-hub/greenlight/internal/validator"
)

type Movie struct {
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	Title     string    `json:"title"`
	Year      int       `json:"year"`
	Runtime   Runtime   `json:"runtime"`
	Genres    []string  `json:"genres"`
	Version   int       `json:"version"`
}

type MovieModel struct {
	DB *sql.DB
}

func (m MovieModel) Insert(movie *Movie) error {
	query := `
		INSERT INTO movies (title, year, runtime, genres)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, version`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return m.DB.QueryRowContext(ctx, query,
		movie.Title,
		movie.Year,
		int(movie.Runtime),
		pq.Array(movie.Genres),
	).Scan(&movie.ID, &movie.CreatedAt, &movie.Version)
}

func (m MovieModel) Get(id int64) (*Movie, error) {
	query := `
		SELECT id, created_at, title, year, runtime, genres, version
		FROM movies
		WHERE id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var movie Movie

	err := m.DB.QueryRowContext(ctx, query, id).Scan(
		&movie.ID,
		&movie.CreatedAt,
		&movie.Title,
		&movie.Year,
		&movie.Runtime,
		pq.Array(&movie.Genres),
		&movie.Version,
	)
	if err != nil {
		switch {
		case err == sql.ErrNoRows:
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &movie, nil
}

func (m MovieModel) Update(movie *Movie) error {
	query := `
		UPDATE movies
		SET title=$1, year=$2, runtime=$3, genres=$4, version = version + 1
		WHERE id = $5 AND version = $6
		RETURNING id, created_at, title, year, runtime, genres, version`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query,
		movie.Title,
		movie.Year,
		int(movie.Runtime),
		pq.Array(movie.Genres),
		movie.ID,
		movie.Version,
	).Scan(
		&movie.ID,
		&movie.CreatedAt,
		&movie.Title,
		&movie.Year,
		&movie.Runtime,
		pq.Array(&movie.Genres),
		&movie.Version,
	)
	if err != nil {
		switch {
		case err == sql.ErrNoRows:
			return ErrEditConflict
		default:
			return err
		}
	}

	return nil
}

func (m MovieModel) Delete(id int64) error {
	query := `
		DELETE FROM movies
		WHERE id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := m.DB.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrRecordNotFound
	}

	return nil
}

func (m MovieModel) GetAll(title string, genres []string, filters Filters) ([]*Movie, Metadata, error) {
	query := `
		SELECT count(*) OVER() AS totalRecords, id, created_at, title, year, runtime, genres, version
		FROM movies
		WHERE (to_tsvector('simple', title) @@ plainto_tsquery('simple', $1) OR $1 = '')
		    AND (genres @> $2::text[] OR array_length($2::text[], 1) IS NULL)
		ORDER BY
		    CASE WHEN $3::text ='title' AND $4::text = 'ASC' THEN title END ASC,
		    CASE WHEN $3::text ='year' AND $4::text = 'ASC' THEN year END ASC,
		    CASE WHEN $3::text ='runtime' AND $4::text = 'ASC' THEN runtime END ASC,
		    CASE WHEN $3::text ='title' AND $4::text = 'DESC' THEN title END DESC,
		    CASE WHEN $3::text ='year' AND $4::text = 'DESC' THEN year END DESC,
		    CASE WHEN $3::text ='runtime' AND $4::text = 'DESC' THEN runtime END DESC,
		    id ASC
		LIMIT $6::int OFFSET $5::int`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query,
		title,
		pq.Array(genres),
		filters.SortColumn(),
		filters.SortDirection(),
		filters.PageOffset(),
		filters.PageSize,
	)
	if err != nil {
		return nil, Metadata{}, err
	}
	defer rows.Close()

	var totalRecords int
	movies := []*Movie{}

	for rows.Next() {
		var movie Movie
		err := rows.Scan(
			&totalRecords,
			&movie.ID,
			&movie.CreatedAt,
			&movie.Title,
			&movie.Year,
			&movie.Runtime,
			pq.Array(&movie.Genres),
			&movie.Version,
		)
		if err != nil {
			return nil, Metadata{}, err
		}
		movies = append(movies, &movie)
	}

	if err = rows.Err(); err != nil {
		return nil, Metadata{}, err
	}

	metadata := CalculateMetadata(totalRecords, filters.Page, filters.PageSize)

	return movies, metadata, nil
}

func ValidateMovie(v *validator.Validator, title string, year int, runtime int, genres []string) {
	v.Check(title != "", "title", "must be provided")
	v.Check(len(title) <= 500, "title", "must not be more than 500 bytes long")

	v.Check(year != 0, "year", "must be provided")
	v.Check(year >= 1888, "year", "must be greater than 88")
	v.Check(year <= time.Now().Year(), "year", "must not be in the future")

	v.Check(runtime != 0, "runtime", "must be provided")
	v.Check(runtime > 0, "runtime", "must be a positive integer")

	v.Check(genres != nil, "genres", "must be provided")
	v.Check(len(genres) >= 1, "genres", "must contain at least 1 genre")
	v.Check(len(genres) <= 5, "genres", "must not contain more than 5 genres")
	v.Check(validator.Unique(genres), "genres", "must not contain duplicate values")
}
