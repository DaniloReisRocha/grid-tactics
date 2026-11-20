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
	Terrain   string         `json:"terrain"`
	Height    int            `json:"height"`
	Object    string         `json:"object,omitempty"`
	Roof      string         `json:"roof,omitempty"`
	Waterfall bool           `json:"waterfall,omitempty"`
}

type Position struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type UnitInstance struct {
	TurnCount int `json:"turnCount"`
	Position
	Player         int            `json:"player"`
	OwnerID        uuid.UUID      `json:"ownerID"`
	CurrentStamina int            `json:"currentStamina"`
	CurrentShield  int            `json:"currentShield"`
	CurrentLife    int            `json:"currentLife"`
	HasAttacked    bool           `json:"hasAttacked"`
	AbilityUses    map[string]int `json:"abilityUses"`
	MovementMode   string         `json:"movementMode,omitempty"`
	Unit           Unit           `json:"unit"`
	Path           []Position     `json:"path"`
}

type Player struct {
	AccoID    uuid.UUID      `json:"accoID"`
	Name      string         `json:"name"`
	Connected bool           `json:"connected"`
	Units     []UnitInstance `json:"units"`
}

type UnitTurn struct {
	PlayerID uuid.UUID `json:"playerID"`
	UnitID   string    `json:"unitID"`
}

type GameStatePayload struct {
	GrenadeEvent     *GrenadeEvent         `json:"grenadeEvent"`
	ServerTimeMS     int64                 `json:"serverTimeMS"`
	TurnDeadlineMS   int64                 `json:"turnDeadlineMS"`
	TurnRemainingMS  int64                 `json:"turnRemainingMS"`
	Orbs             []Orb                 `json:"orbs"`
	RoundOrder       []UnitTurn            `json:"roundOrder"`
	RoundCounter     int                   `json:"roundCounter"`
	RoundTurn        int                   `json:"roundTurn"`
	CurrentUnitID    string                `json:"currentUnitID"`
	ID               string                `json:"id"`
	Players          map[uuid.UUID]*Player `json:"players"`
	Board            []Cell                `json:"board"`
	TurnOrder        []uuid.UUID           `json:"turnOrder"`
	CurrentTurn      *uuid.UUID            `json:"currentTurn"`
	TurnCounter      int                   `json:"turnCounter"`
	Status           GameStatus            `json:"status"`
	Winner           *uuid.UUID            `json:"winner"`
	Revision         uint64                `json:"revision"`
	MatchAbilityUses map[string]int        `json:"matchAbilityUses"`
}
