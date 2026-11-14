package models

const (
	EffectKnockbackID   = "knockback"
	EffectShieldBreakID = "shieldBreak"
	EffectNoAdjacentID  = "noAdjacentTargets"
	EffectSameLevelID   = "sameLevelOnly"
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
		{ID: EffectKnockbackID, Name: "Empurrão", Description: "Empurra o alvo vivo um bloco para trás, sem sair do mapa, subir ou ocupar um slot bloqueado."},
		{ID: EffectShieldBreakID, Name: "Quebra de escudo", Description: "Quebra todo o escudo restante do alvo após aplicar o dano do golpe."},
		{ID: EffectNoAdjacentID, Name: "Sem ataques adjacentes", Description: "Não pode atacar unidades adjacentes.", Negative: true},
		{ID: EffectSameLevelID, Name: "Mesmo nível", Description: "Só pode atacar unidades em terrenos do mesmo nível.", Negative: true},
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
