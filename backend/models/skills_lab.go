package models

// The lab is a stateless simulation. Its tokens never enter multiplayer sessions.
type SkillToken struct {
	Position
	ID        string     `json:"id"`
	Team      int        `json:"team"`
	Model     int        `json:"model"`
	Life      int        `json:"life"`
	Shield    int        `json:"shield"`
	MaxLife   int        `json:"maxLife"`
	MaxShield int        `json:"maxShield"`
	WeaponID  string     `json:"weaponID"`
	Path      []Position `json:"path"`
}

type SkillsLabConfig struct {
	Board   []Cell   `json:"board"`
	Weapons []Weapon `json:"weapons"`
}

type SkillsLabAttack struct {
	Board      []Cell       `json:"board"`
	Tokens     []SkillToken `json:"tokens"`
	AttackerID string       `json:"attackerID"`
	TargetID   string       `json:"targetID"`
}

type SkillsLabResult struct {
	Tokens  []SkillToken `json:"tokens"`
	Message string       `json:"message"`
}
