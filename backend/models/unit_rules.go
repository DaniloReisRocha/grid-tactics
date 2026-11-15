package models

const (
	UnitBaseLife       = 3
	UnitShieldCapacity = 3
	UnitBaseStamina    = 5
	UnitItemSlots      = 1
	UnitAbilitySlots   = 2
)

// InitialShield is applied once at spawn, never on turns or reconnections.
func (unit Unit) InitialShield() int {
	shield := 0
	for _, ability := range unit.Abilities {
		if ability.Passive && ability.HasEffect(EffectGrantShieldID) {
			shield = max(shield, ability.Shield)
		}
	}
	return min(shield, unit.Shield, UnitShieldCapacity)
}

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
