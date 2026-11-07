package game

import (
	"testing"

	"github.com/DaniloReisRocha/grid-tactics/backend/models"
)

func TestMovementCostOnlyChargesClimbing(t *testing.T) {
	for _, test := range []struct{ from, to, want int }{
		{0, 0, 1}, {2, 2, 1}, {3, 3, 1},
		{0, 1, 2}, {1, 2, 2}, {2, 3, 2},
		{0, 2, 3}, {0, 3, 3}, {0, 8, 3},
		{3, 2, 1}, {3, 1, 1}, {3, 0, 1},
	} {
		got := movementCost(models.Cell{Height: test.from}, models.Cell{Height: test.to})
		if got != test.want {
			t.Errorf("height %d -> %d: got %d, want %d", test.from, test.to, got, test.want)
		}
	}
}

func TestPathCostUsesEveryPreviousStepAndCurrentStartingHeight(t *testing.T) {
	board := []models.Cell{}
	for x, height := range []int{0, 1, 1, 3, 3, 1, 2} {
		board = append(board, models.Cell{X: x, Y: 0, Terrain: "grass", Height: height, Elevation: models.ElevationLevel(height)})
	}
	path := findPath(board, models.Position{}, models.Position{X: 6}, nil)
	if len(path) != 6 || pathCost(board, models.Position{}, path) != 10 {
		t.Fatalf("wrong mixed-height path cost: %+v", path)
	}
	if cost := pathCost(board, models.Position{X: 3}, []models.Position{{X: 4}, {X: 5}}); cost != 2 {
		t.Fatalf("charged for plateau/descent after an elevated start: %d", cost)
	}
	if cost := pathCost(board, models.Position{X: 3}, nil); cost != 0 {
		t.Fatal("stationary unit consumed stamina")
	}
}

func TestServerMovementDeductsClimbingCostAndRejectsInsufficientStamina(t *testing.T) {
	s, a, b := startedSession(t)
	s.Board = nil
	for x, height := range []int{0, 1, 1, 3, 3, 1, 0} {
		s.Board = append(s.Board, models.Cell{X: x, Y: 0, Terrain: "grass", Height: height, Elevation: models.ElevationLevel(height)})
	}
	s.Players[a].Units[0].Position = models.Position{}
	s.Players[b].Units[0].Position = models.Position{X: 6}
	s.Players[a].Units[0].CurrentStamina = 5
	if err := s.MovePiece(a, "a", models.Position{X: 3}); err == nil {
		t.Fatal("accepted a route costing 6 with only 5 stamina")
	}
	if unit := s.Players[a].Units[0]; unit.X != 0 || unit.CurrentStamina != 5 {
		t.Fatal("failed movement changed the unit")
	}
	s.Players[a].Units[0].CurrentStamina = 8
	if err := s.MovePiece(a, "a", models.Position{X: 3}); err != nil {
		t.Fatal(err)
	}
	if s.Players[a].Units[0].CurrentStamina != 2 {
		t.Fatal("incorrect climbing charge")
	}
	if err := s.MovePiece(a, "a", models.Position{X: 4}); err != nil {
		t.Fatal(err)
	}
	if s.Players[a].Units[0].CurrentStamina != 1 {
		t.Fatal("plateau movement cost more than 1")
	}
	if err := s.MovePiece(a, "a", models.Position{X: 5}); err != nil {
		t.Fatal(err)
	}
	if s.Players[a].Units[0].CurrentStamina != 0 {
		t.Fatal("descent cost more than 1")
	}
}
