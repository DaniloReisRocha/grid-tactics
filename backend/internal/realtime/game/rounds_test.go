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

func TestDeadNextUnitReplacedByAnotherOpponent(t *testing.T) {
	s, a, b := roundSession(t)
	s.RoundOrder = []models.UnitTurn{
		{PlayerID: a, UnitID: "0-0"}, {PlayerID: b, UnitID: "1-0"},
		{PlayerID: a, UnitID: "0-1"}, {PlayerID: b, UnitID: "1-1"},
		{PlayerID: a, UnitID: "0-2"}, {PlayerID: b, UnitID: "1-2"},
	}
	s.RoundTurn = 0
	s.beginUnitTurn()
	s.Players[b].Units[0].CurrentLife = 0
	s.Players[b].Units[2].CurrentLife = 0
	s.changed()
	if s.RoundOrder[1].UnitID != "1-1" || s.CurrentUnitID != "0-0" {
		t.Fatal("upcoming order was not recalculated immediately after elimination")
	}
	for _, expected := range []string{"1-1", "0-1", "1-1", "0-2", "1-1"} {
		previous := *s.CurrentTurn
		if err := s.EndTurn(previous); err != nil {
			t.Fatal(err)
		}
		if *s.CurrentTurn == previous || s.CurrentUnitID != expected {
			t.Fatalf("expected %s, got %s", expected, s.CurrentUnitID)
		}
	}
	if s.RoundCounter != 2 {
		t.Fatal("round did not advance after every living unit played")
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

func TestEverySurvivingTeamConfigurationAlwaysAlternates(t *testing.T) {
	// Every nonempty subset of both three-unit teams, including dead first/last
	// slots and the active unit dying before advance.
	for maskA := 1; maskA < 8; maskA++ {
		for maskB := 1; maskB < 8; maskB++ {
			t.Run(fmt.Sprintf("%d-%d", maskA, maskB), func(t *testing.T) {
				s, a, b := roundSession(t)
				for _, team := range []struct {
					owner uuid.UUID
					mask  int
				}{{a, maskA}, {b, maskB}} {
					for i := range s.Players[team.owner].Units {
						if team.mask&(1<<i) == 0 {
							s.Players[team.owner].Units[i].CurrentLife = 0
						}
					}
				}
				s.advanceUnitTurn()
				seen := map[string]bool{}
				round := s.RoundCounter
				// Begin checking complete rounds, after finishing the partial first one.
				for i := 0; i < 80; i++ {
					if s.turnUnit(models.UnitTurn{PlayerID: *s.CurrentTurn, UnitID: s.CurrentUnitID}) == nil {
						t.Fatal("dead unit scheduled")
					}
					seen[s.CurrentUnitID] = true
					previous := *s.CurrentTurn
					if err := s.EndTurn(previous); err != nil {
						t.Fatal(err)
					}
					if *s.CurrentTurn == previous {
						t.Fatal("same player received consecutive turns")
					}
					if s.RoundCounter != round {
						if round > 1 {
							for _, p := range s.Players {
								for _, u := range p.Units {
									if u.CurrentLife > 0 && !seen[u.Unit.ID] {
										t.Fatal("living unit starved")
									}
								}
							}
						}
						round = s.RoundCounter
						seen = map[string]bool{}
					}
				}
			})
		}
	}
}
