package service

import (
	"errors"
	"fmt"

	"github.com/DaniloReisRocha/grid-tactics/backend/internal/realtime/game"
	"github.com/DaniloReisRocha/grid-tactics/backend/models"
)

func SkillsLabConfig(seed, layout string) models.SkillsLabConfig {
	var board []models.Cell
	if layout == "procedural" {
		board = game.NewGameSession(seed, 2).Board
	} else {
		board = make([]models.Cell, 0, 225)
		for y := 0; y < 15; y++ {
			for x := 0; x < 15; x++ {
				height := 0
				if x >= 10 {
					height = 1
				}
				if x >= 12 {
					height = 2
				}
				terrain := models.TerrainGrass
				if x >= 3 && x <= 5 && y >= 11 && y <= 12 {
					terrain = models.TerrainWater
				}
				board = append(board, models.Cell{X: x, Y: y, Height: height, Elevation: models.ElevationLevel(height), Terrain: terrain})
			}
		}
	}
	return models.SkillsLabConfig{Board: board, Weapons: models.WeaponOptions(), Abilities: models.AbilityOptions(), BaseStamina: models.UnitBaseStamina}
}

// ResolveSkillsLabAttack works on a copy of caller-supplied tokens, without a
// session, turn allowance or stamina cost. Weapon data always comes from models.
func ResolveSkillsLabAttack(input models.SkillsLabAttack) (models.SkillsLabResult, error) {
	var result models.SkillsLabResult
	cells, err := validateLabScene(input.Board, input.Tokens)
	if err != nil {
		return result, err
	}
	attackerIndex, targetIndex := -1, -1
	occupied := map[models.Position]bool{}
	for i, token := range input.Tokens {
		if token.Life > 0 {
			occupied[token.Position] = true
		}
		if token.ID == input.AttackerID {
			attackerIndex = i
		}
		if token.ID == input.TargetID {
			targetIndex = i
		}
	}
	if attackerIndex < 0 || targetIndex < 0 {
		return result, errors.New("selecione um atacante e um alvo")
	}
	attacker, target := input.Tokens[attackerIndex], input.Tokens[targetIndex]
	if attacker.Life <= 0 || target.Life <= 0 || attacker.Team == target.Team {
		return result, errors.New("o alvo precisa ser um token vivo da outra equipe")
	}
	weapon, _ := models.WeaponByID(attacker.WeaponID)
	dx, dy := target.X-attacker.X, target.Y-attacker.Y
	abs := func(value int) int {
		if value < 0 {
			return -value
		}
		return value
	}
	if dx != 0 && dy != 0 {
		return result, errors.New("o alvo precisa estar na mesma linha ou coluna do atacante")
	}
	distance := abs(dx) + abs(dy)
	if distance > weapon.Range {
		return result, errors.New("alvo fora do alcance da arma")
	}
	from, to := cells[attacker.Position], cells[target.Position]
	if err := game.ValidateWeaponEffects(weapon, distance, from.Height, to.Height); err != nil {
		return result, err
	}
	result.Tokens = append([]models.SkillToken(nil), input.Tokens...)
	defender := &result.Tokens[targetIndex]
	oldLife, oldShield := defender.Life, defender.Shield
	defender.Life, defender.Shield = game.ApplyWeaponDamage(weapon, defender.Life, defender.Shield)
	result.Message = fmt.Sprintf("%s: −%d de vida, −%d de escudo.", weapon.Name, oldLife-defender.Life, oldShield-defender.Shield)
	if weapon.HasEffect(models.EffectKnockbackID) && defender.Life > 0 {
		if destination, ok := game.WeaponKnockback(weapon, attacker.Position, target.Position, cells, occupied); ok {
			defender.Position = destination
			defender.Path = []models.Position{target.Position, destination}
			defender.MovementMode = ""
			result.Message += " Alvo empurrado um bloco."
		} else {
			result.Message += " Empurrão bloqueado pelo terreno, borda ou outro token."
		}
	}
	if defender.Life == 0 {
		result.Message += " Token derrotado."
	}
	return result, nil
}
