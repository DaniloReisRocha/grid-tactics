package game

import "github.com/DaniloReisRocha/grid-tactics/backend/models"

func BlocksDirectAttack(cell models.Cell) bool {
	switch cell.Object {
	case models.ObjectHouse, models.ObjectRuin, models.ObjectBarricade:
		return true
	default:
		return false
	}
}

// A barricade can be the target, but no direct attack passes through it.
func DirectAttackLineClear(from, to models.Position, cells map[models.Position]models.Cell) bool {
	dx, dy := to.X-from.X, to.Y-from.Y
	if (dx == 0 && dy == 0) || (dx != 0 && dy != 0) {
		return false
	}
	sign := func(n int) int {
		if n < 0 {
			return -1
		}
		if n > 0 {
			return 1
		}
		return 0
	}
	for point := (models.Position{X: from.X + sign(dx), Y: from.Y + sign(dy)}); ; point = (models.Position{X: point.X + sign(dx), Y: point.Y + sign(dy)}) {
		cell, exists := cells[point]
		if !exists {
			return false
		}
		if point == to {
			return !BlocksDirectAttack(cell) || cell.Object == models.ObjectBarricade
		}
		if BlocksDirectAttack(cell) {
			return false
		}
	}
}
