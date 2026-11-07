package game

import (
	"reflect"
	"testing"

	"github.com/DaniloReisRocha/grid-tactics/backend/models"
	"github.com/gofrs/uuid"
)

func testUnit(id string) models.Unit {
	return models.Unit{ID: id, LP: 3, Shield: 3, Stamina: 6,
		Weapons: []models.Weapon{{Name: "Pistola", Damage: 4, Range: 2, StaminaPenalty: 2}}}
}

func startedSession(t *testing.T) (*GameSession, uuid.UUID, uuid.UUID) {
	t.Helper()
	a := uuid.Must(uuid.NewV4())
	b := uuid.Must(uuid.NewV4())
	s := NewGameSession("test", 2)
	if err := s.AddPlayer(nil, a, []models.Unit{testUnit("a")}); err != nil {
		t.Fatal(err)
	}
	if err := s.AddPlayer(nil, b, []models.Unit{testUnit("b")}); err != nil {
		t.Fatal(err)
	}
	return s, a, b
}

func TestBoardDeterministicAndElevationBudget(t *testing.T) {
	for i := 0; i < 100; i++ {
		id := string(rune('a' + i))
		board := generateBoard(id, 0.3)
		if !reflect.DeepEqual(board, generateBoard(id, 0.3)) {
			t.Fatal("non-deterministic board")
		}
		count := 0
		for _, cell := range board {
			if cell.Elevation != 0 {
				count++
			}
		}
		if count != 30 {
			t.Fatalf("expected 30 elevated cells; got %d", count)
		}
	}
	for _, cell := range generateBoard("flat", 0) {
		if cell.Elevation != 0 {
			t.Fatal("zero elevation budget not respected")
		}
	}
}

func TestPathUsesCheapestTerrainAndBlocksCorners(t *testing.T) {
	board := []models.Cell{}
	for y := 0; y < 3; y++ {
		for x := 0; x < 5; x++ {
			elevation := models.ElevationLow
			if y == 1 && x > 0 && x < 4 {
				elevation = models.ElevationHigh
			}
			board = append(board, models.Cell{X: x, Y: y, Elevation: elevation})
		}
	}
	from, to := models.Position{X: 0, Y: 1}, models.Position{X: 4, Y: 1}
	path := findPath(board, from, to, nil)
	if len(path) == 0 || pathCost(board, path) != 4 {
		t.Fatalf("expected cheaper detour: %+v", path)
	}
	if same := findPath(board, from, from, nil); same == nil || len(same) != 0 {
		t.Fatal("same-cell path must be empty, not nil")
	}
	if findPath(board, from, models.Position{X: 99}, nil) != nil {
		t.Fatal("outside board accepted")
	}
	blocked := map[models.Position]bool{{X: 1, Y: 0}: true, {X: 0, Y: 1}: true}
	if findPath(board, models.Position{}, models.Position{X: 1, Y: 1}, blocked) != nil {
		t.Fatal("cut through occupied corner")
	}
}

func TestMovementOwnershipOccupancyStaminaAndTurn(t *testing.T) {
	s, a, b := startedSession(t)
	s.Players[a].Units[0].Position = models.Position{X: 0, Y: 0}
	s.Players[b].Units[0].Position = models.Position{X: 1, Y: 0}
	if err := s.MovePiece(b, "b", models.Position{X: 2}); err == nil {
		t.Fatal("opponent moved out of turn")
	}
	if err := s.MovePiece(a, "b", models.Position{X: 2}); err == nil {
		t.Fatal("moved opponent unit")
	}
	if err := s.MovePiece(a, "a", models.Position{X: 1}); err == nil {
		t.Fatal("moved to occupied tile")
	}
	s.Players[a].Units[0].CurrentStamina = 0
	if err := s.MovePiece(a, "a", models.Position{Y: 1}); err == nil {
		t.Fatal("moved without stamina")
	}
	s.Players[a].Units[0].CurrentStamina = 6
	cell := s.Board[10]
	if err := s.MovePiece(a, "a", models.Position{Y: 1}); err != nil {
		t.Fatal(err)
	}
	if s.Players[a].Units[0].CurrentStamina != 6-(1+int(cell.Elevation)) {
		t.Fatal("incorrect stamina cost")
	}
	s.Players[b].Units[0].CurrentStamina = 0
	if err := s.EndTurn(a); err != nil {
		t.Fatal(err)
	}
	if *s.CurrentTurn != b || s.TurnCounter != 2 || s.Players[b].Units[0].CurrentStamina != 6 {
		t.Fatal("turn did not rotate/replenish")
	}
}

func TestAttackShieldLifeVictoryAndFriendlyFire(t *testing.T) {
	s, a, b := startedSession(t)
	s.Players[a].Units[0].Position = models.Position{}
	s.Players[b].Units[0].Position = models.Position{X: 1}
	if err := s.Attack(a, "a", models.Position{}); err == nil {
		t.Fatal("friendly fire accepted")
	}
	if err := s.Attack(a, "a", models.Position{X: 1}); err != nil {
		t.Fatal(err)
	}
	defender := s.Players[b].Units[0]
	if defender.CurrentShield != 0 || defender.CurrentLife != 2 || s.Players[a].Units[0].CurrentStamina != 4 {
		t.Fatal("incorrect combat accounting")
	}
	if err := s.Attack(a, "a", models.Position{X: 1}); err != nil {
		t.Fatal(err)
	}
	if s.Status != "finished" || s.Winner == nil || *s.Winner != a || s.CurrentTurn != nil {
		t.Fatal("victory not recorded")
	}
	if err := s.EndTurn(a); err == nil {
		t.Fatal("action after match ended")
	}
}

func TestRejoinFullSessionKeepsRoster(t *testing.T) {
	s, a, _ := startedSession(t)
	s.Players[a].Units[0].CurrentStamina = 1
	if err := s.AddPlayer(nil, a, []models.Unit{testUnit("replacement")}); err != nil {
		t.Fatal(err)
	}
	if s.Players[a].Units[0].Unit.ID != "a" || s.Players[a].Units[0].CurrentStamina != 1 {
		t.Fatal("rejoin reset player")
	}
	if err := s.AddPlayer(nil, uuid.Must(uuid.NewV4()), []models.Unit{testUnit("third")}); err == nil {
		t.Fatal("third player accepted")
	}
}
