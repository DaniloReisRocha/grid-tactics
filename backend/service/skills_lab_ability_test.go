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
	for i := range input.Board {
		input.Board[i].Height = 0
		input.Board[i].Elevation = models.ElevationLow
		input.Board[i].Terrain = models.TerrainGrass
		input.Board[i].Object = models.ObjectNone
	}
	input.Tokens = nil
	for i := 0; i < 2; i++ {
		input.Tokens = append(input.Tokens, models.SkillToken{ID: fmt.Sprint(i), Team: i, Position: models.Position{X: 5 + i*4, Y: 5}, Life: 3, WeaponID: models.WeaponPistolID, AbilityIDs: []string{models.AbilityBarricadeID}})
	}
	for i := 0; i < 2; i++ {
		for _, delta := range []models.Position{{Y: -1}, {X: -1}, {X: 1}} {
			input.ActorID = fmt.Sprint(i)
			input.Target = models.Position{X: 5 + i*4 + delta.X, Y: 5 + delta.Y}
			result, err := ResolveSkillsLabAbility(input)
			if err != nil {
				t.Fatal("each unit must get three independent uses", err)
			}
			input.Board, input.Tokens, input.MatchAbilityUses = result.Board, result.Tokens, result.MatchAbilityUses
			cell := result.Board[input.Target.Y*15+input.Target.X]
			if cell.Object != models.ObjectBarricade || game.IsWalkable(cell) {
				t.Fatal("barricade lost passability")
			}
		}
		input.Target = models.Position{X: 5 + i*4, Y: 6}
		if _, err := ResolveSkillsLabAbility(input); err == nil {
			t.Fatal("fourth use by the same unit accepted")
		}
	}
	input.Board[4*15+9].Object = models.ObjectNone
	if _, err := ResolveSkillsLabAbility(input); err == nil {
		t.Fatal("destroying a barricade restored a per-match use")
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

func TestGrenadeIncludesAlliesAndHasIndependentUnitLimits(t *testing.T) {
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
	input.Tokens = result.Tokens
	input.MatchAbilityUses = result.MatchAbilityUses
	if _, err := ResolveSkillsLabAbility(input); err == nil {
		t.Fatal("same unit used a second grenade")
	}
	input.ActorID = "friend"
	input.Target = models.Position{X: 6, Y: 4}
	second, err := ResolveSkillsLabAbility(input)
	if err != nil {
		t.Fatal("another unit could not use its own grenade", err)
	}
	if second.Tokens[1].AbilityUses[models.AbilityGrenadeID] != 1 || second.Tokens[0].AbilityUses[models.AbilityGrenadeID] != 1 {
		t.Fatal("unit counters were not independent")
	}
	input.Tokens = second.Tokens
	if _, err := ResolveSkillsLabAbility(input); err == nil {
		t.Fatal("second unit used two grenades")
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
				input.Tokens[0].AbilityUses = map[string]int{models.AbilityBarricadeID: 3}
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
