package postgres

import (
	"context"
	"database/sql"
	"github.com/spacecowboytobykty123/toysProto/gen/go/toys"
	"time"
	"toysService/internal/data"
	"toysService/internal/validator"
)

type Storage struct {
	db *sql.DB
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

func OpenDB(details StorageDetails) (*Storage, error) {
	db, err := sql.Open("postgres", details.DSN)

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
		db: db,
	}, err
}

func ValidateToy(v *validator.Validator, toy *data.Toy) {
	v.Check(toy.Title != "", "title", "title must be provided")
	v.Check(len(toy.Title) <= 500, "title", "title must not be more than 500 bytes long")
	v.Check(len(toy.Desc) <= 5000, "desc", "Description must not be more than 5000 bytes long")
	v.Check(v.ImageUrlsCheck(toy.Images), "image", "some of image urls is wrong")
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
	v.Check(toy.Value >= 1000, "value", "toy value must be more than 1000 tenge")
	v.Check(toy.Value <= 150000, "value", "limit of toy's value is 150.000 tenge")
}

func (s Storage) CreateToy(ctx context.Context, title string, desc string, value int64, images []string, skills []string, categories []string, recAge string, manufacturer string, isAvailable bool) (toys.Status, string, data.Toy) {
	//TODO implement me
	panic("implement me")
}

func (s Storage) DeleteToy(ctx context.Context, toyID int64) (toys.Status, string) {
	//TODO implement me
	panic("implement me")
}

func (s Storage) ChangeToy(ctx context.Context, toy data.Toy) (toys.Status, string) {
	//TODO implement me
	panic("implement me")
}

func (s Storage) GetToy(ctx context.Context, toyID int64) (data.Toy, toys.Status, string) {
	//TODO implement me
	panic("implement me")
}

func (s Storage) ListToy(ctx context.Context, to int64, from int64, filters data.Filters) ([]data.Toy, toys.Status, string, data.Metadata) {
	//TODO implement me
	panic("implement me")
}

func (s Storage) ListRecommended(ctx context.Context) ([]data.Toy, toys.Status, string, data.Metadata) {
	//TODO implement me
	panic("implement me")
}
