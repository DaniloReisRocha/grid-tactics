package game

import (
	"fmt"
	"sort"

	"github.com/DaniloReisRocha/grid-tactics/backend/models"
)

// Called under the session lock, after resolving all damage from one action.
func (s *GameSession) spawnMilestoneOrbs() {
	if s.Status != "playing" || len(s.Orbs) >= 2 {
		return
	}
	total, alive := 0, 0
	cells := map[models.Position]models.Cell{}
	for _, cell := range s.Board {
		cells[models.Position{X: cell.X, Y: cell.Y}] = cell
	}
	reachable := map[models.Position]bool{}
	frontier := []models.Position{}
	used := map[models.Position]bool{}
	for _, player := range s.Players {
		for _, unit := range player.Units {
			total++
			if unit.CurrentLife > 0 {
				alive++
				reachable[unit.Position] = true
				frontier = append(frontier, unit.Position)
			}
			// Avoid spawning under living units or skulls.
			used[unit.Position] = true
		}
	}
	if total != 6 || alive > 4 {
		return
	}
	for index := 0; index < len(frontier); index++ {
		from := frontier[index]
		for _, delta := range []models.Position{{X: 1}, {X: -1}, {Y: 1}, {Y: -1}} {
			to := models.Position{X: from.X + delta.X, Y: from.Y + delta.Y}
			cell, exists := cells[to]
			if exists && !reachable[to] && CanTraverse(cells[from], cell) {
				reachable[to] = true
				frontier = append(frontier, to)
			}
		}
	}
	spawned := map[string]bool{}
	for _, orb := range s.Orbs {
		used[orb.Position] = true
		spawned[orb.ID] = true
	}
	for _, threshold := range []int{4, 2} {
		id := fmt.Sprintf("orb-%d", threshold)
		if alive > threshold || spawned[id] {
			continue
		}
		candidates := []models.Cell{}
		for _, cell := range s.Board {
			position := models.Position{X: cell.X, Y: cell.Y}
			if cell.Terrain != models.TerrainWater && cell.Object == models.ObjectNone && walkable(cell) && reachable[position] && !used[position] {
				candidates = append(candidates, cell)
			}
		}
		// Stable seed ranking avoids depending on Go's player-map iteration order.
		seed := s.ID + ":" + id
		sort.Slice(candidates, func(i, j int) bool {
			a, b := candidates[i], candidates[j]
			x, y := seedHash(fmt.Sprintf("%s:%d:%d", seed, a.X, a.Y)), seedHash(fmt.Sprintf("%s:%d:%d", seed, b.X, b.Y))
			if x != y {
				return x < y
			}
			if a.Y != b.Y {
				return a.Y < b.Y
			}
			return a.X < b.X
		})
		if len(candidates) == 0 {
			continue
		} // Retry on subsequent actions if every slot is occupied.
		position := models.Position{X: candidates[0].X, Y: candidates[0].Y}
		s.Orbs = append(s.Orbs, models.Orb{ID: id, Position: position})
		used[position] = true
	}
}

func (s *GameSession) collectOrbs(unit *models.UnitInstance, path []models.Position) {
	if unit.CurrentLife <= 0 {
		return
	}
	for _, position := range path {
		for i := range s.Orbs {
			orb := &s.Orbs[i]
			if !orb.Triggered && orb.Position == position {
				orb.Triggered = true
				orb.CollectedBy = unit.Unit.ID
				unit.RestoreOrbAbilities()
			}
		}
	}
}

func (s *GameSession) activeOrbAt(position models.Position) bool {
	for _, orb := range s.Orbs {
		if !orb.Triggered && orb.Position == position {
			return true
		}
	}
	return false
}
