package game

import (
	"testing"
	"time"

	"github.com/DaniloReisRocha/grid-tactics/backend/models"
)

func TestGrenadeAnimationEventFollowsSuccessfulDetonation(t *testing.T) {
	s, a, b := startedSession(t)
	actor, enemy := &s.Players[a].Units[0], &s.Players[b].Units[0]
	actor.Position = models.Position{X: 4, Y: 4}
	enemy.Position = models.Position{X: 5, Y: 4}
	enemy.CurrentLife = 30
	grenade, _ := models.AbilityByID(models.AbilityGrenadeID)
	medkit, _ := models.AbilityByID(models.AbilityMedkitID)
	actor.Unit.Abilities = []models.Ability{grenade, medkit}
	actor.CurrentLife = 1
	if err := s.UseAbility(a, "a", medkit.ID, actor.Position); err != nil {
		t.Fatal(err)
	}
	if s.GrenadeEvent != nil {
		t.Fatal("non-grenade ability triggered explosion")
	}
	if s.UseAbility(a, "a", grenade.ID, models.Position{X: 9, Y: 9}) == nil || s.GrenadeEvent != nil {
		t.Fatal("invalid grenade emitted event")
	}
	target := models.Position{X: 6, Y: 4}
	before := s.Revision
	if err := s.UseAbility(a, "a", grenade.ID, target); err != nil {
		t.Fatal(err)
	}
	event := s.GrenadeEvent
	if event == nil || event.Position != target || event.ID != before+1 || time.Now().UnixMilli()-event.CreatedAtMS > 1000 {
		t.Fatal("detonation event omitted target, ID or timestamp")
	}
	if s.UseAbility(a, "a", grenade.ID, target) == nil || s.GrenadeEvent != event {
		t.Fatal("used grenade emitted another event")
	}
	actor.RestoreOrbAbilities()
	if err := s.UseAbility(a, "a", grenade.ID, models.Position{X: 5, Y: 4}); err != nil {
		t.Fatal(err)
	}
	if s.GrenadeEvent.ID <= event.ID || event.Position != target {
		t.Fatal("second detonation reused ID or changed previous snapshot")
	}
}
