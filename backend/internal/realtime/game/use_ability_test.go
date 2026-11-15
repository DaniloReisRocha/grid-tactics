package game

import (
	"github.com/DaniloReisRocha/grid-tactics/backend/models"
	"testing"
)

func TestMultiplayerSkillOwnershipTurnAndConsumption(t *testing.T) {
	s, a, b := startedSession(t)
	shield, _ := models.AbilityByID(models.AbilityShieldID)
	unit := &s.Players[a].Units[0]
	unit.Position = models.Position{X: 1, Y: 1}
	s.Players[b].Units[0].Position = models.Position{X: 2, Y: 1}
	unit.Unit.Abilities = []models.Ability{shield}
	unit.CurrentStamina = 0
	if err := s.UseAbility(b, "a", shield.ID, unit.Position); err == nil {
		t.Fatal("opponent used an ability out of turn")
	}
	if err := s.UseAbility(a, "b", shield.ID, s.Players[b].Units[0].Position); err == nil {
		t.Fatal("used another player's unit")
	}
	if err := s.UseAbility(a, "a", shield.ID, unit.Position); err != nil {
		t.Fatal(err)
	}
	if unit.CurrentShield != 3 || unit.CurrentStamina != 0 || unit.HasAttacked || unit.AbilityUses[shield.ID] != 1 {
		t.Fatal("ability charged attack/stamina or did not apply shield")
	}
	if err := s.EndTurn(a); err != nil {
		t.Fatal(err)
	}
	if err := s.EndTurn(b); err != nil {
		t.Fatal(err)
	}
	unit.CurrentShield = 0
	if err := s.UseAbility(a, "a", shield.ID, unit.Position); err == nil {
		t.Fatal("turn renewed a consumed single-use skill")
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
	if s.Status != "finished" || s.Winner != nil || s.CurrentTurn != nil || s.MatchAbilityUses[grenade.ID] != 1 {
		t.Fatal("simultaneous elimination was not recorded")
	}
}
