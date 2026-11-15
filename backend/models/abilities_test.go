package models

import (
	"encoding/json"
	"testing"
)

func TestAbilityCatalogSerializesAndSharesEffectDefinitions(t *testing.T) {
	options := AbilityOptions()
	encoded, err := json.Marshal(options)
	if err != nil {
		t.Fatal(err)
	}
	var decoded []Ability
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatal(err)
	}
	ids := map[string]bool{}
	for _, ability := range decoded {
		if ability.ID == "" || ids[ability.ID] || ability.Effect == nil {
			t.Fatalf("invalid ability JSON: %+v", ability)
		}
		ids[ability.ID] = true
		if ability.MaxUsesPerMatch < 0 {
			t.Fatalf("invalid usage limit: %d", ability.MaxUsesPerMatch)
		}
		for _, effect := range ability.Effect {
			definition, ok := EffectByID(effect.ID)
			if !ok || effect != definition || !ability.HasEffect(effect.ID) {
				t.Fatalf("unregistered effect in %s", ability.ID)
			}
		}
	}
	if _, ok := AbilityByID("missing"); ok {
		t.Fatal("unknown ability resolved")
	}
	for id, limit := range map[string]int{AbilityBarricadeID: 3, AbilityGrenadeID: 1, AbilityTrampolineID: 1} {
		ability, _ := AbilityByID(id)
		if ability.MaxUsesPerMatch != limit {
			t.Fatalf("%s: incorrect per-unit limit", id)
		}
	}
	var raw []map[string]json.RawMessage
	if err := json.Unmarshal(encoded, &raw); err != nil {
		t.Fatal(err)
	}
	for _, ability := range raw {
		if _, exists := ability["uses"]; exists {
			t.Fatal("redundant uses serialized")
		}
		if _, exists := ability["maxActivePerMatch"]; exists {
			t.Fatal("redundant active limit serialized")
		}
	}
}

func TestAbilityLookupsDoNotShareMutableData(t *testing.T) {
	trampoline, _ := AbilityByID(AbilityTrampolineID)
	trampoline.Movement.ForwardDistance = 99
	trampoline.Effect[0].Name = "modified"
	copy, _ := AbilityByID(AbilityTrampolineID)
	if copy.Movement.ForwardDistance == 99 || copy.Effect[0].Name == "modified" {
		t.Fatal("ability lookups share pointers or effect slices")
	}
	barricade, _ := AbilityByID(AbilityBarricadeID)
	katana, _ := WeaponByID(WeaponKatanaID)
	if !barricade.HasEffect(EffectSameLevelID) || !katana.HasEffect(EffectSameLevelID) {
		t.Fatal("height restriction is not reusable across weapons and abilities")
	}
}
