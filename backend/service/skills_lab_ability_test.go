package service

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	"github.com/DaniloReisRocha/grid-tactics/backend/internal/realtime/game"
	"github.com/DaniloReisRocha/grid-tactics/backend/models"
)

func abilityFixture(id string) models.SkillsLabAbility {
	return models.SkillsLabAbility{Board: SkillsLabConfig("test", "arena").Board, ActorID: "a", AbilityID: id, Target: models.Position{X: 5, Y: 5}, Tokens: []models.SkillToken{
		{ID: "a", Position: models.Position{X: 5, Y: 5}, Life: 3, MaxLife: 3, MaxShield: 3, WeaponID: models.WeaponPistolID, AbilityIDs: []string{id}, StaminaSpent: 4},
	}}
}

func TestPassiveShieldCannotBeActivated(t *testing.T) {
	input := abilityFixture(models.AbilityShieldID)
	before, _ := json.Marshal(input)
	_, err := ResolveSkillsLabAbility(input)
	if err == nil {
		t.Fatal("passive shield accepted as an active ability")
	}
	after, _ := json.Marshal(input)
	if string(before) != string(after) {
		t.Fatal("ability mutated input state")
	}
}

func TestBarricadesBlockMovementAndLimitActiveStructures(t *testing.T) {
	input := abilityFixture(models.AbilityBarricadeID)
	input.Tokens = nil
	for i := 0; i < 4; i++ {
		input.Tokens = append(input.Tokens, models.SkillToken{ID: fmt.Sprint(i), Position: models.Position{X: 2 + i*2, Y: 3}, Life: 3, WeaponID: models.WeaponPistolID, AbilityIDs: []string{models.AbilityBarricadeID}})
	}
	for i := 0; i < 3; i++ {
		input.ActorID = fmt.Sprint(i)
		input.Target = models.Position{X: 2 + i*2, Y: 4}
		result, err := ResolveSkillsLabAbility(input)
		if err != nil {
			t.Fatal(err)
		}
		input.Board, input.Tokens, input.MatchAbilityUses = result.Board, result.Tokens, result.MatchAbilityUses
		cell := result.Board[4*15+2+i*2]
		if cell.Object != models.ObjectBarricade || game.IsWalkable(cell) {
			t.Fatal("barricade did not block its tile")
		}
	}
	input.ActorID = "3"
	input.Target = models.Position{X: 8, Y: 4}
	if _, err := ResolveSkillsLabAbility(input); err == nil {
		t.Fatal("fourth active barricade accepted")
	}
	input.Board[4*15+2].Object = models.ObjectNone
	if _, err := ResolveSkillsLabAbility(input); err != nil {
		t.Fatal("limit must count active barricades, not lifetime uses:", err)
	}
}

func TestTrampolineCanLeapOverWaterOrClimbOneLevel(t *testing.T) {
	for _, climb := range []bool{false, true} {
		input := abilityFixture(models.AbilityTrampolineID)
		input.Target = models.Position{X: 7, Y: 5}
		if climb {
			input.Target.X = 6
			input.Board[5*15+6].Height = 1
		} else {
			input.Board[5*15+6].Terrain = models.TerrainWater
		}
		result, err := ResolveSkillsLabAbility(input)
		if err != nil {
			t.Fatal(err)
		}
		unit := result.Tokens[0]
		if unit.Position != input.Target || unit.StaminaSpent != 4 || unit.MovementMode != "jump" || len(unit.Path) != 2 {
			t.Fatalf("invalid leap: %+v", unit)
		}
	}
}

func TestGrenadeIncludesAlliesBreaksShieldWithoutOverflowAndHasMatchLimit(t *testing.T) {
	input := abilityFixture(models.AbilityGrenadeID)
	input.Target = models.Position{X: 6, Y: 5}
	input.Board[5*15+6].Object = models.ObjectBarricade
	input.Board[5*15+8].Object = models.ObjectBarricade
	input.Board[4*15+7].Object = models.ObjectHouse
	for _, token := range []models.SkillToken{
		{ID: "friend", Team: 0, Position: models.Position{X: 6, Y: 4}, Life: 3, Shield: 1},
		{ID: "enemy", Team: 1, Position: models.Position{X: 7, Y: 6}, Life: 2},
		{ID: "outside", Team: 1, Position: models.Position{X: 8, Y: 5}, Life: 3},
		{ID: "shielded", Team: 1, Position: models.Position{X: 6, Y: 6}, Life: 3, Shield: 4},
	} {
		token.WeaponID = models.WeaponPistolID
		token.AbilityIDs = []string{models.AbilityGrenadeID}
		input.Tokens = append(input.Tokens, token)
	}
	result, err := ResolveSkillsLabAbility(input)
	if err != nil {
		t.Fatal(err)
	}
	if result.Board[5*15+6].Object != models.ObjectNone || result.Board[5*15+8].Object != models.ObjectBarricade || result.Board[4*15+7].Object != models.ObjectHouse {
		t.Fatal("grenade did not respect structure type or blast area")
	}
	if result.Tokens[0].Life != 1 || result.Tokens[1].Life != 3 || result.Tokens[1].Shield != 0 || result.Tokens[2].Life != 0 || result.Tokens[3].Life != 3 || result.Tokens[4].Life != 3 || result.Tokens[4].Shield != 0 {
		t.Fatalf("invalid grenade result: %+v", result.Tokens)
	}
	input.Tokens = result.Tokens[1:] // Removing the original caster must not renew the grenade.
	input.MatchAbilityUses = result.MatchAbilityUses
	input.ActorID = "friend"
	input.Target = models.Position{X: 6, Y: 4}
	if _, err := ResolveSkillsLabAbility(input); err == nil {
		t.Fatal("another unit bypassed the match-wide grenade limit")
	}
}

func TestInvalidAbilitiesDoNotChangeOrConsumeState(t *testing.T) {
	for _, name := range []string{"diagonal", "height", "occupied", "water", "outside", "used", "unequipped", "dead", "jump one flat", "jump two high", "jump blocked", "unknown counter"} {
		t.Run(name, func(t *testing.T) {
			input := abilityFixture(models.AbilityBarricadeID)
			input.Target = models.Position{X: 6, Y: 5}
			switch name {
			case "diagonal":
				input.Target.Y = 6
			case "height":
				input.Board[5*15+6].Height = 1
			case "occupied":
				input.Target = input.Tokens[0].Position
			case "water":
				input.Board[5*15+6].Terrain = models.TerrainWater
			case "outside":
				input.Target.X = 99
			case "used":
				input.Tokens[0].AbilityUses = map[string]int{models.AbilityBarricadeID: 1}
			case "unequipped":
				input.Tokens[0].AbilityIDs = nil
			case "dead":
				input.Tokens[0].Life = 0
			case "jump one flat", "jump two high", "jump blocked":
				input.AbilityID = models.AbilityTrampolineID
				input.Tokens[0].AbilityIDs = []string{models.AbilityTrampolineID}
				if name == "jump two high" {
					input.Board[5*15+6].Height = 2
				}
				if name == "jump blocked" {
					input.Target.X = 7
					input.Board[5*15+7].Object = models.ObjectHouse
				}
			case "unknown counter":
				input.MatchAbilityUses = map[string]int{"invalid": 1}
			}
			before, _ := json.Marshal(input)
			if _, err := ResolveSkillsLabAbility(input); err == nil {
				t.Fatal("invalid action accepted")
			}
			after, _ := json.Marshal(input)
			if !reflect.DeepEqual(before, after) {
				t.Fatal("invalid action mutated state")
			}
		})
	}
}
