package service

import (
	"database/sql"
	"time"

	"github.com/DaniloReisRocha/grid-tactics/backend/internal/utils/database"
	"github.com/DaniloReisRocha/grid-tactics/backend/models"
	sq "github.com/elgris/sqrl"
	"github.com/gofrs/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

type ExampleCreateService struct {
	DB *sqlx.DB
}

func (c *ExampleCreateService) Run(form models.ExampleDatabaseForm) (*models.ExampleDatabaseView, error) {
	if err := form.Validate(); err != nil {
		return nil, err
	}
	var id uuid.UUID
	err := database.WithinTransaction(c.DB, func(tx *sqlx.Tx) error {
		ex, err := createExample(tx, form)
		if err != nil {
			return err
		}
		id = ex.ID

		return nil
	})
	if err != nil {
		return nil, err
	}

	return getExample(c.DB, id)
}

func createExample(db *sqlx.Tx, form models.ExampleDatabaseForm) (*models.ExampleDatabase, error) {

	id, err := uuid.NewV7()
	if err != nil {
		return nil, errors.Wrap(err, "generate id")
	}

	query := psql.Insert("example_database").
		Columns("id", `"name"`).
		Values(id, form.Name).
		Returning("id", `"name"`, "created_at", "deleted_at")

	qSQL, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}

	resp := &models.ExampleDatabase{}
	err = db.Get(resp, qSQL, args...)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

type ExampleUpdateService struct {
	DB *sqlx.DB
}

func (c *ExampleUpdateService) Run(id uuid.UUID, form models.ExampleDatabaseForm) (*models.ExampleDatabaseView, error) {
	if err := form.Validate(); err != nil {
		return nil, err
	}

	err := database.WithinTransaction(c.DB, func(tx *sqlx.Tx) error {
		_, err := updateExample(tx, id, form)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return getExample(c.DB, id)
}

func updateExample(db *sqlx.Tx, id uuid.UUID, form models.ExampleDatabaseForm) (*models.ExampleDatabase, error) {

	query := psql.Update("example_database").
		Set(`"name"`, form.Name).
		Where(sq.Eq{"id": id}).
		Where(sq.Eq{"deleted_at": nil}).
		Returning("id", `"name"`, "created_at", "deleted_at")

	qSQL, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}

	resp := &models.ExampleDatabase{}
	err = db.Get(resp, qSQL, args...)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

type ExampleDeleteService struct {
	DB *sqlx.DB
}

func (c *ExampleDeleteService) Run(id uuid.UUID) (*models.ExampleDatabaseView, error) {

	err := database.WithinTransaction(c.DB, func(tx *sqlx.Tx) error {
		_, err := deleteExample(tx, id)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return getExample(c.DB, id)
}

func deleteExample(db *sqlx.Tx, id uuid.UUID) (*models.ExampleDatabase, error) {

	query := psql.Update("example_database").
		Set("deleted_at", time.Now()).
		Where(sq.Eq{"id": id}).
		Where(sq.Eq{"deleted_at": nil}).
		Returning("id", `"name"`, "created_at", "deleted_at")

	qSQL, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}

	resp := &models.ExampleDatabase{}
	err = db.Get(resp, qSQL, args...)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

type ExampleGetService struct {
	DB *sqlx.DB
}

func (c *ExampleGetService) Run(id uuid.UUID) (*models.ExampleDatabaseView, error) {
	return getExample(c.DB, id)
}

func getExample(db *sqlx.DB, id uuid.UUID) (*models.ExampleDatabaseView, error) {

	query := psql.Select("id", `"name"`, "created_at", "deleted_at").
		From("example_database").
		Where(sq.Eq{"id": id})

	qSQL, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}

	resp := &models.ExampleDatabaseView{}
	err = db.Get(resp, qSQL, args...)
	if err != nil {
		return nil, err
	}
	return resp, err
}

type ExampleListService struct {
	DB *sqlx.DB
}

func (c *ExampleListService) Run(filter models.ExampleDatabaseFilter) ([]models.ExampleDatabaseView, error) {
	return listExample(c.DB, filter)
}

func listExample(db *sqlx.DB, filter models.ExampleDatabaseFilter) ([]models.ExampleDatabaseView, error) {

	query := psql.Select("id", `"name"`, "created_at", "deleted_at").
		From("example_database").
		OrderBy("created_at DESC").
		Where(sq.Eq{"deleted_at": nil})

	if filter.Name != nil {
		name := "%" + *filter.Name + "%"
		query = query.Where(`"name" ILIKE ?`, name)
	}
	if filter.Limit != nil {
		query = query.Limit(*filter.Limit)
	}
	if filter.Offset != nil {
		query = query.Offset(*filter.Offset)
	}

	qSQL, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}

	resp := []models.ExampleDatabaseView{}
	err = db.Select(&resp, qSQL, args...)
	if err != nil {
		if err == sql.ErrNoRows {
			return resp, nil
		}
		return nil, err
	}
	return resp, err
}
