package game

import (
	"testing"

	"github.com/DaniloReisRocha/grid-tactics/backend/models"
)

func TestEachUnitCanAttackOnceAtZeroStaminaAndResetsOnItsNextTurn(t *testing.T) {
	s, a, b := startedSession(t)
	s.Players[a].Units[0].Position = models.Position{}
	s.Players[a].Units[0].CurrentStamina = 0
	s.Players[a].Units[0].Unit.Weapons[0].StaminaPenalty = 999
	s.Players[a].Units = append(s.Players[a].Units, models.UnitInstance{
		Position: models.Position{Y: 1}, OwnerID: a, Player: 1, Unit: testUnit("a2"), CurrentLife: 3, CurrentStamina: 0,
	})
	s.Players[b].Units[0].Position = models.Position{X: 1}
	s.Players[b].Units[0].CurrentLife = 30
	s.Players[b].Units[0].CurrentShield = 0
	for _, id := range []string{"a", "a2"} {
		if err := s.Attack(a, id, models.Position{X: 1}); err != nil {
			t.Fatalf("%s could not attack at zero stamina: %v", id, err)
		}
		life := s.Players[b].Units[0].CurrentLife
		if err := s.Attack(a, id, models.Position{X: 1}); err == nil {
			t.Fatalf("%s attacked twice", id)
		}
		if s.Players[b].Units[0].CurrentLife != life {
			t.Fatal("rejected attack changed target life")
		}
	}
	for _, unit := range s.Players[a].Units {
		if !unit.HasAttacked || unit.CurrentStamina != 0 || unit.Unit.Stamina != 6 {
			t.Fatal("attack availability/stamina was not kept independent")
		}
	}
	if err := s.AddPlayer(nil, a, []models.Unit{testUnit("replacement")}); err != nil {
		t.Fatal(err)
	}
	if !s.Players[a].Units[0].HasAttacked {
		t.Fatal("rejoining restored an already-used attack")
	}
	if err := s.EndTurn(a); err != nil {
		t.Fatal(err)
	}
	if !s.Players[a].Units[0].HasAttacked {
		t.Fatal("attack was reset during the opponent's turn")
	}
	if err := s.EndTurn(b); err != nil {
		t.Fatal(err)
	}
	for i := range s.Players[a].Units {
		unit := &s.Players[a].Units[i]
		if unit.HasAttacked {
			t.Fatal("attack not restored for the next own turn")
		}
		unit.CurrentStamina = 0
	}
	if err := s.Attack(a, "a", models.Position{X: 1}); err != nil {
		t.Fatal(err)
	}
}

func TestInvalidAttackDoesNotConsumeAllowance(t *testing.T) {
	s, a, b := startedSession(t)
	s.Players[a].Units[0].Position = models.Position{}
	s.Players[a].Units[0].CurrentStamina = 0
	s.Players[b].Units[0].Position = models.Position{X: 1}
	for _, target := range []models.Position{{}, {X: 99}, {X: 2}} {
		if err := s.Attack(a, "a", target); err == nil {
			t.Fatal("invalid target accepted")
		}
		if s.Players[a].Units[0].HasAttacked {
			t.Fatal("invalid attempt used attack allowance")
		}
	}
	if err := s.Attack(b, "b", models.Position{}); err == nil {
		t.Fatal("opponent attacked out of turn")
	}
	if s.Players[b].Units[0].HasAttacked {
		t.Fatal("out-of-turn attempt consumed attack")
	}
	if err := s.Attack(a, "a", models.Position{X: 1}); err != nil {
		t.Fatal(err)
	}
}

func TestEmptyAndDeadSlotsNeverConsumeAttack(t *testing.T) {
	s, a, b := startedSession(t)
	s.Players[a].Units[0].Position = models.Position{}
	s.Players[a].Units[0].CurrentStamina = 0
	s.Players[b].Units[0].Position = models.Position{X: 1}
	s.Players[b].Units = append(s.Players[b].Units, models.UnitInstance{
		Position: models.Position{X: 2}, OwnerID: b, Unit: testUnit("dead"), CurrentLife: 0,
	})
	revision := s.Revision
	for _, target := range []models.Position{{Y: 1}, {X: 2, Y: 1}, {X: 2}, {Y: 1}} {
		if err := s.Attack(a, "a", target); err == nil {
			t.Fatal("accepted an attack on an empty/dead slot")
		}
		if unit := s.Players[a].Units[0]; unit.HasAttacked || unit.CurrentStamina != 0 {
			t.Fatal("empty-slot attempt consumed attack or stamina")
		}
		if s.Revision != revision {
			t.Fatal("empty-slot attempt changed authoritative game state")
		}
	}
	if err := s.Attack(a, "a", models.Position{X: 1}); err != nil {
		t.Fatal(err)
	}
	if !s.Players[a].Units[0].HasAttacked {
		t.Fatal("valid attack was not consumed")
	}
}
