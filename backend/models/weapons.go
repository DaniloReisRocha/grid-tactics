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
	ID             string `json:"id"`
	Name           string `json:"name"`
	Icon           string `json:"icon"`
	Damage         int    `json:"damage"`
	Range          int    `json:"range"`
	LifePenalty    int    `json:"lifePenalty"`
	StaminaPenalty int    `json:"staminaPenalty"`
	ToolPenalty    int    `json:"toolPenalty"`
	Description    string `json:"description"`
}

// WeaponOptions returns independent copies of the available weapon definitions.
// Penalties are reduction amounts: 1 means -1, and 0 means no penalty.
// Special effects are documented here; combat rules implement their behavior.
func WeaponOptions() []Weapon {
	return []Weapon{
		{
			ID:             WeaponPistolID,
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
			Name:           "Rifle sniper",
			Icon:           "🎯",
			Damage:         3,
			Range:          5,
			StaminaPenalty: 2,
			ToolPenalty:    1,
			Description:    "Não pode causar dano a unidades adjacentes.",
		},
		{
			ID:             WeaponKatanaID,
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
