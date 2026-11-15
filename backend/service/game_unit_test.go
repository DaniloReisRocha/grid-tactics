package service

import (
	"testing"

	"github.com/DaniloReisRocha/grid-tactics/backend/internal/realtime/game"
	"github.com/DaniloReisRocha/grid-tactics/backend/models"
	"github.com/gofrs/uuid"
)

func TestStarterUnitResourcesAndShieldLifecycle(t *testing.T) {
	id := uuid.Must(uuid.NewV4())
	units, err := (&GameUnitsGetService{}).Run(id)
	if err != nil {
		t.Fatal(err)
	}
	for _, unit := range units {
		if unit.LP != models.UnitBaseLife || unit.Shield != models.UnitShieldCapacity || unit.Stamina != models.UnitBaseStamina {
			t.Fatalf("incorrect resources: %+v", unit)
		}
		if len(unit.Weapons) > unit.ItemSlots || len(unit.Abilities) > unit.AbilitySlots || unit.Abilities == nil {
			t.Fatal("invalid inventory or ability slots")
		}
	}
	session := game.NewGameSession("unit-resource-check", 2)
	if err := session.AddPlayer(nil, id, units); err != nil {
		t.Fatal(err)
	}
	for _, unit := range session.Players[id].Units {
		if unit.CurrentShield != 0 || unit.Unit.Shield != 3 || unit.CurrentLife != 3 || unit.CurrentStamina != 4 {
			t.Fatalf("unit started with automatic shield or wrong resources: %+v", unit)
		}
	}
	// A reconnect preserves a shield granted later; the zero-shield rule is spawn-only.
	session.Players[id].Units[0].CurrentShield = 2
	if err := session.AddPlayer(nil, id, units); err != nil {
		t.Fatal(err)
	}
	if session.Players[id].Units[0].CurrentShield != 2 {
		t.Fatal("reconnect reset the active shield")
	}
}
