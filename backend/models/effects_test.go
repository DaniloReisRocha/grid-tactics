package models

import (
	"encoding/json"
	"testing"
)

func TestWeaponEffectsJSONAndIndependentCatalogCopies(t *testing.T) {
	weapons := WeaponOptions()
	encoded, err := json.Marshal(weapons)
	if err != nil {
		t.Fatal(err)
	}
	var decoded []map[string]json.RawMessage
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatal(err)
	}
	for _, weapon := range decoded {
		if string(weapon["effect"]) == "null" || len(weapon["effect"]) == 0 {
			t.Fatal("effect must always serialize as an array")
		}
	}
	sniper, _ := WeaponByID(WeaponSniperID)
	if !sniper.HasEffect(EffectNoAdjacentID) || !sniper.Effect[0].Negative {
		t.Fatal("adjacent-target restriction must be a negative effect")
	}
	hammer, _ := WeaponByID(WeaponWarHammerID)
	hammer.Effect[0].Name = "modified"
	hammer.Effect[1].Negative = false
	original, _ := WeaponByID(WeaponWarHammerID)
	if original.Effect[0].Name == "modified" || !original.Effect[1].Negative {
		t.Fatal("weapon lookups share mutable effects")
	}
	for _, weapon := range weapons {
		for _, effect := range weapon.Effect {
			definition, ok := EffectByID(effect.ID)
			if !ok || effect != definition {
				t.Fatalf("weapon %s has unknown or duplicated definition: %+v", weapon.ID, effect)
			}
		}
	}
}
