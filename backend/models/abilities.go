package models

const (
	AbilityShieldID     = "shield"
	AbilityBarricadeID  = "barricade"
	AbilityTrampolineID = "trampoline"
	AbilityGrenadeID    = "grenade"
	AbilityMedkitID     = "medkit"
)

type AbilityMovement struct {
	ForwardDistance int `json:"forwardDistance"`
	ClimbDistance   int `json:"climbDistance"`
	ClimbLevels     int `json:"climbLevels"`
}

type Ability struct {
	Passive         bool             `json:"passive"`
	ID              string           `json:"id"`
	Name            string           `json:"name"`
	Icon            string           `json:"icon"`
	Range           int              `json:"range"`
	Damage          int              `json:"damage"`
	Heal            int              `json:"heal"`
	Shield          int              `json:"shield"`
	AreaSize        int              `json:"areaSize"`
	MaxUsesPerMatch int              `json:"maxUsesPerMatch"`
	Movement        *AbilityMovement `json:"movement,omitempty"`
	Description     string           `json:"description"`
	Effect          []Effect         `json:"effect"`
}

func (ability Ability) HasEffect(id string) bool {
	for _, effect := range ability.Effect {
		if effect.ID == id {
			return true
		}
	}
	return false
}

// MaxUsesPerMatch applies independently to each unit; zero means unlimited.
// Passive abilities cannot be activated and do not consume this allowance.
func AbilityOptions() []Ability {
	grantShield, _ := EffectByID(EffectGrantShieldID)
	healAlly, _ := EffectByID(EffectHealAllyID)
	createBarricade, _ := EffectByID(EffectCreateBarricadeID)
	leap, _ := EffectByID(EffectLeapID)
	areaDamageOrShieldBreak, _ := EffectByID(EffectAreaDamageOrShieldBreakID)
	sameLevel, _ := EffectByID(EffectSameLevelID)
	return []Ability{
		{
			ID: AbilityShieldID, Name: "Escudo", Icon: "🛡️",
			Shield: UnitShieldCapacity, Passive: true,
			Description: "Passiva: a unidade inicia a partida com 3 pontos de escudo. A proteção não se regenera entre turnos ou rounds.",
			Effect:      []Effect{grantShield},
		},
		{
			ID: AbilityBarricadeID, Name: "Barricada", Icon: "🚧",
			Range: 1, MaxUsesPerMatch: 3,
			Description: "Cada unidade equipada pode criar até 3 barricadas por partida, em terrenos adjacentes livres do mesmo nível. Barricadas destruídas não devolvem usos.",
			Effect:      []Effect{createBarricade, sameLevel},
		},
		{
			ID: AbilityTrampolineID, Name: "Trampolim", Icon: "↗️",
			Range: 2, MaxUsesPerMatch: 1, Movement: &AbilityMovement{ForwardDistance: 2, ClimbDistance: 1, ClimbLevels: 1},
			Description: "A unidade pode se deslocar 2 terrenos à frente ou subir 1 terreno de nível superior.",
			Effect:      []Effect{leap},
		},
		{
			ID: AbilityGrenadeID, Name: "Granada", Icon: "💣",
			Range: 2, Damage: 2, AreaSize: 3, MaxUsesPerMatch: 1,
			Description: "Cada unidade equipada pode usar uma granada por partida, independentemente das outras unidades. Escolha uma área de 3 × 3 dentro do alcance máximo de 2 terrenos da unidade. Todas as unidades na área recebem 2 pontos de dano ou têm seus escudos quebrados.",
			Effect:      []Effect{areaDamageOrShieldBreak},
		},
		{
			ID: AbilityMedkitID, Name: "Kit Médico", Icon: "🩹",
			Range: 1, Heal: 2, MaxUsesPerMatch: 1,
			Description: "Recupera até 2 pontos de vida de uma unidade aliada adjacente ou da própria unidade, sem ultrapassar a vida máxima. Um uso por unidade na partida, renovável por orb.",
			Effect:      []Effect{healAlly},
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
