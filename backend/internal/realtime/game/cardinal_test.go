package game

import (
	"testing"

	"github.com/DaniloReisRocha/grid-tactics/backend/models"
)

func TestMovementUsesOnlyCardinalSteps(t *testing.T) {
	center := models.Cell{X: 2, Y: 2, Terrain: models.TerrainGrass}
	for dx := -1; dx <= 1; dx++ {
		for dy := -1; dy <= 1; dy++ {
			to := models.Cell{X: 2 + dx, Y: 2 + dy, Terrain: models.TerrainGrass}
			want := abs(dx)+abs(dy) == 1
			if CanTraverse(center, to) != want {
				t.Fatalf("incorrect traversal for offset %d,%d", dx, dy)
			}
		}
	}
	board := []models.Cell{}
	for y := 0; y < 5; y++ {
		for x := 0; x < 5; x++ {
			board = append(board, models.Cell{X: x, Y: y, Terrain: models.TerrainGrass})
		}
	}
	path := findPath(board, models.Position{}, models.Position{X: 2, Y: 2}, nil)
	if len(path) != 4 || pathCost(board, models.Position{}, path) != 4 {
		t.Fatalf("expected four steps, got %+v", path)
	}
	previous := models.Position{}
	for _, step := range path {
		if abs(step.X-previous.X)+abs(step.Y-previous.Y) != 1 {
			t.Fatalf("diagonal step: %+v -> %+v", previous, step)
		}
		previous = step
	}
	blocked := map[models.Position]bool{{X: 1}: true, {Y: 1}: true}
	if findPath(board, models.Position{}, models.Position{X: 1, Y: 1}, blocked) != nil {
		t.Fatal("escaped diagonally between blockers")
	}
}

func TestAttacksAllowAdjacentDiagonalsAndCardinalRange(t *testing.T) {
	for _, offset := range []models.Position{{X: 2}, {X: -2}, {Y: 2}, {Y: -2}, {X: 1, Y: 1}, {X: -1, Y: 1}, {X: 1, Y: -1}, {X: -1, Y: -1}, {X: 1, Y: 2}} {
		s, a, b := startedSession(t)
		attacker := &s.Players[a].Units[0]
		defender := &s.Players[b].Units[0]
		attacker.Position = models.Position{X: 4, Y: 4}
		attacker.CurrentStamina = 0
		defender.Position = models.Position{X: 4 + offset.X, Y: 4 + offset.Y}
		defender.CurrentLife, defender.CurrentShield = 30, 3
		err := s.Attack(a, "a", defender.Position)
		aligned := offset.X == 0 || offset.Y == 0 || max(abs(offset.X), abs(offset.Y)) == 1
		if (err == nil) != aligned {
			t.Fatalf("offset %+v: %v", offset, err)
		}
		if !aligned && (attacker.HasAttacked || defender.CurrentLife != 30 || defender.CurrentShield != 3) {
			t.Fatal("rejected diagonal attack changed state")
		}
	}
}
