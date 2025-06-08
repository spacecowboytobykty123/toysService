package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/lib/pq"
	"github.com/spacecowboytobykty123/toysProto/gen/go/toys"
	"log"
	"strings"
	"time"
	"toysService/internal/data"
	"toysService/internal/jsonlog"
	"toysService/internal/validator"
)

type Storage struct {
	db  *sql.DB
	log *jsonlog.Logger
}

const (
	emptyValue = 0
)

type StorageDetails struct {
	DSN          string
	MaxOpenConns int
	MaxIdleConns int
	MaxIdleTime  string
}

func OpenDB(details StorageDetails, logger *jsonlog.Logger) (*Storage, error) {
	var db *sql.DB
	var err error
	for i := 0; i < 10; i++ {
		db, err = sql.Open("postgres", details.DSN)
		if err == nil {
			err = db.Ping()
		}
		if err == nil {
			break
		}
		time.Sleep(2 * time.Second)
		log.Printf("retrying DB connection... (%d/10)", i+1)
	}

	if err != nil {

		log.Fatal("failed to connect to database after retries:", err)
	}

	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(details.MaxOpenConns)
	db.SetMaxIdleConns(details.MaxIdleConns)

	duration, err := time.ParseDuration(details.MaxIdleTime)

	db.SetConnMaxIdleTime(duration)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = db.PingContext(ctx)
	if err != nil {
		return nil, err
	}
	return &Storage{
		db:  db,
		log: logger,
	}, err
}

func ValidateToy(v *validator.Validator, toy *data.Toy) {
	v.Check(toy.Title != "", "title", "title must be provided")
	v.Check(len(toy.Title) <= 500, "title", "title must not be more than 500 bytes long")
	v.Check(len(toy.Desc) <= 5000, "desc", "Description must not be more than 5000 bytes long")
	//v.Check(v.ImageUrlsCheck(toy.Images), "image", "some of image urls is wrong")
	v.Check(toy.Categories != nil, "categories", "categories must be provided")
	v.Check(toy.Skills != nil, "skills", "skills must be provided")
	v.Check(len(toy.Categories) >= 1, "categories", "at least 1 category")
	v.Check(len(toy.Skills) >= 1, "skills", "at least 1 skill")
	v.Check(len(toy.Categories) <= 7, "categories", "no more than 7 categories")
	v.Check(len(toy.Skills) <= 7, "Skills", "no more than 7 skills")
	v.Check(validator.Unique(toy.Categories), "categories", "categories should not contain duplicate values")
	v.Check(validator.Unique(toy.Skills), "skills", "skills should not contain duplicate values")
	v.Check(toy.RecAge != "", "recAge", "age must be provided")
	v.Check(toy.Manufacturer != "", "manufacturer", "manufacturer must be provided")
	v.Check(toy.Value >= 2000, "value", "toy value must be more than 1000 tenge")
	v.Check(toy.Value <= 150000, "value", "limit of toy's value is 150.000 tenge")
}

func (s *Storage) CreateToy(ctx context.Context, inputToy data.Toy) (toys.Status, string, data.Toy) {
	s.log.PrintInfo("DB part", map[string]string{
		"method": "postgres.CreateToy",
	})
	query := `
INSERT INTO toys (title, description, skills, categories, images, recommended_age, manufacturer, value, is_available)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING id`

	args := []any{inputToy.Title, inputToy.Desc, pq.Array(inputToy.Skills), pq.Array(inputToy.Categories), pq.Array(inputToy.Images), inputToy.RecAge, inputToy.Manufacturer, inputToy.Value, inputToy.IsAvailable}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	var toyID int64
	err := s.db.QueryRowContext(ctx, query, args...).Scan(&toyID)
	println("skibidi")
	if err != nil {
		return toys.Status_STATUS_INTERNAL_ERROR, "internal error!", data.Toy{}
	}
	toy := data.Toy{
		ID:           toyID,
		Title:        inputToy.Title,
		Desc:         inputToy.Desc,
		Value:        inputToy.Value,
		Images:       inputToy.Images,
		Skills:       inputToy.Skills,
		Categories:   inputToy.Categories,
		RecAge:       inputToy.RecAge,
		Manufacturer: inputToy.Manufacturer,
		IsAvailable:  inputToy.IsAvailable,
	}
	return toys.Status_STATUS_OK, "toy added successfuly!", toy
}

func (s *Storage) DeleteToy(ctx context.Context, toyID int64) (toys.Status, string) {
	s.log.PrintInfo("DB part", map[string]string{
		"method": "postgres.DeleteToy",
	})
	query := `
DELETE FROM toys
WHERE id = $1
`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := s.db.ExecContext(ctx, query, toyID)

	if err != nil {
		return toys.Status_STATUS_INTERNAL_ERROR, "internal error!"
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return toys.Status_STATUS_INTERNAL_ERROR, "internal error!"
	}

	if rowsAffected == 0 {
		return toys.Status_STATUS_INTERNAL_ERROR, "it affected 0 rows!"
	}
	return toys.Status_STATUS_OK, "toy deletion was successful"
}

func (s *Storage) ChangeToy(ctx context.Context, toy data.Toy) (toys.Status, string) {
	s.log.PrintInfo("DB part", map[string]string{
		"method": "postgres.ChangeToy",
	})
	query := `UPDATE toys
SET title = $1, description = $2, skills = $3, images = $4, categories = $5, recommended_age = $6, manufacturer = $7, value = $8
WHERE id = $9
RETURNING id
`
	args := []any{
		toy.Title,
		toy.Desc,
		pq.Array(toy.Skills),
		pq.Array(toy.Images),
		pq.Array(toy.Categories),
		toy.RecAge,
		toy.Manufacturer,
		toy.Value,
		toy.ID,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := s.db.QueryRowContext(ctx, query, args...).Scan(&toy.ID)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return toys.Status_STATUS_INTERNAL_ERROR, "operation affect zero rows!"
		default:
			return toys.Status_STATUS_INTERNAL_ERROR, "internal error!"
		}

	}
	return toys.Status_STATUS_OK, "toys updated successfully!"
}

func (s *Storage) GetToy(ctx context.Context, toyID int64) (data.Toy, toys.Status, string) {
	s.log.PrintInfo("DB part", map[string]string{
		"method": "postgres.GetToy",
	})
	if toyID < 1 {
		return data.Toy{}, toys.Status_STATUS_INTERNAL_ERROR, "invalid toy id"
	}

	query := `
SELECT id, created_at, title, description ,skills, categories, images, recommended_age, manufacturer, value, is_available
FROM toys
WHERE id = $1
`

	var toy data.Toy

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := s.db.QueryRowContext(ctx, query, toyID).Scan(
		&toy.ID,
		&toy.CreatedAt,
		&toy.Title,
		&toy.Desc,
		pq.Array(&toy.Skills),
		pq.Array(&toy.Categories),
		pq.Array(&toy.Images),
		&toy.RecAge,
		&toy.Manufacturer,
		&toy.Value,
		&toy.IsAvailable,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return data.Toy{}, toys.Status_STATUS_INTERNAL_ERROR, "operation affect zero rows!"
		default:
			return data.Toy{}, toys.Status_STATUS_INTERNAL_ERROR, "internal error!"
		}
	}

	return toy, toys.Status_STATUS_OK, "toy get successfully"
}

func (s *Storage) ListToy(
	ctx context.Context,
	to int64,
	from int64,
	filters data.Filters,
	categories []string,
	skills []string,
	title string,
) ([]*data.Toy, toys.Status, string, data.Metadata) {

	s.log.PrintInfo("DB part", map[string]string{
		"method": "postgres.ListToy",
	})

	query := `
SELECT count(*) OVER(), id, title, categories, skills, recommended_age, value
FROM toys
WHERE (to_tsvector('simple', title) @@ plainto_tsquery('simple', $1) OR $1 = '')`
	args := []any{title}
	argIndex := 2

	if len(categories) > 0 {
		query += fmt.Sprintf(" AND categories @> $%d", argIndex)
		args = append(args, pq.Array(categories))
		argIndex++
	}

	if len(skills) > 0 {
		query += fmt.Sprintf(" AND skills @> $%d", argIndex)
		args = append(args, pq.Array(skills))
		argIndex++
	}

	query += fmt.Sprintf(" AND value BETWEEN $%d AND $%d", argIndex, argIndex+1)
	args = append(args, from, to)
	argIndex += 2

	query += fmt.Sprintf(" ORDER BY %s %s, id ASC LIMIT $%d OFFSET $%d", filters.SortColumn(), filters.SortDirection(), argIndex, argIndex+1)
	args = append(args, filters.Limit(), filters.Offset())

	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, toys.Status_STATUS_INTERNAL_ERROR, "could not fetch toys from db", data.Metadata{}
	}
	defer rows.Close()

	totalRecords := 0
	toysList := []*data.Toy{}

	for rows.Next() {
		var toy data.Toy
		err := rows.Scan(
			&totalRecords,
			&toy.ID,
			&toy.Title,
			pq.Array(&toy.Categories),
			pq.Array(&toy.Skills),
			&toy.RecAge,
			&toy.Value,
		)
		if err != nil {
			return nil, toys.Status_STATUS_INTERNAL_ERROR, "could not fetch toys from db", data.Metadata{}
		}
		toysList = append(toysList, &toy)
	}

	if err = rows.Err(); err != nil {
		return nil, toys.Status_STATUS_INTERNAL_ERROR, "could not fetch toys from db", data.Metadata{}
	}

	metadata := filters.CalculateMetadata(totalRecords, filters.Page, filters.PageSize)
	return toysList, toys.Status_STATUS_OK, "toy listing was successful", metadata
}

func (s *Storage) GetToysByIds(ctx context.Context, ids []int64) ([]*data.ToySummary, string) {
	s.log.PrintInfo("DB part", map[string]string{
		"method": "postgres.gettoysbyid",
	})
	placeholders := make([]string, len(ids))
	args := make([]interface{}, len(ids))
	for i, id := range ids {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = id
	}
	query := fmt.Sprintf(`SELECT id, title, value, images[1] AS image_url FROM toys WHERE id IN (%s)`, strings.Join(placeholders, ","))

	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return []*data.ToySummary{}, "could not query"
	}
	defer rows.Close()
	results := []*data.ToySummary{}

	for rows.Next() {
		var toy data.ToySummary
		err := rows.Scan(
			&toy.ID,
			&toy.Title,
			&toy.Value,
			&toy.URL,
		)
		println(toy.URL)
		println(toy.Value)
		if err != nil {
			return nil, "internal error"
		}

		results = append(results, &toy)
	}
	if err = rows.Err(); err != nil {
		return nil, "internal"
	}

	return results, "toy fetch was successful"

}

func (s *Storage) ListRecommended(ctx context.Context, userID int64) ([]*data.Toy, toys.Status, string, data.Metadata) {
	s.log.PrintInfo("DB part", map[string]string{
		"method": "postgres.ListRec",
	})
	query := `SELECT id, title, categories, recommended_age, value FROM toys
WHERE categories = $1
AND (categories @> $2 OR $2 = '{}')
AND (skills @> $3 OR $3 = '{}')
`

	// TODO: брать с пользователя категории и скилы игрушек которые он уже купил

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	panic("potom")
	panic(query)
}
