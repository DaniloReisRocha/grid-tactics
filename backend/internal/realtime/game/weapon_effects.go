package game

import (
	"errors"

	"github.com/DaniloReisRocha/grid-tactics/backend/models"
)

// Shared by multiplayer and the lab. The weapon's ID has no role in its effects.
func ValidateWeaponEffects(weapon models.Weapon, distance, attackerHeight, targetHeight int) error {
	if weapon.HasEffect(models.EffectNoAdjacentID) && distance <= 1 {
		return errors.New("esta arma não atinge unidades adjacentes")
	}
	if weapon.HasEffect(models.EffectSameLevelID) && attackerHeight != targetHeight {
		return errors.New("esta arma exige que atacante e alvo estejam no mesmo nível")
	}
	return nil
}

func ApplyWeaponDamage(weapon models.Weapon, life, shield int) (int, int) {
	damage := max(0, weapon.Damage)
	absorbed := min(shield, damage)
	shield -= absorbed
	life = max(0, life-(damage-absorbed))
	if weapon.HasEffect(models.EffectShieldBreakID) {
		shield = 0
	}
	return life, shield
}

func WeaponKnockback(weapon models.Weapon, attacker, target models.Position, cells map[models.Position]models.Cell, occupied map[models.Position]bool) []models.Position {
	if !weapon.HasEffect(models.EffectKnockbackID) {
		return nil
	}
	dx, dy := target.X-attacker.X, target.Y-attacker.Y
	if (dx == 0 && dy == 0) || (dx != 0 && dy != 0) {
		return nil
	}
	sign := func(value int) int {
		if value < 0 {
			return -1
		}
		if value > 0 {
			return 1
		}
		return 0
	}
	path := []models.Position{target}
	current := target
	for step := 0; step < models.KnockbackDistance; step++ {
		destination := models.Position{X: current.X + sign(dx), Y: current.Y + sign(dy)}
		from, sourceExists := cells[current]
		to, exists := cells[destination]
		if !sourceExists || !exists || occupied[destination] || to.Height > from.Height || !CanTraverse(from, to) {
			break
		}
		path = append(path, destination)
		current = destination
	}
	if len(path) == 1 {
		return nil
	}
	return path
}
