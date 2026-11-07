package game

import "github.com/DaniloReisRocha/grid-tactics/backend/models"

func isBridge(cell models.Cell) bool { return cell.Object == "bridgeX" || cell.Object == "bridgeY" }
func walkable(cell models.Cell) bool {
	return isBridge(cell) || (cell.Terrain != "water" && cell.Object == "")
}

// CanTraverse is shared by spawn connectivity and weighted pathfinding.
func CanTraverse(from, to models.Cell) bool {
	if !walkable(from) || !walkable(to) || abs(from.Height-to.Height) > 1 {
		return false
	}
	dx, dy := abs(from.X-to.X), abs(from.Y-to.Y)
	if max(dx, dy) != 1 {
		return false
	}
	for _, cell := range []models.Cell{from, to} {
		if cell.Object == "bridgeX" && dy != 0 {
			return false
		}
		if cell.Object == "bridgeY" && dx != 0 {
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
