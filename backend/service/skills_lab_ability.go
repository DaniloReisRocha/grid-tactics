package service

import (
	"errors"
	"fmt"
	"maps"
	"slices"

	"github.com/DaniloReisRocha/grid-tactics/backend/internal/realtime/game"
	"github.com/DaniloReisRocha/grid-tactics/backend/models"
)

// ResolveSkillsLabAbility validates the entire action before modifying copies.
// Effect IDs determine the behavior; ability IDs identify equipment and usage.
func ResolveSkillsLabAbility(input models.SkillsLabAbility) (models.SkillsLabAbilityResult, error) {
	var result models.SkillsLabAbilityResult
	// Reuse scene validation from the attack endpoint without executing an attack.
	cells, err := validateLabScene(input.Board, input.Tokens)
	if err != nil {
		return result, err
	}
	ability, exists := models.AbilityByID(input.AbilityID)
	actorIndex := slices.IndexFunc(input.Tokens, func(token models.SkillToken) bool { return token.ID == input.ActorID })
	if !exists || actorIndex < 0 {
		return result, errors.New("habilidade ou unidade inválida")
	}
	actor := input.Tokens[actorIndex]
	if actor.Life <= 0 || !slices.Contains(actor.AbilityIDs, ability.ID) {
		return result, errors.New("a unidade precisa estar viva e ter a habilidade equipada")
	}
	if ability.Uses == models.AbilityUsageSingle && actor.AbilityUses[ability.ID] > 0 {
		return result, errors.New("esta unidade já utilizou a habilidade")
	}
	matchCount := input.MatchAbilityUses[ability.ID]
	tokenUses := 0
	for _, token := range input.Tokens {
		tokenUses += token.AbilityUses[ability.ID]
	}
	matchCount = max(matchCount, tokenUses)
	if ability.MaxUsesPerMatch > 0 && matchCount >= ability.MaxUsesPerMatch {
		return result, errors.New("limite de uso da habilidade por partida atingido")
	}
	for id, count := range input.MatchAbilityUses {
		if _, ok := models.AbilityByID(id); !ok || count < 0 || count > 10000 {
			return result, errors.New("contador de habilidades inválido")
		}
	}
	from := cells[actor.Position]
	to, onBoard := cells[input.Target]
	if !onBoard {
		return result, errors.New("terreno alvo fora do mapa")
	}
	abs := func(n int) int {
		if n < 0 {
			return -n
		}
		return n
	}
	dx, dy := input.Target.X-actor.X, input.Target.Y-actor.Y
	distance := abs(dx) + abs(dy)
	if dx != 0 && dy != 0 {
		return result, errors.New("escolha um alvo na mesma linha ou coluna")
	}
	if distance > ability.Range {
		return result, errors.New("alvo fora do alcance da habilidade")
	}
	if ability.HasEffect(models.EffectSameLevelID) && from.Height != to.Height {
		return result, errors.New("a habilidade exige um terreno do mesmo nível")
	}
	occupied := func(position models.Position) bool {
		return slices.ContainsFunc(input.Tokens, func(token models.SkillToken) bool { return token.Life > 0 && token.Position == position })
	}
	if ability.HasEffect(models.EffectGrantShieldID) {
		if input.Target != actor.Position {
			return result, errors.New("o escudo só pode ser aplicado à própria unidade")
		}
		if actor.Shield >= ability.Shield {
			return result, errors.New("a unidade já possui o escudo máximo da habilidade")
		}
	}
	if ability.HasEffect(models.EffectCreateBarricadeID) {
		if distance != 1 || to.Terrain == models.TerrainWater || to.Object != models.ObjectNone || occupied(input.Target) {
			return result, errors.New("a barricada exige um terreno adjacente livre, seco e do mesmo nível")
		}
		active := 0
		for _, cell := range input.Board {
			if cell.Object == models.ObjectBarricade {
				active++
			}
		}
		if ability.MaxActivePerMatch > 0 && active >= ability.MaxActivePerMatch {
			return result, errors.New("limite de três barricadas ativas atingido")
		}
	}
	if ability.HasEffect(models.EffectLeapID) {
		movement := ability.Movement
		if movement == nil || occupied(input.Target) || !game.IsWalkable(to) {
			return result, errors.New("o salto exige um destino acessível e livre")
		}
		flat := distance == movement.ForwardDistance && to.Height == from.Height
		uphill := distance == movement.ClimbDistance && to.Height-from.Height == movement.ClimbLevels
		if !flat && !uphill {
			return result, errors.New("salte dois terrenos no mesmo nível ou suba um terreno adjacente de um nível acima")
		}
		if flat {
			middle := models.Position{X: actor.X + dx/2, Y: actor.Y + dy/2}
			if _, ok := cells[middle]; !ok {
				return result, errors.New("o salto não pode atravessar uma lacuna fora do mapa")
			}
		}
	}
	if !ability.HasEffect(models.EffectGrantShieldID) && !ability.HasEffect(models.EffectCreateBarricadeID) && !ability.HasEffect(models.EffectLeapID) && !ability.HasEffect(models.EffectAreaDamageOrShieldBreakID) {
		return result, errors.New("habilidade sem efeito executável")
	}
	result.Board = slices.Clone(input.Board)
	result.Tokens = slices.Clone(input.Tokens)
	result.MatchAbilityUses = maps.Clone(input.MatchAbilityUses)
	if result.MatchAbilityUses == nil {
		result.MatchAbilityUses = map[string]int{}
	}
	unit := &result.Tokens[actorIndex]
	unit.AbilityUses = maps.Clone(actor.AbilityUses)
	if unit.AbilityUses == nil {
		unit.AbilityUses = map[string]int{}
	}
	if ability.HasEffect(models.EffectGrantShieldID) {
		unit.Shield = min(ability.Shield, models.UnitShieldCapacity)
		unit.MaxShield = max(unit.MaxShield, unit.Shield)
		result.Message = fmt.Sprintf("Escudo ativado: %d pontos de proteção.", unit.Shield)
	}
	if ability.HasEffect(models.EffectCreateBarricadeID) {
		for i := range result.Board {
			if result.Board[i].X == input.Target.X && result.Board[i].Y == input.Target.Y {
				result.Board[i].Object = models.ObjectBarricade
			}
		}
		result.Message = "Barricada conjurada. O terreno agora bloqueia a passagem."
	}
	if ability.HasEffect(models.EffectLeapID) {
		unit.Position = input.Target
		unit.Path = []models.Position{actor.Position, input.Target}
		unit.MovementMode = "jump"
		result.Message = "Trampolim ativado. Salto realizado sem gasto de estamina."
	}
	if ability.HasEffect(models.EffectAreaDamageOrShieldBreakID) {
		radius, hits := ability.AreaSize/2, 0
		for i := range result.Tokens {
			target := &result.Tokens[i]
			if target.Life <= 0 || abs(target.X-input.Target.X) > radius || abs(target.Y-input.Target.Y) > radius {
				continue
			}
			if target.Shield > 0 {
				target.Shield = 0
			} else {
				target.Life = max(0, target.Life-ability.Damage)
			}
			hits++
		}
		result.Message = fmt.Sprintf("Granada detonada: %d unidade(s) atingida(s), incluindo aliadas na área.", hits)
	}
	unit.AbilityUses[ability.ID]++
	result.MatchAbilityUses[ability.ID] = matchCount + 1
	return result, nil
}
