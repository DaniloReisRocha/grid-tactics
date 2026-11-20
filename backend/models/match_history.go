package models

import (
	"encoding/json"
	"time"

	"github.com/gofrs/uuid"
)

type HistoryActionType string

const (
	HistoryCreated     HistoryActionType = "created"
	HistoryJoin        HistoryActionType = "join"
	HistoryReconnect   HistoryActionType = "reconnect"
	HistoryDisconnect  HistoryActionType = "disconnect"
	HistoryLeave       HistoryActionType = "leave"
	HistoryMove        HistoryActionType = "move"
	HistoryAttack      HistoryActionType = "attack"
	HistoryAbility     HistoryActionType = "ability"
	HistoryEndTurn     HistoryActionType = "endTurn"
	HistoryTimeout     HistoryActionType = "timeout"
	HistoryClose       HistoryActionType = "close"
	HistoryStateChange HistoryActionType = "stateChange"
)

// Only accepted commands are recorded. State contains the authoritative result,
// so playback does not need to rerun potentially changed game rules or RNG.
type HistoryAction struct {
	Type      HistoryActionType `json:"type"`
	PlayerID  *uuid.UUID        `json:"playerID,omitempty"`
	UnitID    string            `json:"unitID,omitempty"`
	AbilityID string            `json:"abilityID,omitempty"`
	Target    *Position         `json:"target,omitempty"`
}

type HistoryEvent struct {
	Sequence uint64    `json:"sequence"`
	At       time.Time `json:"at"`
	Revision uint64    `json:"revision"`
	HistoryAction
	State json.RawMessage `json:"state"`
}

type HistoryTurn struct {
	Number       int               `json:"number"`
	RoundSlot    int               `json:"roundSlot"`
	PlayerID     uuid.UUID         `json:"playerID"`
	UnitID       string            `json:"unitID"`
	StartedAt    time.Time         `json:"startedAt"`
	EndedAt      *time.Time        `json:"endedAt,omitempty"`
	EndReason    HistoryActionType `json:"endReason,omitempty"`
	InitialState json.RawMessage   `json:"initialState"`
	Events       []HistoryEvent    `json:"events"`
}

type HistoryRound struct {
	Number    int           `json:"number"`
	StartedAt time.Time     `json:"startedAt"`
	EndedAt   *time.Time    `json:"endedAt,omitempty"`
	Turns     []HistoryTurn `json:"turns"`
}

type MatchHistory struct {
	SchemaVersion  int             `json:"schemaVersion"`
	ID             string          `json:"id"`
	SessionID      string          `json:"sessionID"`
	Seed           string          `json:"seed"`
	MaximumPlayers int             `json:"maximumPlayers"`
	TurnDurationMS int64           `json:"turnDurationMS"`
	CreatedAt      time.Time       `json:"createdAt"`
	UpdatedAt      time.Time       `json:"updatedAt"`
	StartedAt      *time.Time      `json:"startedAt,omitempty"`
	FinishedAt     *time.Time      `json:"finishedAt,omitempty"`
	ClosedAt       *time.Time      `json:"closedAt,omitempty"`
	Status         GameStatus      `json:"status"`
	Winner         *uuid.UUID      `json:"winner"`
	InitialState   json.RawMessage `json:"initialState"`
	LatestState    json.RawMessage `json:"latestState"`
	LobbyEvents    []HistoryEvent  `json:"lobbyEvents"`
	Rounds         []HistoryRound  `json:"rounds"`
}
