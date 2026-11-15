package service

import (
	"errors"

	"github.com/DaniloReisRocha/grid-tactics/backend/models"
)

func validateLabScene(board []models.Cell, tokens []models.SkillToken) (map[models.Position]models.Cell, error) {
	if len(board) == 0 || len(board) > 1024 || len(tokens) > 128 {
		return nil, errors.New("cenário inválido ou grande demais")
	}
	cells := make(map[models.Position]models.Cell, len(board))
	for _, cell := range board {
		position := models.Position{X: cell.X, Y: cell.Y}
		if _, duplicate := cells[position]; duplicate || cell.X < 0 || cell.Y < 0 || cell.X >= 32 || cell.Y >= 32 || cell.Height < 0 || cell.Height > 3 {
			return nil, errors.New("terreno inválido")
		}
		cells[position] = cell
	}
	ids, occupied := map[string]bool{}, map[models.Position]bool{}
	for _, token := range tokens {
		if token.TurnCount < 0 || token.TurnCount > 1000000 {
			return nil, errors.New("contador de turnos inválido")
		}
		_, onBoard := cells[token.Position]
		_, validWeapon := models.WeaponByID(token.WeaponID)
		if token.ID == "" || ids[token.ID] || !onBoard || !validWeapon || token.Team < 0 || token.Team > 1 || token.Model < 0 || token.Model > 2 || token.Life < 0 || token.Life > 99 || token.Shield < 0 || token.Shield > 99 {
			return nil, errors.New("token inválido")
		}
		if token.StaminaSpent < 0 || token.StaminaSpent > models.UnitBaseStamina {
			return nil, errors.New("estamina gasta inválida")
		}
		if len(token.AbilityIDs) > models.UnitAbilitySlots {
			return nil, errors.New("cada token pode equipar até duas habilidades")
		}
		abilities := map[string]bool{}
		for _, id := range token.AbilityIDs {
			if id == "" {
				continue
			}
			if _, ok := models.AbilityByID(id); !ok || abilities[id] {
				return nil, errors.New("habilidade inválida ou repetida no token")
			}
			abilities[id] = true
		}
		for id, count := range token.AbilityUses {
			if _, ok := models.AbilityByID(id); !ok || count < 0 || count > 10000 {
				return nil, errors.New("contador de habilidades inválido")
			}
		}
		ids[token.ID] = true
		if token.Life > 0 {
			if occupied[token.Position] {
				return nil, errors.New("dois tokens vivos no mesmo slot")
			}
			occupied[token.Position] = true
		}
	}
	return cells, nil
}
