package game

import (
	"errors"
	"github.com/DaniloReisRocha/grid-tactics/backend/models"
	"github.com/gofrs/uuid"
	"sort"
)

func (s *GameSession) MovePiece(id uuid.UUID, unitID string, to models.Position) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	p, err := s.actingPlayer(id)
	if err != nil {
		return err
	}
	for i := range p.Units {
		unit := &p.Units[i]
		if unit.Unit.ID != unitID {
			continue
		}
		if unit.CurrentLife <= 0 {
			return errors.New("unidade eliminada")
		}
		blocked := s.occupied()
		delete(blocked, unit.Position)
		path := findPath(s.Board, unit.Position, to, blocked)
		if path == nil {
			return errors.New("destino inacessível")
		}
		cost := pathCost(s.Board, unit.Position, path)
		if cost > unit.CurrentStamina {
			return errors.New("stamina insuficiente")
		}
		unit.Path = append([]models.Position{unit.Position}, path...)
		unit.Position = to
		unit.CurrentStamina -= cost
		s.changed()
		return nil
	}
	return errors.New("unidade não pertence ao jogador")
}

// Dijkstra: movement cost depends on the upward height difference of each edge.
func findPath(board []models.Cell, from, to models.Position, blocked map[models.Position]bool) []models.Position {
	cells := make(map[models.Position]models.Cell, len(board))
	for _, cell := range board {
		cells[models.Position{X: cell.X, Y: cell.Y}] = cell
	}
	if cell, ok := cells[from]; !ok || !walkable(cell) {
		return nil
	}
	if cell, ok := cells[to]; !ok || !walkable(cell) || blocked[to] {
		return nil
	}
	type node struct {
		position models.Position
		cost     int
	}
	open := []node{{from, 0}}
	costs := map[models.Position]int{from: 0}
	parents := map[models.Position]models.Position{}
	directions := []models.Position{{X: 1}, {Y: 1}, {X: -1}, {Y: -1}, {X: 1, Y: 1}, {X: -1, Y: 1}, {X: 1, Y: -1}, {X: -1, Y: -1}}
	for len(open) > 0 {
		sort.SliceStable(open, func(i, j int) bool { return open[i].cost < open[j].cost })
		current := open[0]
		open = open[1:]
		if current.cost != costs[current.position] {
			continue
		}
		if current.position == to {
			path := []models.Position{}
			for step := to; step != from; step = parents[step] {
				path = append(path, step)
			}
			for i, j := 0, len(path)-1; i < j; i, j = i+1, j-1 {
				path[i], path[j] = path[j], path[i]
			}
			return path
		}
		for _, direction := range directions {
			next := models.Position{X: current.position.X + direction.X, Y: current.position.Y + direction.Y}
			cell, exists := cells[next]
			if !exists || blocked[next] || !CanTraverse(cells[current.position], cell) {
				continue
			}
			if direction.X != 0 && direction.Y != 0 {
				sideX := models.Position{X: next.X, Y: current.position.Y}
				sideY := models.Position{X: current.position.X, Y: next.Y}
				if side, ok := cells[sideX]; !ok || !CanTraverse(cells[current.position], side) {
					continue
				}
				if side, ok := cells[sideY]; !ok || !CanTraverse(cells[current.position], side) {
					continue
				}
				if blocked[sideX] || blocked[sideY] {
					continue
				}
			}
			cost := current.cost + movementCost(cells[current.position], cell)
			previous, seen := costs[next]
			if seen && previous <= cost {
				continue
			}
			costs[next] = cost
			parents[next] = current.position
			open = append(open, node{next, cost})
		}
	}
	return nil
}

func pathCost(board []models.Cell, from models.Position, path []models.Position) int {
	cells := make(map[models.Position]models.Cell, len(board))
	for _, cell := range board {
		cells[models.Position{X: cell.X, Y: cell.Y}] = cell
	}
	cost := 0
	for _, position := range path {
		cost += movementCost(cells[from], cells[position])
		from = position
	}
	return cost
}
