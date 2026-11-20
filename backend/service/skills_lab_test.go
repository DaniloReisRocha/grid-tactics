package service

import (
	"reflect"
	"testing"

	"github.com/DaniloReisRocha/grid-tactics/backend/models"
)

func labFixture(weaponID string, targetX int) models.SkillsLabAttack {
	config := SkillsLabConfig("test", "arena")
	return models.SkillsLabAttack{
		Board: config.Board,
		Tokens: []models.SkillToken{
			{ID: "attacker", Team: 0, Position: models.Position{X: 4, Y: 5}, Life: 6, MaxLife: 6, WeaponID: weaponID, TurnCount: 2},
			{ID: "target", Team: 1, Position: models.Position{X: targetX, Y: 5}, Life: 6, MaxLife: 6, Shield: 3, MaxShield: 3, WeaponID: models.WeaponPistolID},
		},
		AttackerID: "attacker", TargetID: "target",
	}
}

func TestLabWeaponEffectsAndRepeatedAttacks(t *testing.T) {
	for _, test := range []struct {
		weapon                  string
		x, life, shield, finalX int
	}{
		{models.WeaponPistolID, 6, 6, 2, 6},
		{models.WeaponShotgunID, 6, 6, 0, 6},
		{models.WeaponSniperID, 9, 6, 0, 9},
		{models.WeaponKatanaID, 5, 6, 0, 5},
		{models.WeaponWarHammerID, 5, 6, 1, 7},
	} {
		t.Run(test.weapon, func(t *testing.T) {
			input := labFixture(test.weapon, test.x)
			// Even an exhausted token can attack, without any further stamina cost.
			input.Tokens[0].StaminaSpent = models.UnitBaseStamina
			before := append([]models.SkillToken(nil), input.Tokens...)
			result, err := ResolveSkillsLabAttack(input)
			if err != nil {
				t.Fatal(err)
			}
			target := result.Tokens[1]
			if result.Tokens[0].StaminaSpent != models.UnitBaseStamina {
				t.Fatal("attack charged or restored movement stamina")
			}
			if target.Life != test.life || target.Shield != test.shield || target.X != test.finalX {
				t.Fatalf("unexpected target: %+v", target)
			}
			if !reflect.DeepEqual(before, input.Tokens) {
				t.Fatal("simulation mutated caller's tokens")
			}
			input.Tokens = result.Tokens
			// The lab allows further attacks without turns or movement stamina.
			if test.weapon != models.WeaponWarHammerID {
				if _, err := ResolveSkillsLabAttack(input); err != nil {
					t.Fatalf("repeat attack failed: %v", err)
				}
			}
		})
	}
	input := labFixture(models.WeaponPistolID, 6)
	input.Tokens[1].Shield = 0
	input.Tokens[1].Life = 1
	result, err := ResolveSkillsLabAttack(input)
	if err != nil || result.Tokens[1].Life != 0 {
		t.Fatalf("lethal hit: %+v, %v", result, err)
	}
	input.Tokens = result.Tokens
	if _, err := ResolveSkillsLabAttack(input); err == nil {
		t.Fatal("dead target accepted")
	}
}

func TestLabSniperReloadTurns(t *testing.T) {
	for turn := 0; turn <= 4; turn++ {
		input := labFixture(models.WeaponSniperID, 6)
		input.Tokens[0].TurnCount = turn
		before := append([]models.SkillToken(nil), input.Tokens...)
		_, err := ResolveSkillsLabAttack(input)
		if (err != nil) != (turn == 0 || turn%2 == 1) {
			t.Fatalf("unexpected reload behavior on turn %d: %v", turn, err)
		}
		if !reflect.DeepEqual(before, input.Tokens) {
			t.Fatal("reload/attack mutated input")
		}
	}
}

func TestLabAttackRestrictionsDoNotChangeTokens(t *testing.T) {
	for _, name := range []string{"sniper adjacent", "katana height", "hammer height", "out of range", "friendly", "dead attacker", "unknown weapon", "missing target", "too many abilities", "duplicate ability", "unknown ability"} {
		t.Run(name, func(t *testing.T) {
			input := labFixture(models.WeaponPistolID, 5)
			switch name {
			case "sniper adjacent":
				input.Tokens[0].WeaponID = models.WeaponSniperID
			case "katana height", "hammer height":
				input.Tokens[0].WeaponID = models.WeaponKatanaID
				if name == "hammer height" {
					input.Tokens[0].WeaponID = models.WeaponWarHammerID
				}
				for i := range input.Board {
					if input.Board[i].X == 5 && input.Board[i].Y == 5 {
						input.Board[i].Height = 1
					}
				}
			case "out of range":
				input.Tokens[1].X = 8
			case "friendly":
				input.Tokens[1].Team = 0
			case "dead attacker":
				input.Tokens[0].Life = 0
			case "unknown weapon":
				input.Tokens[0].WeaponID = "invalid"
			case "missing target":
				input.TargetID = "missing"
			case "too many abilities":
				input.Tokens[0].AbilityIDs = []string{models.AbilityShieldID, models.AbilityBarricadeID, models.AbilityTrampolineID}
			case "duplicate ability":
				input.Tokens[0].AbilityIDs = []string{models.AbilityShieldID, models.AbilityShieldID}
			case "unknown ability":
				input.Tokens[0].AbilityIDs = []string{"unknown"}
			}
			before := append([]models.SkillToken(nil), input.Tokens...)
			if _, err := ResolveSkillsLabAttack(input); err == nil {
				t.Fatal("invalid attack accepted")
			}
			if !reflect.DeepEqual(before, input.Tokens) {
				t.Fatal("rejected attack changed tokens")
			}
		})
	}
}

func TestLabHammerRespectsPushDestination(t *testing.T) {
	for _, obstacle := range []string{"higher", "water", "house", "occupied", "edge", "downhill"} {
		t.Run(obstacle, func(t *testing.T) {
			input := labFixture(models.WeaponWarHammerID, 5)
			if obstacle == "occupied" {
				token := input.Tokens[0]
				token.ID = "blocker"
				token.X = 6
				input.Tokens = append(input.Tokens, token)
			}
			board := []models.Cell{}
			for _, cell := range input.Board {
				if obstacle == "downhill" && cell.Y == 5 && (cell.X == 4 || cell.X == 5) {
					cell.Height = 1
				}
				if cell.X == 6 && cell.Y == 5 {
					switch obstacle {
					case "higher":
						cell.Height = 1
					case "water":
						cell.Terrain = models.TerrainWater
					case "house":
						cell.Object = models.ObjectHouse
					case "edge":
						continue
					}
				}
				board = append(board, cell)
			}
			input.Board = board
			result, err := ResolveSkillsLabAttack(input)
			if err != nil {
				t.Fatal(err)
			}
			wantX := 5
			if obstacle == "downhill" {
				wantX = 7
			}
			weapon, _ := models.WeaponByID(models.WeaponWarHammerID)
			if result.Tokens[1].X != wantX || result.Tokens[1].Shield != max(0, input.Tokens[1].Shield-weapon.Damage) {
				t.Fatalf("incorrect push: %+v", result.Tokens[1])
			}
		})
	}
}

func TestLabScenarioGenerationIsDeterministic(t *testing.T) {
	a, b := SkillsLabConfig("same-seed", "procedural"), SkillsLabConfig("same-seed", "procedural")
	if len(a.Board) != 225 || len(a.Weapons) != 5 || !reflect.DeepEqual(a, b) {
		t.Fatal("invalid or non-deterministic scenario")
	}
}

func TestLabRejectsDistantOffAxisTargetsForEveryWeapon(t *testing.T) {
	for _, weapon := range models.WeaponOptions() {
		for _, offset := range []models.Position{{X: 2, Y: 2}, {X: 1, Y: 2}, {X: -2, Y: 2}, {X: 2, Y: -2}, {X: -2, Y: -2}} {
			input := labFixture(weapon.ID, 5)
			input.Tokens[1].Position = models.Position{X: input.Tokens[0].X + offset.X, Y: input.Tokens[0].Y + offset.Y}
			if _, err := ResolveSkillsLabAttack(input); err == nil {
				t.Fatalf("%s accepted off-axis target %+v", weapon.ID, offset)
			}
		}
	}
}

func TestLabSupportsAllAdjacentAttackDirections(t *testing.T) {
	for _, weapon := range models.WeaponOptions() {
		for dx := -1; dx <= 1; dx++ {
			for dy := -1; dy <= 1; dy++ {
				if dx == 0 && dy == 0 {
					continue
				}
				input := labFixture(weapon.ID, 5)
				for i := range input.Board {
					input.Board[i].Height = 0
					input.Board[i].Object = models.ObjectNone
					input.Board[i].Terrain = models.TerrainGrass
				}
				input.Tokens[1].Position = models.Position{X: input.Tokens[0].X + dx, Y: input.Tokens[0].Y + dy}
				_, err := ResolveSkillsLabAttack(input)
				if (err != nil) != (weapon.ID == models.WeaponSniperID) {
					t.Fatalf("%s at %d,%d: %v", weapon.ID, dx, dy, err)
				}
			}
		}
	}
}

func TestLabSniperRespectsRelativeElevation(t *testing.T) {
	for _, heights := range [][2]int{{0, 1}, {1, 2}, {1, 1}, {2, 2}, {2, 1}, {1, 0}} {
		input := labFixture(models.WeaponSniperID, 6)
		input.Board[5*15+4].Height = heights[0]
		input.Board[5*15+6].Height = heights[1]
		before := append([]models.SkillToken(nil), input.Tokens...)
		_, err := ResolveSkillsLabAttack(input)
		if (err != nil) != (heights[1] > heights[0]) {
			t.Fatalf("incorrect sniper elevation rule for %v: %v", heights, err)
		}
		if !reflect.DeepEqual(before, input.Tokens) {
			t.Fatal("lab attack mutated input")
		}
	}
}
