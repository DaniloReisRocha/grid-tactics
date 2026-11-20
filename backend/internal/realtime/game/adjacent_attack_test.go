package game

import (
	"testing"

	"github.com/DaniloReisRocha/grid-tactics/backend/models"
)

func TestWeaponsHitAllEightAdjacentCellsExceptSniper(t *testing.T) {
	for _, weapon := range models.WeaponOptions() {
		for dx := -1; dx <= 1; dx++ {
			for dy := -1; dy <= 1; dy++ {
				if dx == 0 && dy == 0 {
					continue
				}
				s, a, b := startedSession(t)
				attacker, defender := &s.Players[a].Units[0], &s.Players[b].Units[0]
				attacker.Position = models.Position{X: 4, Y: 4}
				attacker.Unit.Weapons = []models.Weapon{weapon}
				attacker.TurnCount = 2
				defender.Position = models.Position{X: 4 + dx, Y: 4 + dy}
				defender.CurrentLife = 30
				err := s.Attack(a, "a", defender.Position)
				if weapon.ID == models.WeaponSniperID {
					if err == nil || attacker.HasAttacked || defender.CurrentLife != 30 {
						t.Fatal("sniper hit adjacent cell")
					}
				} else if err != nil || !attacker.HasAttacked || defender.CurrentLife != 30-weapon.Damage {
					t.Fatalf("%s failed at %d,%d: %v", weapon.ID, dx, dy, err)
				}
			}
		}
	}
}

func TestDiagonalMeleeCannotCutAnyBlockedCorner(t *testing.T) {
	for _, direction := range []models.Position{{X: 1, Y: 1}, {X: -1, Y: 1}, {X: 1, Y: -1}, {X: -1, Y: -1}} {
		for _, object := range []string{models.ObjectHouse, models.ObjectRuin, models.ObjectBarricade, models.ObjectTree, models.ObjectRock, models.ObjectForest} {
			for side := 0; side < 2; side++ {
				s, a, b := startedSession(t)
				attacker, defender := &s.Players[a].Units[0], &s.Players[b].Units[0]
				attacker.Position = models.Position{X: 4, Y: 4}
				katana, _ := models.WeaponByID(models.WeaponKatanaID)
				attacker.Unit.Weapons = []models.Weapon{katana}
				defender.Position = models.Position{X: 4 + direction.X, Y: 4 + direction.Y}
				defender.CurrentLife = 30
				corner := models.Position{X: defender.X, Y: attacker.Y}
				if side == 1 {
					corner = models.Position{X: attacker.X, Y: defender.Y}
				}
				s.Board[corner.Y*10+corner.X].Object = object
				err := s.Attack(a, "a", defender.Position)
				blocked := object == models.ObjectHouse || object == models.ObjectRuin || object == models.ObjectBarricade
				if (err != nil) != blocked {
					t.Fatalf("incorrect corner rule for %s, side %d, direction %v: %v", object, side, direction, err)
				}
				if blocked && (attacker.HasAttacked || defender.CurrentLife != 30) {
					t.Fatal("blocked diagonal consumed attack or life")
				}
			}
		}
	}
}

func TestDiagonalBarricadeCanBeBrokenAndHammerPushFollowsDiagonal(t *testing.T) {
	s, a, b := startedSession(t)
	s.Players[a].Units[0].Position = models.Position{X: 4, Y: 4}
	s.Players[b].Units[0].Position = models.Position{X: 8, Y: 8}
	point := models.Position{X: 5, Y: 5}
	s.Board[55].Object = models.ObjectBarricade
	if err := s.Attack(a, "a", point); err != nil {
		t.Fatal(err)
	}
	if s.Board[55].Object != models.ObjectNone || !s.Players[a].Units[0].HasAttacked {
		t.Fatal("diagonal barricade did not break")
	}
	weapon, _ := models.WeaponByID(models.WeaponWarHammerID)
	for _, d := range []models.Position{{X: 1, Y: 1}, {X: -1, Y: 1}, {X: 1, Y: -1}, {X: -1, Y: -1}} {
		for blockedStep := 0; blockedStep <= 2; blockedStep++ {
			cells := map[models.Position]models.Cell{}
			for _, cell := range s.Board {
				cells[models.Position{X: cell.X, Y: cell.Y}] = cell
			}
			from := models.Position{X: 4, Y: 4}
			target := models.Position{X: 4 + d.X, Y: 4 + d.Y}
			if blockedStep > 0 {
				corner := models.Position{X: target.X + d.X*blockedStep, Y: target.Y + d.Y*(blockedStep-1)}
				cell := cells[corner]
				cell.Object = models.ObjectHouse
				cells[corner] = cell
			}
			path := WeaponKnockback(weapon, from, target, cells, nil)
			length := 3
			if blockedStep == 1 {
				length = 0
			}
			if blockedStep == 2 {
				length = 2
			}
			if len(path) != length {
				t.Fatalf("diagonal knockback ignored corner at step %d: %v", blockedStep, path)
			}
		}
	}
}
