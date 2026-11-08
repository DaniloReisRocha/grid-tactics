package game

import "github.com/DaniloReisRocha/grid-tactics/backend/models"

func isBridge(cell models.Cell) bool {
	return cell.Object == models.ObjectBridgeX || cell.Object == models.ObjectBridgeY
}
func walkable(cell models.Cell) bool {
	if isBridge(cell) {
		return true
	}
	if cell.Terrain == models.TerrainWater {
		return false
	}
	switch cell.Object {
	case models.ObjectNone, models.ObjectTree, models.ObjectForest, models.ObjectRock, models.ObjectRockDirt:
		return true
	default:
		return false
	}
}

// Moving laterally or downhill costs 1. Only climbing adds up to 2 stamina.
func movementCost(from, to models.Cell) int {
	return 1 + min(2, max(0, to.Height-from.Height))
}

// CanTraverse is shared by spawn connectivity and weighted pathfinding.
func CanTraverse(from, to models.Cell) bool {
	if !walkable(from) || !walkable(to) {
		return false
	}
	dx, dy := abs(from.X-to.X), abs(from.Y-to.Y)
	if max(dx, dy) != 1 {
		return false
	}
	for _, cell := range []models.Cell{from, to} {
		if cell.Object == models.ObjectBridgeX && dy != 0 {
			return false
		}
		if cell.Object == models.ObjectBridgeY && dx != 0 {
			return false
		}
	}
	return true
}

func largestRegion(board []models.Cell) []models.Cell {
	cells := make(map[models.Position]models.Cell, len(board))
	key := func(c models.Cell) models.Position { return models.Position{X: c.X, Y: c.Y} }
	for _, cell := range board {
		cells[key(cell)] = cell
	}
	visited := map[models.Position]bool{}
	largest := []models.Cell{}
	for _, cell := range board {
		if visited[key(cell)] || !walkable(cell) {
			continue
		}
		region := []models.Cell{cell}
		visited[key(cell)] = true
		for index := 0; index < len(region); index++ {
			current := region[index]
			for _, d := range [][2]int{{1, 0}, {0, 1}, {-1, 0}, {0, -1}} {
				next, ok := cells[models.Position{X: current.X + d[0], Y: current.Y + d[1]}]
				if !ok || visited[key(next)] || !CanTraverse(current, next) {
					continue
				}
				visited[key(next)] = true
				region = append(region, next)
			}
		}
		if len(region) > len(largest) {
			largest = region
		}
	}
	return largest
}
