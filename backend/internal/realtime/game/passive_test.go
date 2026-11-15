package game

import (
	"testing"

	"github.com/DaniloReisRocha/grid-tactics/backend/models"
	"github.com/gofrs/uuid"
)

func TestPassiveShieldGrantedOnlyAtSpawn(t *testing.T) {
	s := NewGameSession("passive-shield", 2)
	a, b := uuid.Must(uuid.NewV4()), uuid.Must(uuid.NewV4())
	shield, _ := models.AbilityByID(models.AbilityShieldID)
	unit := testUnit("shielded")
	unit.Abilities = []models.Ability{shield}
	if err := s.AddPlayer(nil, a, []models.Unit{unit}); err != nil {
		t.Fatal(err)
	}
	if err := s.AddPlayer(nil, b, []models.Unit{testUnit("plain")}); err != nil {
		t.Fatal(err)
	}
	actor := &s.Players[a].Units[0]
	if !shield.Passive || actor.CurrentShield != 3 || s.Players[b].Units[0].CurrentShield != 0 {
		t.Fatal("spawn did not distinguish equipped passive from shield capacity")
	}
	for _, remaining := range []int{1, 0} {
		actor.CurrentShield = remaining
		for i := 0; i < 4; i++ {
			if err := s.EndTurn(*s.CurrentTurn); err != nil {
				t.Fatal(err)
			}
			if actor.CurrentShield != remaining {
				t.Fatal("turn or round regenerated passive shield")
			}
		}
		if err := s.AddPlayer(nil, a, []models.Unit{unit}); err != nil {
			t.Fatal(err)
		}
		if actor.CurrentShield != remaining {
			t.Fatal("reconnecting regenerated passive shield")
		}
		if *s.CurrentTurn != a {
			if err := s.EndTurn(*s.CurrentTurn); err != nil {
				t.Fatal(err)
			}
		}
		revision := s.Revision
		if s.UseAbility(a, unit.ID, shield.ID, actor.Position) == nil {
			t.Fatal("passive shield accepted as active ability")
		}
		if s.Revision != revision || actor.CurrentShield != remaining || actor.AbilityUses[shield.ID] != 0 {
			t.Fatal("activation attempt mutated passive state")
		}
	}
}
