package game

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/DaniloReisRocha/grid-tactics/backend/models"
	"github.com/gofrs/uuid"
)

func flatSpawnBoard(width, height int) []models.Cell {
	board := []models.Cell{}
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			board = append(board, models.Cell{X: x, Y: y, Terrain: "grass"})
		}
	}
	return board
}

func TestSpawnsKeepFiveTileDistanceAndTwoFreeExits(t *testing.T) {
	accounts := []uuid.UUID{
		uuid.Must(uuid.FromString("98cad047-c065-4011-bce0-bf7558bb54b9")),
		uuid.Must(uuid.FromString("fcfc0336-c317-41e7-b7d5-6b2377938bd3")),
	}
	for seed := 0; seed < 500; seed++ {
		s := NewGameSession(fmt.Sprintf("spawn-check-%d", seed), 2)
		for _, id := range accounts {
			units := []models.Unit{testUnit(id.String() + "1"), testUnit(id.String() + "2"), testUnit(id.String() + "3")}
			if err := s.AddPlayer(nil, id, units); err != nil {
				t.Fatalf("seed %d: %v", seed, err)
			}
		}
		units := []models.UnitInstance{}
		cells := map[models.Position]models.Cell{}
		for _, cell := range s.Board {
			cells[models.Position{X: cell.X, Y: cell.Y}] = cell
		}
		for _, player := range s.Players {
			units = append(units, player.Units...)
		}
		occupied := s.occupied()
		for i, unit := range units {
			for j := 0; j < i; j++ {
				if spawnDistance(unit.Position, units[j].Position) < minimumSpawnDistance {
					t.Fatalf("seed %d: spawns too close: %+v and %+v", seed, unit.Position, units[j].Position)
				}
			}
			cell := cells[unit.Position]
			if cell.Object != "" || cell.Terrain == "water" {
				t.Fatalf("seed %d: spawn inside an obstacle", seed)
			}
			exits := 0
			for _, d := range [][2]int{{1, 0}, {0, 1}, {-1, 0}, {0, -1}} {
				position := models.Position{X: unit.X + d[0], Y: unit.Y + d[1]}
				next, exists := cells[position]
				if exists && !occupied[position] && CanTraverse(cell, next) && movementCost(cell, next) <= unit.CurrentStamina {
					exits++
				}
			}
			if exits < 2 {
				t.Fatalf("seed %d: character has only %d usable exits", seed, exits)
			}
			if findPath(s.Board, units[0].Position, unit.Position, nil) == nil {
				t.Fatalf("seed %d: disconnected team spawn", seed)
			}
		}
	}
}

func TestSpawnCandidatesRejectStructureTrapsAndDeadEnds(t *testing.T) {
	board := flatSpawnBoard(15, 15)
	// Enclose (7,7) with structures; leave (0,0) only one possible exit.
	for _, p := range []models.Position{{X: 6, Y: 7}, {X: 8, Y: 7}, {X: 7, Y: 6}, {X: 7, Y: 8}, {X: 1, Y: 0}} {
		board[p.Y*15+p.X].Object = "house"
	}
	for _, position := range safeSpawnCandidates(board, 6) {
		if position == (models.Position{X: 7, Y: 7}) || position == (models.Position{}) {
			t.Fatalf("unsafe spawn candidate: %+v", position)
		}
	}
	if _, err := planSpawns(flatSpawnBoard(4, 4), "tiny", 6); err == nil {
		t.Fatal("accepted a map that cannot fit six separated spawns")
	}
}

func TestSpawnPlanDeterminismAndAtomicFailure(t *testing.T) {
	a, b := NewGameSession("spawn-repeat", 2), NewGameSession("spawn-repeat", 2)
	if err := a.prepareSpawns(3); err != nil {
		t.Fatal(err)
	}
	if err := b.prepareSpawns(3); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(a.Board, b.Board) || !reflect.DeepEqual(a.spawnSlots, b.spawnSlots) {
		t.Fatal("spawn plan changed for identical seed")
	}
	s := NewGameSession("tiny", 2)
	s.Board = flatSpawnBoard(4, 4)
	before := append([]models.Cell{}, s.Board...)
	id := uuid.Must(uuid.NewV4())
	if err := s.AddPlayer(nil, id, []models.Unit{testUnit("a"), testUnit("b"), testUnit("c")}); err == nil {
		t.Fatal("unsafe partial spawn accepted")
	}
	if len(s.Players) != 0 || len(s.TurnOrder) != 0 || !reflect.DeepEqual(s.Board, before) {
		t.Fatal("failed spawn mutated session")
	}
}
