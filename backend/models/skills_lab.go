package models

// The lab is a stateless simulation. Its tokens never enter multiplayer sessions.
type SkillToken struct {
	TurnCount int `json:"turnCount"`
	Position
	ID        string `json:"id"`
	Team      int    `json:"team"`
	Model     int    `json:"model"`
	Life      int    `json:"life"`
	Shield    int    `json:"shield"`
	MaxLife   int    `json:"maxLife"`
	MaxShield int    `json:"maxShield"`
	WeaponID  string `json:"weaponID"`
	// Spent movement is retained across weapon swaps; penalties change the cap.
	StaminaSpent int `json:"staminaSpent"`
	// Entries correspond to slots; an empty string represents an empty slot.
	AbilityIDs   []string       `json:"abilityIDs"`
	AbilityUses  map[string]int `json:"abilityUses"`
	MovementMode string         `json:"movementMode,omitempty"`
	Path         []Position     `json:"path"`
}

type SkillsLabConfig struct {
	Board       []Cell    `json:"board"`
	Weapons     []Weapon  `json:"weapons"`
	Abilities   []Ability `json:"abilities"`
	BaseStamina int       `json:"baseStamina"`
}

type SkillsLabAttack struct {
	Board      []Cell       `json:"board"`
	Tokens     []SkillToken `json:"tokens"`
	AttackerID string       `json:"attackerID"`
	TargetID   string       `json:"targetID"`
	Target     *Position    `json:"target,omitempty"`
}

type SkillsLabResult struct {
	Board   []Cell       `json:"board"`
	Tokens  []SkillToken `json:"tokens"`
	Message string       `json:"message"`
}

type SkillsLabAbility struct {
	Board            []Cell         `json:"board"`
	Tokens           []SkillToken   `json:"tokens"`
	ActorID          string         `json:"actorID"`
	AbilityID        string         `json:"abilityID"`
	Target           Position       `json:"target"`
	MatchAbilityUses map[string]int `json:"matchAbilityUses"`
}

type SkillsLabAbilityResult struct {
	Board            []Cell         `json:"board"`
	Tokens           []SkillToken   `json:"tokens"`
	MatchAbilityUses map[string]int `json:"matchAbilityUses"`
	Message          string         `json:"message"`
}
