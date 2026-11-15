package models

const (
	KnockbackDistance               = 2
	EffectReloadID                  = "reloadTurn"
	EffectKnockbackID               = "knockback"
	EffectShieldBreakID             = "shieldBreak"
	EffectNoAdjacentID              = "noAdjacentTargets"
	EffectSameLevelID               = "sameLevelOnly"
	EffectGrantShieldID             = "grantShield"
	EffectHealAllyID                = "healAlly"
	EffectCreateBarricadeID         = "createBarricade"
	EffectLeapID                    = "leap"
	EffectAreaDamageOrShieldBreakID = "areaDamageOrShieldBreak"
)

// Negative describes a drawback to the wielder, rather than disabling the effect.
// Effects are identified independently of the weapon that grants them.
type Effect struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Negative    bool   `json:"negative"`
}

func EffectOptions() []Effect {
	return []Effect{
		{ID: EffectReloadID, Name: "Turno de recarga", Description: "Alterna um turno de recarga, sem ataques, com um turno disponível para atacar. O primeiro turno da unidade é de recarga.", Negative: true},
		{ID: EffectKnockbackID, Name: "Empurrão", Description: "Empurra o alvo vivo até dois blocos para trás, parando antes de bordas, subidas ou slots bloqueados."},
		{ID: EffectShieldBreakID, Name: "Quebra de escudo", Description: "Quebra todo o escudo restante do alvo após aplicar o dano do golpe."},
		{ID: EffectNoAdjacentID, Name: "Sem ataques adjacentes", Description: "Não pode atacar unidades adjacentes.", Negative: true},
		{ID: EffectSameLevelID, Name: "Mesmo nível", Description: "Exige que a origem e o terreno alvo estejam no mesmo nível.", Negative: true},
		{ID: EffectGrantShieldID, Name: "Conceder escudo", Description: "Concede à unidade a quantidade de escudo definida pela habilidade."},
		{ID: EffectHealAllyID, Name: "Curar aliado", Description: "Recupera a vida de uma unidade aliada viva, incluindo quem utiliza, sem ultrapassar sua vida máxima."},
		{ID: EffectCreateBarricadeID, Name: "Conjurar barricada", Description: "Cria uma barricada no terreno alvo, respeitando o limite de barricadas ativas por partida."},
		{ID: EffectLeapID, Name: "Salto", Description: "Desloca a unidade para frente ou para um nível superior, conforme as distâncias definidas pela habilidade."},
		{ID: EffectAreaDamageOrShieldBreakID, Name: "Dano em área ou quebra de escudo", Description: "Afeta todas as unidades da área, incluindo aliadas. Cada unidade com escudo tem o escudo quebrado; unidades sem escudo recebem o dano definido pela habilidade."},
	}
}

func EffectByID(id string) (Effect, bool) {
	for _, effect := range EffectOptions() {
		if effect.ID == id {
			return effect, true
		}
	}
	return Effect{}, false
}
