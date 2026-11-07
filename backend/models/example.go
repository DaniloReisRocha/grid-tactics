package models

import (
	"strings"
	"time"

	"github.com/gofrs/uuid"
	"github.com/pkg/errors"
	"gopkg.in/guregu/null.v3"
)

type ExampleDatabase struct {
	ID        uuid.UUID `db:"id" json:"ID"`
	Name      string    `db:"name" json:"name"`
	CreatedAt time.Time `db:"created_at" json:"createdAt"`
	DeletedAt null.Time `db:"deleted_at" json:"deletedAt"`
}

type ExampleDatabaseForm struct {
	Name string `db:"name" json:"name"`
}

func (form ExampleDatabaseForm) Validate() error {
	if len(strings.TrimSpace(form.Name)) == 0 {
		return errors.New("[form.name]: required")
	}

	return nil
}

type ExampleDatabaseView struct {
	ExampleDatabase
}

type ExampleDatabaseFilter struct {
	Name   *string `form:"name"`
	Limit  *uint64 `form:"limit"`
	Offset *uint64 `form:"offset"`
}

// services
type ExampleCreate func(form ExampleDatabaseForm) (*ExampleDatabaseView, error)
type ExampleUpdate func(id uuid.UUID, form ExampleDatabaseForm) (*ExampleDatabaseView, error)
type ExampleDelete func(id uuid.UUID) (*ExampleDatabaseView, error)
type ExampleGet func(id uuid.UUID) (*ExampleDatabaseView, error)
type ExampleList func(filter ExampleDatabaseFilter) ([]ExampleDatabaseView, error)
