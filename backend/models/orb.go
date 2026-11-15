package models

type Orb struct {
	Position
	ID          string `json:"id"`
	Triggered   bool   `json:"triggered"`
	CollectedBy string `json:"collectedBy,omitempty"`
}

func (unit *UnitInstance) RestoreOrbAbilities() {
	if unit.CurrentLife <= 0 {
		return
	}
	unit.AbilityUses = map[string]int{}
	unit.CurrentShield = max(unit.CurrentShield, unit.Unit.InitialShield())
}
