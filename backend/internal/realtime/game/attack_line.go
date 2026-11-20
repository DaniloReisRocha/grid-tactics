package game

import (
	"github.com/DaniloReisRocha/grid-tactics/backend/models"
	"math"
)

// Aim halfway above each supporting surface. Check the entire portion of the
// sight line crossing a cube, including its near/far edges, rather than its center.
func terrainBlocksAttack(cell, source, target models.Cell, entry, exit float64) bool {
	start := float64(source.Height) + 0.5
	delta := float64(target.Height - source.Height)
	return float64(cell.Height)+1e-9 >= math.Min(start+delta*entry, start+delta*exit)
}

func BlocksDirectAttack(cell models.Cell) bool {
	switch cell.Object {
	case models.ObjectHouse, models.ObjectRuin, models.ObjectBarricade:
		return true
	default:
		return false
	}
}

// Adjacent diagonals count as distance one; longer attacks stay in a row/column.
func DirectAttackDistance(from, to models.Position) (int, bool) {
	dx, dy := abs(to.X-from.X), abs(to.Y-from.Y)
	if dx == 0 && dy == 0 {
		return 0, false
	}
	if max(dx, dy) == 1 {
		return 1, true
	}
	return dx + dy, dx == 0 || dy == 0
}

// A barricade can be the target, but no direct attack passes through it.
func DirectAttackLineClear(from, to models.Position, cells map[models.Position]models.Cell) bool {
	dx, dy := to.X-from.X, to.Y-from.Y
	distance, valid := DirectAttackDistance(from, to)
	if !valid {
		return false
	}
	source, sourceExists := cells[from]
	target, targetExists := cells[to]
	if !sourceExists || !targetExists {
		return false
	}
	if dx != 0 && dy != 0 {
		for _, point := range []models.Position{{X: from.X + dx, Y: from.Y}, {X: from.X, Y: from.Y + dy}} {
			cell, exists := cells[point]
			if !exists || BlocksDirectAttack(cell) || terrainBlocksAttack(cell, source, target, 0.5, 0.5) {
				return false
			}
		}
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
	step := 1
	for point := (models.Position{X: from.X + sign(dx), Y: from.Y + sign(dy)}); ; point = (models.Position{X: point.X + sign(dx), Y: point.Y + sign(dy)}) {
		cell, exists := cells[point]
		if !exists {
			return false
		}
		if point == to {
			return !BlocksDirectAttack(cell) || cell.Object == models.ObjectBarricade
		}
		entry, exit := (float64(step)-0.5)/float64(distance), (float64(step)+0.5)/float64(distance)
		if BlocksDirectAttack(cell) || terrainBlocksAttack(cell, source, target, entry, exit) {
			return false
		}
		step++
	}
}
