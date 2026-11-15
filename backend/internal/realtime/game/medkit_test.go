package game

import (
	"testing"

	"github.com/DaniloReisRocha/grid-tactics/backend/models"
)

func TestMedkitInCombatCanHealSelfThenAllyAfterOrb(t *testing.T) {
	s, a, b := startedSession(t)
	medkit, _ := models.AbilityByID(models.AbilityMedkitID)
	unit := &s.Players[a].Units[0]
	unit.Unit.Abilities = []models.Ability{medkit}
	unit.Position = models.Position{}
	unit.CurrentLife, unit.CurrentStamina, unit.HasAttacked = 1, 0, true
	s.Players[b].Units[0].Position = models.Position{X: 8, Y: 8}
	if err := s.UseAbility(a, "a", medkit.ID, unit.Position); err != nil {
		t.Fatal(err)
	}
	if unit.CurrentLife != 3 || unit.CurrentStamina != 0 || !unit.HasAttacked {
		t.Fatal("medkit charged stamina/attack or failed healing")
	}
	if err := s.UseAbility(b, "b", medkit.ID, unit.Position); err == nil {
		t.Fatal("out-of-turn healing accepted")
	}
	s.Players[a].Units = append(s.Players[a].Units, models.UnitInstance{Unit: testUnit("ally"), OwnerID: a, Player: 1, Position: models.Position{X: 1}, CurrentLife: 1})
	unit = &s.Players[a].Units[0]
	if s.UseAbility(a, "a", medkit.ID, models.Position{X: 1}) == nil {
		t.Fatal("used medkit twice")
	}
	unit.RestoreOrbAbilities()
	if err := s.UseAbility(a, "a", medkit.ID, models.Position{X: 1}); err != nil {
		t.Fatal("orb did not restore medkit", err)
	}
	if s.Players[a].Units[1].CurrentLife != 3 || unit.AbilityUses[medkit.ID] != 1 {
		t.Fatal("ally not healed after orb")
	}
}
