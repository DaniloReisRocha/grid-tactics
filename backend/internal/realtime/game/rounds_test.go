package game

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/DaniloReisRocha/grid-tactics/backend/models"
	"github.com/gofrs/uuid"
)

func roundSession(t *testing.T) (*GameSession, uuid.UUID, uuid.UUID) {
	t.Helper()
	s, a, b := startedSession(t)
	for team, owner := range []uuid.UUID{a, b} {
		s.Players[owner].Units = nil
		for i := 0; i < 3; i++ {
			unit := testUnit(fmt.Sprintf("%d-%d", team, i))
			unit.Abilities = []models.Ability{models.AbilityOptions()[0]}
			s.Players[owner].Units = append(s.Players[owner].Units, models.UnitInstance{
				Unit: unit, OwnerID: owner, Player: team + 1, CurrentLife: 3,
				Position:       models.Position{X: i * 2, Y: team},
				CurrentStamina: 0, HasAttacked: true,
			})
		}
	}
	s.RoundCounter, s.TurnCounter = 0, 0
	s.RoundOrder = nil
	s.startRound()
	s.beginUnitTurn()
	return s, a, b
}

func TestRoundsGiveEveryLivingUnitOneAlternatingTurn(t *testing.T) {
	s, _, _ := roundSession(t)
	var previous uuid.UUID
	order := append([]models.UnitTurn(nil), s.RoundOrder...)
	for round := 1; round <= 30; round++ {
		if s.RoundCounter != round || len(s.RoundOrder) != 6 || s.RoundTurn != 0 {
			t.Fatalf("invalid round: %d %+v", s.RoundCounter, s.RoundOrder)
		}
		if !reflect.DeepEqual(order, s.RoundOrder) {
			t.Fatal("round changed the match turn order")
		}
		seen := map[string]bool{}
		for i := 0; i < 6; i++ {
			owner := *s.CurrentTurn
			unit := s.turnUnit(s.RoundOrder[s.RoundTurn])
			if owner == previous || seen[unit.Unit.ID] || s.CurrentUnitID != unit.Unit.ID || unit.HasAttacked || unit.CurrentStamina != unit.Unit.EffectiveStamina() {
				t.Fatal("turn order, uniqueness, or per-unit resource reset failed")
			}
			seen[unit.Unit.ID] = true
			unit.CurrentStamina = 0
			unit.HasAttacked = true
			previous = owner
			if err := s.EndTurn(owner); err != nil {
				t.Fatal(err)
			}
		}
		if len(seen) != 6 {
			t.Fatal("unit missed its turn")
		}
	}
}

func TestOnlyActiveUnitCanMoveAttackOrUseAbilities(t *testing.T) {
	s, _, _ := roundSession(t)
	owner := *s.CurrentTurn
	var other *models.UnitInstance
	for i := range s.Players[owner].Units {
		if s.Players[owner].Units[i].Unit.ID != s.CurrentUnitID {
			other = &s.Players[owner].Units[i]
			break
		}
	}
	before := *other
	revision := s.Revision
	if s.MovePiece(owner, other.Unit.ID, models.Position{Y: 3}) == nil {
		t.Fatal("inactive unit moved")
	}
	if s.Attack(owner, other.Unit.ID, models.Position{}) == nil {
		t.Fatal("inactive unit attacked")
	}
	if s.UseAbility(owner, other.Unit.ID, other.Unit.Abilities[0].ID, other.Position) == nil {
		t.Fatal("inactive unit used a skill")
	}
	if s.Revision != revision || !reflect.DeepEqual(before, *other) {
		t.Fatal("rejected actions changed state")
	}
	activeID := s.CurrentUnitID
	if err := s.EndTurn(owner); err != nil {
		t.Fatal(err)
	}
	// Entering a teammate's future turn must not restore this inactive unit yet.
	for _, unit := range s.Players[owner].Units {
		if unit.Unit.ID != activeID && (unit.CurrentStamina != 0 || !unit.HasAttacked) {
			t.Fatal("restored a whole team instead of one unit")
		}
	}
}

func TestDeadUnitsSkippedAndUnequalTeamsFinishTheirRound(t *testing.T) {
	s, a, b := roundSession(t)
	s.RoundOrder = []models.UnitTurn{
		{PlayerID: a, UnitID: "0-0"}, {PlayerID: b, UnitID: "1-0"},
		{PlayerID: a, UnitID: "0-1"}, {PlayerID: b, UnitID: "1-1"},
		{PlayerID: a, UnitID: "0-2"}, {PlayerID: b, UnitID: "1-2"},
	}
	s.RoundTurn = 0
	s.beginUnitTurn()
	s.Players[b].Units[0].CurrentLife = 0
	if err := s.EndTurn(a); err != nil {
		t.Fatal(err)
	}
	order := append([]models.UnitTurn(nil), s.RoundOrder...)
	if s.CurrentUnitID != "0-1" || s.RoundTurn != 2 {
		t.Fatal("dead unit not skipped in the original sequence")
	}
	s.Players[b].Units[2].CurrentLife = 0
	if err := s.EndTurn(a); err != nil {
		t.Fatal(err)
	}
	if s.CurrentUnitID != "1-1" {
		t.Fatal("remaining unit lost its turn")
	}
	if err := s.EndTurn(b); err != nil {
		t.Fatal(err)
	}
	if s.CurrentUnitID != "0-2" {
		t.Fatal("last ally lost its turn")
	}
	if err := s.EndTurn(a); err != nil {
		t.Fatal(err)
	}
	if s.RoundCounter != 2 || !reflect.DeepEqual(order, s.RoundOrder) {
		t.Fatal("new round changed the match sequence")
	}
	seen := map[string]bool{}
	for i := 0; i < 4; i++ {
		if seen[s.CurrentUnitID] {
			t.Fatal("duplicate turn in unequal round")
		}
		seen[s.CurrentUnitID] = true
		if err := s.EndTurn(*s.CurrentTurn); err != nil {
			t.Fatal(err)
		}
	}
	if s.RoundCounter != 3 {
		t.Fatal("unequal round failed to complete")
	}
}

func TestGrenadeSelfEliminationAdvancesActiveUnit(t *testing.T) {
	s, a, b := roundSession(t)
	s.RoundOrder = []models.UnitTurn{{PlayerID: a, UnitID: "0-0"}, {PlayerID: b, UnitID: "1-0"}, {PlayerID: a, UnitID: "0-1"}, {PlayerID: b, UnitID: "1-1"}, {PlayerID: a, UnitID: "0-2"}, {PlayerID: b, UnitID: "1-2"}}
	s.RoundTurn = 0
	s.beginUnitTurn()
	actor := &s.Players[a].Units[0]
	grenade, _ := models.AbilityByID(models.AbilityGrenadeID)
	actor.Unit.Abilities = []models.Ability{grenade}
	actor.CurrentLife = 2
	if err := s.UseAbility(a, actor.Unit.ID, grenade.ID, actor.Position); err != nil {
		t.Fatal(err)
	}
	if s.Status != models.GameStatusPlaying || actor.CurrentLife != 0 || s.CurrentUnitID != "1-0" || *s.CurrentTurn != b {
		t.Fatal("match stuck on eliminated actor")
	}
}

func TestFixedOrderSkipsDeadFirstAndLastSlotsAcrossRounds(t *testing.T) {
	s, _, _ := roundSession(t)
	order := append([]models.UnitTurn(nil), s.RoundOrder...)
	s.turnUnit(order[0]).CurrentLife = 0
	s.turnUnit(order[len(order)-1]).CurrentLife = 0
	s.advanceUnitTurn()
	for round := 1; round <= 3; round++ {
		for index := 1; index < len(order)-1; index++ {
			if s.RoundCounter != round || s.RoundTurn != index || s.CurrentUnitID != order[index].UnitID || *s.CurrentTurn != order[index].PlayerID {
				t.Fatal("dead boundary slots changed the surviving sequence")
			}
			if err := s.EndTurn(*s.CurrentTurn); err != nil {
				t.Fatal(err)
			}
		}
		if !reflect.DeepEqual(order, s.RoundOrder) {
			t.Fatal("boundary deaths reordered the match")
		}
	}
}
