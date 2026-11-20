package models

// Retained in snapshots so both clients can identify the same detonation.
type GrenadeEvent struct {
	Position
	ID          uint64 `json:"id"`
	CreatedAtMS int64  `json:"createdAtMS"`
}
