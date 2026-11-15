package service

import (
	"testing"

	"github.com/DaniloReisRocha/grid-tactics/backend/models"
)

func TestLabBarricadeBlocksThenBreaksWithoutUnitTarget(t *testing.T) {
	input := labFixture(models.WeaponPistolID, 6)
	input.Board[5*15+5].Object = models.ObjectBarricade
	if _, err := ResolveSkillsLabAttack(input); err == nil {
		t.Fatal("shot passed through barricade")
	}
	point := models.Position{X: 5, Y: 5}
	input.Target = &point
	input.TargetID = ""
	result, err := ResolveSkillsLabAttack(input)
	if err != nil {
		t.Fatal(err)
	}
	if result.Board[5*15+5].Object != models.ObjectNone || result.Tokens[1].Shield != 3 || result.Tokens[1].Life != 6 {
		t.Fatal("barricade destruction pierced into target")
	}
	if input.Board[5*15+5].Object != models.ObjectBarricade {
		t.Fatal("lab mutated caller board")
	}
	input.Board = result.Board
	input.Tokens = result.Tokens
	input.Target = nil
	input.TargetID = "target"
	if _, err := ResolveSkillsLabAttack(input); err != nil {
		t.Fatal("cleared attack line still blocked", err)
	}
}

func TestLabCannotDestroyBarricadeThroughAnotherStructure(t *testing.T) {
	for _, object := range []string{models.ObjectHouse, models.ObjectRuin, models.ObjectBarricade} {
		input := labFixture(models.WeaponPistolID, 7)
		input.Board[5*15+5].Object = object
		input.Board[5*15+6].Object = models.ObjectBarricade
		point := models.Position{X: 6, Y: 5}
		input.Target = &point
		input.TargetID = ""
		if _, err := ResolveSkillsLabAttack(input); err == nil {
			t.Fatalf("attack crossed %s", object)
		}
		if input.Board[5*15+6].Object != models.ObjectBarricade {
			t.Fatal("blocked shot destroyed the target")
		}
	}
}
