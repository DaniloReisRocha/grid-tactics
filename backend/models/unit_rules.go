package models

const (
	UnitBaseLife       = 3
	UnitShieldCapacity = 3
	UnitBaseStamina    = 4
	UnitItemSlots      = 1
	UnitAbilitySlots   = 2
)

// Penalties reduce the movement budget, rather than charge an attack cost.
func (unit Unit) EffectiveStamina() int {
	penalty := 0
	for _, weapon := range unit.Weapons {
		penalty += max(0, weapon.StaminaPenalty)
	}
	return max(0, unit.Stamina-penalty)
}

func (unit Unit) EffectiveAbilitySlots() int {
	penalty := 0
	for _, weapon := range unit.Weapons {
		penalty += max(0, weapon.ToolPenalty)
	}
	return max(0, unit.AbilitySlots-penalty)
}
