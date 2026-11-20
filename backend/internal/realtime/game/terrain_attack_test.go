package game

import (
	"testing"

	"github.com/DaniloReisRocha/grid-tactics/backend/models"
)

func TestReliefBlocksAttackButEqualPlateausRemainClear(t *testing.T) {
	for _, d := range []models.Position{{X: 1}, {X: -1}, {Y: 1}, {Y: -1}} {
		for _, tc := range []struct {
			source, target, middle int
			blocked                bool
		}{
			{0, 0, 0, false}, {0, 0, 1, true}, {1, 1, 1, false}, {1, 1, 2, true},
			{2, 2, 1, false}, {2, 2, 3, true}, {2, 0, 0, false}, {2, 0, 1, true},
			{0, 2, 0, false}, {0, 2, 1, true},
		} {
			s, a, b := startedSession(t)
			attacker, defender := &s.Players[a].Units[0], &s.Players[b].Units[0]
			attacker.Position = models.Position{X: 4, Y: 4}
			defender.Position = models.Position{X: 4 + 2*d.X, Y: 4 + 2*d.Y}
			defender.CurrentLife = 30
			s.Board[44].Height = tc.source
			s.Board[defender.Y*10+defender.X].Height = tc.target
			s.Board[(4+d.Y)*10+4+d.X].Height = tc.middle
			revision := s.Revision
			err := s.Attack(a, "a", defender.Position)
			if (err != nil) != tc.blocked {
				t.Fatalf("incorrect relief blocking for %+v in %v: %v", tc, d, err)
			}
			if tc.blocked && (attacker.HasAttacked || defender.CurrentLife != 30 || s.Revision != revision) {
				t.Fatal("blocked relief attack changed state")
			}
		}
	}
}

func TestReliefAlsoBlocksDiagonalCorners(t *testing.T) {
	for _, d := range []models.Position{{X: 1, Y: 1}, {X: -1, Y: 1}, {X: 1, Y: -1}, {X: -1, Y: -1}} {
		for side := 0; side < 2; side++ {
			for _, higher := range []bool{false, true} {
				s, a, b := startedSession(t)
				attacker, defender := &s.Players[a].Units[0], &s.Players[b].Units[0]
				attacker.Position = models.Position{X: 4, Y: 4}
				defender.Position = models.Position{X: 4 + d.X, Y: 4 + d.Y}
				defender.CurrentLife = 30
				for i := range s.Board {
					s.Board[i].Height = 1
				}
				corner := models.Position{X: defender.X, Y: attacker.Y}
				if side == 1 {
					corner = models.Position{X: attacker.X, Y: defender.Y}
				}
				if higher {
					s.Board[corner.Y*10+corner.X].Height = 2
				}
				if err := s.Attack(a, "a", defender.Position); (err != nil) != higher {
					t.Fatal("diagonal relief ignored height or blocked flat terrain", err)
				}
			}
		}
	}
}

func TestReliefUsesWholeCubeWhenFiringDownhill(t *testing.T) {
	cells := map[models.Position]models.Cell{}
	for x := 0; x < 4; x++ {
		cells[models.Position{X: x}] = models.Cell{X: x, Terrain: models.TerrainGrass}
	}
	from, to := models.Position{}, models.Position{X: 3}
	source := cells[from]
	source.Height = 2
	cells[from] = source
	near := cells[models.Position{X: 1}]
	near.Height = 1
	cells[models.Position{X: 1}] = near
	if !DirectAttackLineClear(from, to, cells) {
		t.Fatal("low relief below sight line blocked downhill attack")
	}
	far := cells[models.Position{X: 2}]
	far.Height = 1
	cells[models.Position{X: 2}] = far
	if DirectAttackLineClear(from, to, cells) {
		t.Fatal("sight line passed through far edge of elevated cube")
	}
}
