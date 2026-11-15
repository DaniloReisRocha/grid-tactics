package game

import (
	"testing"

	"github.com/DaniloReisRocha/grid-tactics/backend/models"
	"github.com/gofrs/uuid"
)

func TestWeaponPenaltyReducesSpawnAndTurnStamina(t *testing.T) {
	for _, weapon := range models.WeaponOptions() {
		t.Run(weapon.ID, func(t *testing.T) {
			s := NewGameSession("stamina-"+weapon.ID, 2)
			a, b := uuid.Must(uuid.NewV4()), uuid.Must(uuid.NewV4())
			unit := models.Unit{ID: "a", LP: 3, Stamina: 4, Weapons: []models.Weapon{weapon}}
			if err := s.AddPlayer(nil, a, []models.Unit{unit}); err != nil {
				t.Fatal(err)
			}
			unit.ID = "b"
			if err := s.AddPlayer(nil, b, []models.Unit{unit}); err != nil {
				t.Fatal(err)
			}
			expected := 4 - weapon.StaminaPenalty
			if s.Players[a].Units[0].CurrentStamina != expected {
				t.Fatal("spawn ignored equipped weapon penalty")
			}
			s.Players[a].Units[0].CurrentStamina = 0
			if err := s.EndTurn(a); err != nil {
				t.Fatal(err)
			}
			if err := s.EndTurn(b); err != nil {
				t.Fatal(err)
			}
			if s.Players[a].Units[0].CurrentStamina != expected || s.Players[a].Units[0].Unit.Stamina != 4 {
				t.Fatal("turn reset ignored penalty or modified base stamina")
			}
		})
	}
	unit := models.Unit{Stamina: 4, Weapons: []models.Weapon{{StaminaPenalty: 999}}}
	if unit.EffectiveStamina() != 0 {
		t.Fatal("penalty made maximum stamina negative")
	}
}
