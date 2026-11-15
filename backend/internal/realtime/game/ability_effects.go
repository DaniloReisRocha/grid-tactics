package game

import (
	"errors"
	"fmt"
	"maps"
	"slices"

	"github.com/DaniloReisRocha/grid-tactics/backend/models"
)

// ResolveAbilityEffects validates the entire action before modifying copies.
// Effect IDs determine the behavior; ability IDs identify equipment and usage.
func ResolveAbilityEffects(input models.SkillsLabAbility) (models.SkillsLabAbilityResult, error) {
	var result models.SkillsLabAbilityResult
	cells := make(map[models.Position]models.Cell, len(input.Board))
	for _, cell := range input.Board {
		cells[models.Position{X: cell.X, Y: cell.Y}] = cell
	}
	ability, exists := models.AbilityByID(input.AbilityID)
	actorIndex := slices.IndexFunc(input.Tokens, func(token models.SkillToken) bool { return token.ID == input.ActorID })
	if !exists || actorIndex < 0 {
		return result, errors.New("habilidade ou unidade inválida")
	}
	if ability.Passive {
		return result, errors.New("habilidades passivas são aplicadas automaticamente e não podem ser ativadas")
	}
	actor := input.Tokens[actorIndex]
	if actor.Life <= 0 || !slices.Contains(actor.AbilityIDs, ability.ID) {
		return result, errors.New("a unidade precisa estar viva e ter a habilidade equipada")
	}
	if ability.MaxUsesPerMatch > 0 && actor.AbilityUses[ability.ID] >= ability.MaxUsesPerMatch {
		return result, errors.New("esta unidade atingiu o limite de usos da habilidade por partida")
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
	healTargetIndex := -1
	if ability.HasEffect(models.EffectHealAllyID) {
		healTargetIndex = slices.IndexFunc(input.Tokens, func(token models.SkillToken) bool { return token.Position == input.Target && token.Life > 0 })
		if healTargetIndex < 0 || input.Tokens[healTargetIndex].Team != actor.Team {
			return result, errors.New("escolha uma unidade aliada viva ou a própria unidade")
		}
		if input.Tokens[healTargetIndex].Life >= input.Tokens[healTargetIndex].MaxLife {
			return result, errors.New("a unidade alvo já está com a vida máxima")
		}
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
	}
	if ability.HasEffect(models.EffectLeapID) {
		movement := ability.Movement
		if movement == nil || occupied(input.Target) || !IsWalkable(to) {
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
	if !ability.HasEffect(models.EffectHealAllyID) && !ability.HasEffect(models.EffectGrantShieldID) && !ability.HasEffect(models.EffectCreateBarricadeID) && !ability.HasEffect(models.EffectLeapID) && !ability.HasEffect(models.EffectAreaDamageOrShieldBreakID) {
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
	if healTargetIndex >= 0 {
		target := &result.Tokens[healTargetIndex]
		before := target.Life
		target.Life = min(target.MaxLife, target.Life+ability.Heal)
		result.Message = fmt.Sprintf("%s: %d ponto(s) de vida recuperado(s).", ability.Name, target.Life-before)
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
		destroyed := 0
		for i := range result.Board {
			cell := &result.Board[i]
			if cell.Object == models.ObjectBarricade && abs(cell.X-input.Target.X) <= radius && abs(cell.Y-input.Target.Y) <= radius {
				cell.Object = models.ObjectNone
				destroyed++
			}
		}
		if destroyed > 0 {
			result.Message += fmt.Sprintf(" %d barricada(s) destruída(s).", destroyed)
		}
	}
	unit.AbilityUses[ability.ID]++
	// Aggregate retained for snapshots; only the acting unit's counter enforces limits.
	result.MatchAbilityUses[ability.ID]++
	return result, nil
}
