package models

const (
	AbilityShieldID     = "shield"
	AbilityBarricadeID  = "barricade"
	AbilityTrampolineID = "trampoline"
	AbilityGrenadeID    = "grenade"
)

type AbilityUsage string

const (
	AbilityUsageSingle    AbilityUsage = "single"
	AbilityUsageUnlimited AbilityUsage = "unlimited"
)

type AbilityMovement struct {
	ForwardDistance int `json:"forwardDistance"`
	ClimbDistance   int `json:"climbDistance"`
	ClimbLevels     int `json:"climbLevels"`
}

type Ability struct {
	ID                string           `json:"id"`
	Name              string           `json:"name"`
	Icon              string           `json:"icon"`
	Uses              AbilityUsage     `json:"uses"`
	Range             int              `json:"range"`
	Damage            int              `json:"damage"`
	Shield            int              `json:"shield"`
	AreaSize          int              `json:"areaSize"`
	MaxActivePerMatch int              `json:"maxActivePerMatch"`
	MaxUsesPerMatch   int              `json:"maxUsesPerMatch"`
	Movement          *AbilityMovement `json:"movement,omitempty"`
	Description       string           `json:"description"`
	Effect            []Effect         `json:"effect"`
}

func (ability Ability) HasEffect(id string) bool {
	for _, effect := range ability.Effect {
		if effect.ID == id {
			return true
		}
	}
	return false
}

// Usage and per-match limits are independent. Zero means no extra match limit;
// single-use abilities are still consumed once under their usage policy.
// The grenade retains the supplied unlimited policy with a one-use match limit.
func AbilityOptions() []Ability {
	grantShield, _ := EffectByID(EffectGrantShieldID)
	createBarricade, _ := EffectByID(EffectCreateBarricadeID)
	leap, _ := EffectByID(EffectLeapID)
	areaDamageOrShieldBreak, _ := EffectByID(EffectAreaDamageOrShieldBreakID)
	sameLevel, _ := EffectByID(EffectSameLevelID)
	return []Ability{
		{
			ID: AbilityShieldID, Name: "Escudo", Icon: "🛡️", Uses: AbilityUsageSingle,
			Shield:      UnitShieldCapacity,
			Description: "A unidade recebe um escudo equivalente a 3 pontos de vida.",
			Effect:      []Effect{grantShield},
		},
		{
			ID: AbilityBarricadeID, Name: "Barricada", Icon: "🚧", Uses: AbilityUsageSingle,
			Range: 1, MaxActivePerMatch: 3,
			Description: "A unidade pode conjurar uma barricada em um terreno adjacente de mesmo nível. É possível manter até 3 barricadas ativas por partida.",
			Effect:      []Effect{createBarricade, sameLevel},
		},
		{
			ID: AbilityTrampolineID, Name: "Trampolim", Icon: "↗️", Uses: AbilityUsageSingle,
			Range: 2, Movement: &AbilityMovement{ForwardDistance: 2, ClimbDistance: 1, ClimbLevels: 1},
			Description: "A unidade pode se deslocar 2 terrenos à frente ou subir 1 terreno de nível superior.",
			Effect:      []Effect{leap},
		},
		{
			ID: AbilityGrenadeID, Name: "Granada", Icon: "💣", Uses: AbilityUsageUnlimited,
			Range: 2, Damage: 2, AreaSize: 3, MaxUsesPerMatch: 1,
			Description: "Pode ser usada apenas uma vez por partida. Escolha uma área de 3 × 3 dentro do alcance máximo de 2 terrenos da unidade. Todas as unidades na área recebem 2 pontos de dano ou têm seus escudos quebrados.",
			Effect:      []Effect{areaDamageOrShieldBreak},
		},
	}
}

func AbilityByID(id string) (Ability, bool) {
	for _, ability := range AbilityOptions() {
		if ability.ID == id {
			return ability, true
		}
	}
	return Ability{}, false
}
