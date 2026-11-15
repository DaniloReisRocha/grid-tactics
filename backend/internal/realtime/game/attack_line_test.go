package game

import (
	"testing"

	"github.com/DaniloReisRocha/grid-tactics/backend/models"
)

func TestStructuresBlockDirectAttacksInEveryDirection(t *testing.T) {
	for _, object := range []string{models.ObjectHouse, models.ObjectRuin, models.ObjectBarricade, models.ObjectTree, models.ObjectForest, models.ObjectRock} {
		for _, d := range []models.Position{{X: 1}, {X: -1}, {Y: 1}, {Y: -1}} {
			s, a, b := startedSession(t)
			attacker, defender := &s.Players[a].Units[0], &s.Players[b].Units[0]
			attacker.Position = models.Position{X: 4, Y: 4}
			defender.Position = models.Position{X: 4 + 2*d.X, Y: 4 + 2*d.Y}
			defender.CurrentLife = 30
			s.Board[(4+d.Y)*10+4+d.X].Object = object
			err := s.Attack(a, "a", defender.Position)
			blocked := object == models.ObjectHouse || object == models.ObjectRuin || object == models.ObjectBarricade
			if (err != nil) != blocked {
				t.Fatalf("%s direction %+v: %v", object, d, err)
			}
			if blocked && (attacker.HasAttacked || defender.CurrentLife != 30) {
				t.Fatal("blocked attack changed allowance or damaged target")
			}
		}
	}
}

func TestBarricadeBreaksWithoutPiercingAndConsumesAttack(t *testing.T) {
	s, a, b := startedSession(t)
	attacker, defender := &s.Players[a].Units[0], &s.Players[b].Units[0]
	attacker.Position = models.Position{}
	defender.Position = models.Position{X: 2}
	defender.CurrentLife = 30
	s.Board[1].Object = models.ObjectBarricade
	if err := s.Attack(a, "a", models.Position{X: 1}); err != nil {
		t.Fatal(err)
	}
	if s.Board[1].Object != models.ObjectNone || !walkable(s.Board[1]) || !attacker.HasAttacked || defender.CurrentLife != 30 {
		t.Fatal("barricade hit did not stop at the structure")
	}
	if err := s.Attack(a, "a", defender.Position); err == nil {
		t.Fatal("destroying a barricade did not consume the attack")
	}
	if err := s.EndTurn(a); err != nil {
		t.Fatal(err)
	}
	if err := s.EndTurn(b); err != nil {
		t.Fatal(err)
	}
	if err := s.Attack(a, "a", defender.Position); err != nil {
		t.Fatal(err)
	}
	if defender.CurrentLife != 26 {
		t.Fatal("target remained blocked after destruction")
	}
}
