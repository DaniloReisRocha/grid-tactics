package models

import (
	"github.com/gofrs/uuid"
)

type ElevationLevel int

const (
	ElevationLow    ElevationLevel = iota // 0
	ElevationMedium                       // 1
	ElevationHigh                         // 2
)

type Cell struct {
	X         int            `json:"x"`
	Y         int            `json:"y"`
	Elevation ElevationLevel `json:"elevation"`
}

type Position struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type UnitInstance struct {
	Position
	Player         int       `json:"player"`
	OwnerID        uuid.UUID `json:"ownerID"`
	CurrentStamina int       `json:"currentStamina"`
	CurrentShield  int       `json:"currentShield"`
	CurrentLife    int       `json:"currentLife"`
	Unit           Unit      `json:"unit"`
}

type Weapon struct {
	Name           string `json:"name"`
	Icon           string `json:"icon"`
	Damage         int    `json:"damage"`
	Range          int    `json:"range"`
	LifePenalty    int    `json:"lifePenalty"`
	StaminaPenalty int    `json:"staminaPenalty"`
	ToolPenalty    int    `json:"toolPenalty"`
	Description    string `json:"description"`
}

type Player struct {
	AccoID    uuid.UUID      `json:"accoID"`
	Name      string         `json:"name"`
	Connected bool           `json:"connected"`
	Units     []UnitInstance `json:"units"`
}

type GameStatePayload struct {
	ID          string                `json:"id"`
	Players     map[uuid.UUID]*Player `json:"players"`
	Board       []Cell                `json:"board"`
	TurnOrder   []uuid.UUID           `json:"turnOrder"`
	CurrentTurn *uuid.UUID            `json:"currentTurn"`
	TurnCounter int                   `json:"turnCounter"`
	Status      string                `json:"status"`
	Winner      *uuid.UUID            `json:"winner"`
	Revision    uint64                `json:"revision"`
}
