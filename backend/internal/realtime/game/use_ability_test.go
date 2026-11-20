package game

import (
	"github.com/DaniloReisRocha/grid-tactics/backend/models"
	"testing"
)

func TestMultiplayerSkillOwnershipTurnAndConsumption(t *testing.T) {
	s, a, b := startedSession(t)
	ability, _ := models.AbilityByID(models.AbilityTrampolineID)
	unit := &s.Players[a].Units[0]
	unit.Position = models.Position{X: 1, Y: 1}
	s.Players[b].Units[0].Position = models.Position{X: 2, Y: 1}
	unit.Unit.Abilities = []models.Ability{ability}
	target := models.Position{X: 1, Y: 3}
	unit.CurrentStamina = 0
	if err := s.UseAbility(b, "a", ability.ID, target); err == nil {
		t.Fatal("opponent used an ability out of turn")
	}
	if err := s.UseAbility(a, "b", ability.ID, target); err == nil {
		t.Fatal("used another player's unit")
	}
	if err := s.UseAbility(a, "a", ability.ID, target); err != nil {
		t.Fatal(err)
	}
	if unit.CurrentShield != 0 || unit.CurrentStamina != 0 || unit.HasAttacked || unit.AbilityUses[ability.ID] != 1 {
		t.Fatal("ability charged attack/stamina or did not record usage")
	}
	if err := s.EndTurn(a); err != nil {
		t.Fatal(err)
	}
	if err := s.EndTurn(b); err != nil {
		t.Fatal(err)
	}
	unit.CurrentShield = 0
	target = models.Position{X: 1, Y: 5}
	if err := s.UseAbility(a, "a", ability.ID, target); err == nil {
		t.Fatal("turn renewed a consumed single-use skill")
	}
}

func TestMultiplayerAbilityLimitsArePerUnitAndPersistAcrossRounds(t *testing.T) {
	s, a, b := startedSession(t)
	barricade, _ := models.AbilityByID(models.AbilityBarricadeID)
	grenade, _ := models.AbilityByID(models.AbilityGrenadeID)
	s.Players[a].Units[0].Position = models.Position{X: 3, Y: 3}
	s.Players[b].Units[0].Position = models.Position{X: 7, Y: 7}
	for _, player := range s.Players {
		player.Units[0].Unit.Abilities = []models.Ability{barricade, grenade}
	}
	for _, delta := range []models.Position{{Y: -1}, {X: -1}, {X: 1}} {
		for _, owner := range s.TurnOrder {
			unit := &s.Players[owner].Units[0]
			if err := s.UseAbility(owner, unit.Unit.ID, barricade.ID, models.Position{X: unit.X + delta.X, Y: unit.Y + delta.Y}); err != nil {
				t.Fatal("unit's barricade allowance was shared or reset", err)
			}
			if err := s.EndTurn(owner); err != nil {
				t.Fatal(err)
			}
		}
	}
	for _, owner := range s.TurnOrder {
		unit := &s.Players[owner].Units[0]
		if err := s.AddPlayer(nil, owner, []models.Unit{testUnit("replacement")}); err != nil {
			t.Fatal(err)
		}
		if s.UseAbility(owner, unit.Unit.ID, barricade.ID, models.Position{X: unit.X, Y: unit.Y + 1}) == nil {
			t.Fatal("fourth barricade accepted after reconnect")
		}
		if err := s.UseAbility(owner, unit.Unit.ID, grenade.ID, unit.Position); err != nil {
			t.Fatal("another unit's grenade prevented this use", err)
		}
		if s.UseAbility(owner, unit.Unit.ID, grenade.ID, unit.Position) == nil {
			t.Fatal("second grenade accepted for same unit")
		}
		if unit.AbilityUses[barricade.ID] != 3 || unit.AbilityUses[grenade.ID] != 1 {
			t.Fatal("individual counters not preserved")
		}
		if err := s.EndTurn(owner); err != nil {
			t.Fatal(err)
		}
	}
}

func TestMultiplayerGrenadeCanEliminateBothTeams(t *testing.T) {
	s, a, b := startedSession(t)
	grenade, _ := models.AbilityByID(models.AbilityGrenadeID)
	unit := &s.Players[a].Units[0]
	unit.Unit.Abilities = []models.Ability{grenade}
	unit.Position = models.Position{X: 4, Y: 4}
	unit.CurrentLife = 2
	s.Players[b].Units[0].Position = models.Position{X: 5, Y: 4}
	s.Players[b].Units[0].CurrentLife = 2
	if err := s.UseAbility(a, "a", grenade.ID, unit.Position); err != nil {
		t.Fatal(err)
	}
	if s.Status != models.GameStatusFinished || s.Winner != nil || s.CurrentTurn != nil || s.MatchAbilityUses[grenade.ID] != 1 {
		t.Fatal("simultaneous elimination was not recorded")
	}
}
