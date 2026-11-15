package models

const TeamUnitCount = 3

// The client supplies catalog IDs only. Resources and effects come from the server.
type UnitLoadout struct {
	WeaponID   string   `json:"weaponID"`
	AbilityIDs []string `json:"abilityIDs"`
}

type TeamOptions struct {
	Weapons        []Weapon  `json:"weapons"`
	Abilities      []Ability `json:"abilities"`
	UnitCount      int       `json:"unitCount"`
	BaseLife       int       `json:"baseLife"`
	ShieldCapacity int       `json:"shieldCapacity"`
	BaseStamina    int       `json:"baseStamina"`
	ItemSlots      int       `json:"itemSlots"`
	AbilitySlots   int       `json:"abilitySlots"`
}
