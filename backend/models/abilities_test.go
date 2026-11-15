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
		if ability.Uses != AbilityUsageSingle && ability.Uses != AbilityUsageUnlimited {
			t.Fatalf("unknown usage: %s", ability.Uses)
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
	// The grenade's inventory usage policy does not remove its match restriction.
	grenade, _ := AbilityByID(AbilityGrenadeID)
	if grenade.Uses != AbilityUsageUnlimited || grenade.MaxUsesPerMatch != 1 {
		t.Fatal("grenade lost its separate one-use match limit")
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
