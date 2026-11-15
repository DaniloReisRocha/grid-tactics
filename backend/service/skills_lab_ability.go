package service

import (
	"github.com/DaniloReisRocha/grid-tactics/backend/internal/realtime/game"
	"github.com/DaniloReisRocha/grid-tactics/backend/models"
)

func ResolveSkillsLabAbility(input models.SkillsLabAbility) (models.SkillsLabAbilityResult, error) {
	if _, err := validateLabScene(input.Board, input.Tokens); err != nil {
		return models.SkillsLabAbilityResult{}, err
	}
	return game.ResolveAbilityEffects(input)
}
