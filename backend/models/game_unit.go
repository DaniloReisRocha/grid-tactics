package models

import (
	"github.com/gofrs/uuid"
)

type Unit struct {
	ID      string   `json:"id"`
	LP      int      `json:"lp"`
	Icon    string   `json:"icon"`
	Shield  int      `json:"shield"`
	Stamina int      `json:"stamina"`
	Weapons []string `json:"weapons"`
	Tools   []string `json:"tools"`
}

// services
// type ExampleCreate func(form ExampleDatabaseForm) (*ExampleDatabaseView, error)
// type ExampleUpdate func(id uuid.UUID, form ExampleDatabaseForm) (*ExampleDatabaseView, error)
// type ExampleDelete func(id uuid.UUID) (*ExampleDatabaseView, error)
type GameUnitsGet func(accoID uuid.UUID) ([]Unit, error)

// type ExampleList func(filter ExampleDatabaseFilter) ([]ExampleDatabaseView, error)
