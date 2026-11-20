package game

import (
	"errors"
	"time"

	"github.com/DaniloReisRocha/grid-tactics/backend/models"
	"github.com/gofrs/uuid"
)

func (s *GameSession) UseAbility(id uuid.UUID, unitID, abilityID string, target models.Position) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	p, err := s.actingUnit(id, unitID)
	if err != nil {
		return err
	}
	var actor *models.UnitInstance
	for i := range p.Units {
		if p.Units[i].Unit.ID == unitID {
			actor = &p.Units[i]
			break
		}
	}
	if actor == nil || actor.CurrentLife <= 0 {
		return errors.New("unidade inválida ou eliminada")
	}
	ability, _ := models.AbilityByID(abilityID)
	if ability.HasEffect(models.EffectCreateBarricadeID) && s.activeOrbAt(target) {
		return errors.New("o terreno está ocupado por uma orb")
	}
	from := actor.Position
	input := models.SkillsLabAbility{Board: s.Board, ActorID: unitID, AbilityID: abilityID, Target: target, MatchAbilityUses: s.MatchAbilityUses}
	for _, player := range s.Players {
		for _, unit := range player.Units {
			token := models.SkillToken{ID: unit.Unit.ID, Position: unit.Position, Team: unit.Player - 1, Life: unit.CurrentLife, Shield: unit.CurrentShield, MaxLife: unit.Unit.LP, MaxShield: unit.Unit.Shield, AbilityUses: unit.AbilityUses, Path: unit.Path}
			for _, ability := range unit.Unit.Abilities {
				token.AbilityIDs = append(token.AbilityIDs, ability.ID)
			}
			input.Tokens = append(input.Tokens, token)
		}
	}
	result, err := ResolveAbilityEffects(input)
	if err != nil {
		return err
	}
	byID := make(map[string]models.SkillToken, len(result.Tokens))
	for _, token := range result.Tokens {
		byID[token.ID] = token
	}
	for _, player := range s.Players {
		for i := range player.Units {
			unit := &player.Units[i]
			token := byID[unit.Unit.ID]
			unit.Position = token.Position
			unit.CurrentLife = token.Life
			unit.CurrentShield = token.Shield
			unit.AbilityUses = token.AbilityUses
			if token.MovementMode == "jump" {
				unit.Path = token.Path
				unit.MovementMode = token.MovementMode
			}
		}
	}
	s.Board = result.Board
	s.MatchAbilityUses = result.MatchAbilityUses
	if ability.HasEffect(models.EffectAreaDamageOrShieldBreakID) {
		s.GrenadeEvent = &models.GrenadeEvent{ID: s.Revision + 1, Position: target, CreatedAtMS: time.Now().UnixMilli()}
	}
	if actor.Position != from {
		s.collectOrbs(actor, []models.Position{actor.Position})
	}
	aliveTeams := []uuid.UUID{}
	for owner, player := range s.Players {
		for _, unit := range player.Units {
			if unit.CurrentLife > 0 {
				aliveTeams = append(aliveTeams, owner)
				break
			}
		}
	}
	if len(aliveTeams) < 2 {
		s.Status = models.GameStatusFinished
		s.CurrentTurn = nil
		s.CurrentUnitID = ""
		if len(aliveTeams) == 1 {
			winner := aliveTeams[0]
			s.Winner = &winner
		} else {
			s.Winner = nil
		}
	}
	if s.Status == models.GameStatusPlaying && actor.CurrentLife <= 0 {
		s.advanceUnitTurn()
	}
	s.spawnMilestoneOrbs()
	s.changed()
	return nil
}
