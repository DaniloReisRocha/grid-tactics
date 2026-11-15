package service

import (
	"encoding/json"
	"testing"

	"github.com/DaniloReisRocha/grid-tactics/backend/models"
)

func TestMedkitHealsSelfAndAlliesAndCapsAtMaximum(t *testing.T) {
	for _, self := range []bool{true, false} {
		for _, life := range []int{1, 2} {
			input := abilityFixture(models.AbilityMedkitID)
			input.Tokens[0].Life = life
			index := 0
			if !self {
				index = 1
				input.Target.X++
				input.Tokens = append(input.Tokens, models.SkillToken{ID: "ally", Position: input.Target, Life: life, MaxLife: 3, Shield: 2, WeaponID: models.WeaponPistolID})
			}
			before, _ := json.Marshal(input)
			result, err := ResolveSkillsLabAbility(input)
			if err != nil {
				t.Fatal(err)
			}
			if result.Tokens[index].Life != 3 || result.Tokens[index].Shield != input.Tokens[index].Shield || result.Tokens[0].StaminaSpent != 4 || result.Tokens[0].AbilityUses[models.AbilityMedkitID] != 1 {
				t.Fatal("healing changed unrelated resources or exceeded maximum")
			}
			after, _ := json.Marshal(input)
			if string(before) != string(after) {
				t.Fatal("healing mutated input")
			}
			input.Tokens = result.Tokens
			input.Tokens[index].Life = 1
			if _, err := ResolveSkillsLabAbility(input); err == nil {
				t.Fatal("medkit reused without recovery")
			}
		}
	}
}

func TestMedkitRejectsInvalidTargetsWithoutConsumingUse(t *testing.T) {
	for _, kind := range []string{"enemy", "dead", "full", "empty", "diagonal", "far"} {
		t.Run(kind, func(t *testing.T) {
			input := abilityFixture(models.AbilityMedkitID)
			input.Target.X++
			ally := models.SkillToken{ID: "target", Position: input.Target, Life: 1, MaxLife: 3, WeaponID: models.WeaponPistolID}
			switch kind {
			case "enemy":
				ally.Team = 1
			case "dead":
				ally.Life = 0
			case "full":
				ally.Life = 3
			case "diagonal":
				input.Target.Y++
				ally.Position = input.Target
			case "far":
				input.Target.X++
				ally.Position = input.Target
			}
			if kind != "empty" {
				input.Tokens = append(input.Tokens, ally)
			}
			before, _ := json.Marshal(input)
			if _, err := ResolveSkillsLabAbility(input); err == nil {
				t.Fatal("invalid healing accepted")
			}
			after, _ := json.Marshal(input)
			if string(before) != string(after) {
				t.Fatal("invalid healing consumed use")
			}
		})
	}
}
