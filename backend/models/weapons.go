package models

// Stable identifiers serialized in each weapon's JSON, independent of its name.
const (
	WeaponPistolID    = "pistol"
	WeaponShotgunID   = "shotgun"
	WeaponSniperID    = "sniper"
	WeaponKatanaID    = "katana"
	WeaponWarHammerID = "warHammer"
)

type Weapon struct {
	ID             string   `json:"id"`
	Name           string   `json:"name"`
	Icon           string   `json:"icon"`
	Damage         int      `json:"damage"`
	Range          int      `json:"range"`
	LifePenalty    int      `json:"lifePenalty"`
	StaminaPenalty int      `json:"staminaPenalty"`
	ToolPenalty    int      `json:"toolPenalty"`
	Description    string   `json:"description"`
	Effect         []Effect `json:"effect"`
}

func (weapon Weapon) HasEffect(id string) bool {
	for _, effect := range weapon.Effect {
		if effect.ID == id {
			return true
		}
	}
	return false
}

// WeaponOptions returns independent copies of the available weapon definitions.
// Penalties are reduction amounts: 1 means -1, and 0 means no penalty.
// Each weapon owns its effect slice; catalog lookups never share mutable slices.
func WeaponOptions() []Weapon {
	knockback, _ := EffectByID(EffectKnockbackID)
	shieldBreak, _ := EffectByID(EffectShieldBreakID)
	noAdjacent, _ := EffectByID(EffectNoAdjacentID)
	sameLevel, _ := EffectByID(EffectSameLevelID)
	return []Weapon{
		{
			ID:             WeaponPistolID,
			Effect:         []Effect{},
			Name:           "Pistola",
			Icon:           "🔫",
			Damage:         1,
			Range:          2,
			StaminaPenalty: 0,
			ToolPenalty:    0,
			Description:    "Item base do jogo.",
		},
		{
			ID:             WeaponShotgunID,
			Effect:         []Effect{shieldBreak},
			Name:           "Escopeta",
			Icon:           "🏹",
			Damage:         2,
			Range:          2,
			StaminaPenalty: 1,
			ToolPenalty:    1,
			Description:    "Pode quebrar o escudo do oponente com um golpe.",
		},
		{
			ID:             WeaponSniperID,
			Effect:         []Effect{noAdjacent},
			Name:           "Rifle sniper",
			Icon:           "🎯",
			Damage:         3,
			Range:          5,
			StaminaPenalty: 1,
			ToolPenalty:    1,
			Description:    "Não pode causar dano a unidades adjacentes.",
		},
		{
			ID:             WeaponKatanaID,
			Effect:         []Effect{sameLevel},
			Name:           "Katana",
			Icon:           "🗡️",
			Damage:         3,
			Range:          1,
			StaminaPenalty: 0,
			ToolPenalty:    0,
			Description:    "Só pode golpear unidades em terrenos do mesmo nível.",
		},
		{
			ID:             WeaponWarHammerID,
			Effect:         []Effect{knockback, sameLevel},
			Name:           "Martelo de Guerra",
			Icon:           "🔨",
			Damage:         3,
			Range:          1,
			StaminaPenalty: 2,
			ToolPenalty:    1,
			Description:    "Empurra a unidade atingida um quadrado para trás. Só pode golpear unidades em terrenos do mesmo nível. Não é possível empurrar uma unidade para fora do mapa ou para um terreno mais alto.",
		},
	}
}

func WeaponByID(id string) (Weapon, bool) {
	for _, weapon := range WeaponOptions() {
		if weapon.ID == id {
			return weapon, true
		}
	}
	return Weapon{}, false
}
