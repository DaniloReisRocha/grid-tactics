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
	return models.SkillsLabConfig{Board: board, Weapons: models.WeaponOptions()}
}

// ResolveSkillsLabAttack works on a copy of caller-supplied tokens, without a
// session, turn allowance or stamina cost. Weapon data always comes from models.
func ResolveSkillsLabAttack(input models.SkillsLabAttack) (models.SkillsLabResult, error) {
	var result models.SkillsLabResult
	if len(input.Board) == 0 || len(input.Board) > 1024 || len(input.Tokens) > 128 {
		return result, errors.New("cenário inválido ou grande demais")
	}
	cells := make(map[models.Position]models.Cell, len(input.Board))
	for _, cell := range input.Board {
		position := models.Position{X: cell.X, Y: cell.Y}
		if _, duplicate := cells[position]; duplicate || cell.X < 0 || cell.Y < 0 || cell.X >= 32 || cell.Y >= 32 || cell.Height < 0 || cell.Height > 3 {
			return result, errors.New("terreno inválido")
		}
		cells[position] = cell
	}
	attackerIndex, targetIndex := -1, -1
	ids := map[string]bool{}
	occupied := map[models.Position]bool{}
	for i, token := range input.Tokens {
		_, onBoard := cells[token.Position]
		_, validWeapon := models.WeaponByID(token.WeaponID)
		if token.ID == "" || ids[token.ID] || !onBoard || !validWeapon || token.Team < 0 || token.Team > 1 || token.Model < 0 || token.Model > 2 || token.Life < 0 || token.Life > 99 || token.Shield < 0 || token.Shield > 99 {
			return result, errors.New("token inválido")
		}
		ids[token.ID] = true
		if token.Life > 0 {
			if occupied[token.Position] {
				return result, errors.New("dois tokens vivos no mesmo slot")
			}
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
	if weapon.ID == models.WeaponSniperID && distance <= 1 {
		return result, errors.New("o rifle sniper não atinge unidades adjacentes")
	}
	from, to := cells[attacker.Position], cells[target.Position]
	if (weapon.ID == models.WeaponKatanaID || weapon.ID == models.WeaponWarHammerID) && from.Height != to.Height {
		return result, errors.New("esta arma exige que atacante e alvo estejam no mesmo nível")
	}
	result.Tokens = append([]models.SkillToken(nil), input.Tokens...)
	defender := &result.Tokens[targetIndex]
	oldLife, oldShield := defender.Life, defender.Shield
	absorbed := min(defender.Shield, weapon.Damage)
	defender.Shield -= absorbed
	defender.Life = max(0, defender.Life-(weapon.Damage-absorbed))
	if weapon.ID == models.WeaponShotgunID {
		defender.Shield = 0
	}
	result.Message = fmt.Sprintf("%s: −%d de vida, −%d de escudo.", weapon.Name, oldLife-defender.Life, oldShield-defender.Shield)
	if weapon.ID == models.WeaponWarHammerID && defender.Life > 0 {
		sign := func(value int) int {
			if value < 0 {
				return -1
			}
			if value > 0 {
				return 1
			}
			return 0
		}
		destination := models.Position{X: target.X + sign(dx), Y: target.Y + sign(dy)}
		next, exists := cells[destination]
		if exists && !occupied[destination] && next.Height <= to.Height && game.CanTraverse(to, next) {
			defender.Position = destination
			defender.Path = []models.Position{target.Position, destination}
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
