package game

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/DaniloReisRocha/grid-tactics/backend/models"
)

func TestOrbMilestonesAreSeededDistinctAndNeverRepeated(t *testing.T) {
	positions := map[models.Position]bool{}
	for seed := 0; seed < 12; seed++ {
		var expected []models.Orb
		for repeat := 0; repeat < 2; repeat++ {
			s, a, b := roundSession(t)
			s.ID = fmt.Sprintf("orb-seed-%d", seed)
			s.spawnMilestoneOrbs()
			if len(s.Orbs) != 0 {
				t.Fatal("orb appeared with six living units")
			}
			s.Players[a].Units[0].CurrentLife = 0
			s.spawnMilestoneOrbs()
			if len(s.Orbs) != 0 {
				t.Fatal("orb appeared after only one defeat")
			}
			s.Players[b].Units[0].CurrentLife = 0
			s.spawnMilestoneOrbs()
			if len(s.Orbs) != 1 || s.Orbs[0].ID != "orb-4" {
				t.Fatal("missing four-unit milestone")
			}
			s.Orbs[0].Triggered = true
			s.Players[a].Units[1].CurrentLife = 0
			s.spawnMilestoneOrbs()
			if len(s.Orbs) != 1 {
				t.Fatal("orb appeared at three living units")
			}
			s.Players[b].Units[1].CurrentLife = 0
			s.spawnMilestoneOrbs()
			if len(s.Orbs) != 2 || s.Orbs[0].Position == s.Orbs[1].Position {
				t.Fatal("second milestone reused the first orb position")
			}
			for _, orb := range s.Orbs {
				for _, player := range s.Players {
					for _, unit := range player.Units {
						if orb.Position == unit.Position {
							t.Fatal("spawned under unit or skull")
						}
					}
				}
				for _, cell := range s.Board {
					if cell.X == orb.X && cell.Y == orb.Y && (cell.Terrain == models.TerrainWater || cell.Object != models.ObjectNone) {
						t.Fatal("spawned on blocked terrain")
					}
				}
			}
			before := append([]models.Orb(nil), s.Orbs...)
			s.spawnMilestoneOrbs()
			if !reflect.DeepEqual(before, s.Orbs) {
				t.Fatal("milestone repeated")
			}
			if repeat == 0 {
				expected = before
				positions[s.Orbs[0].Position] = true
			} else if !reflect.DeepEqual(expected, s.Orbs) {
				t.Fatal("seed was not deterministic")
			}
		}
	}
	if len(positions) < 2 {
		t.Fatal("different seeds never changed orb position")
	}
}

func TestMovementCollectsIntermediateOrbOnlyOnce(t *testing.T) {
	s, a, b := startedSession(t)
	unit := &s.Players[a].Units[0]
	unit.Position = models.Position{}
	s.Players[b].Units[0].Position = models.Position{X: 8, Y: 8}
	shield, _ := models.AbilityByID(models.AbilityShieldID)
	unit.Unit.Abilities = []models.Ability{shield}
	unit.AbilityUses = map[string]int{models.AbilityGrenadeID: 1, models.AbilityBarricadeID: 3}
	unit.HasAttacked = true
	unit.TurnCount = 3
	s.Orbs = []models.Orb{{ID: "orb-4", Position: models.Position{X: 1}}}
	if err := s.MovePiece(a, "a", models.Position{X: 3}); err != nil {
		t.Fatal(err)
	}
	if !s.Orbs[0].Triggered || s.Orbs[0].CollectedBy != "a" || len(unit.AbilityUses) != 0 || unit.CurrentShield != 3 {
		t.Fatal("orb was not collected along the route")
	}
	if unit.CurrentLife != 3 || unit.CurrentStamina != 3 || !unit.HasAttacked || unit.TurnCount != 3 {
		t.Fatal("orb changed unrelated resources")
	}
	unit.CurrentShield = 0
	unit.AbilityUses = map[string]int{models.AbilityGrenadeID: 1}
	if err := s.MovePiece(a, "a", models.Position{}); err != nil {
		t.Fatal(err)
	}
	if unit.CurrentShield != 0 || unit.AbilityUses[models.AbilityGrenadeID] != 1 {
		t.Fatal("collected orb was reusable")
	}
	if err := s.AddPlayer(nil, a, []models.Unit{testUnit("replacement")}); err != nil {
		t.Fatal(err)
	}
	if !s.Orbs[0].Triggered || unit.CurrentShield != 0 {
		t.Fatal("reconnect revived collected orb")
	}
}

func TestJumpCollectsLandingOrbAndKnockbackCollectsAlongPath(t *testing.T) {
	s, a, b := startedSession(t)
	unit := &s.Players[a].Units[0]
	unit.Position = models.Position{}
	s.Players[b].Units[0].Position = models.Position{X: 8, Y: 8}
	jump, _ := models.AbilityByID(models.AbilityTrampolineID)
	unit.Unit.Abilities = []models.Ability{jump}
	s.Orbs = []models.Orb{{ID: "middle", Position: models.Position{X: 1}}, {ID: "landing", Position: models.Position{X: 2}}}
	if err := s.UseAbility(a, "a", jump.ID, models.Position{X: 2}); err != nil {
		t.Fatal(err)
	}
	if s.Orbs[0].Triggered || !s.Orbs[1].Triggered || unit.AbilityUses[jump.ID] != 0 {
		t.Fatal("jump did not collect only landing orb or renew skill")
	}
	s, a, b = startedSession(t)
	attacker, defender := &s.Players[a].Units[0], &s.Players[b].Units[0]
	attacker.Position = models.Position{}
	hammer, _ := models.WeaponByID(models.WeaponWarHammerID)
	attacker.Unit.Weapons = []models.Weapon{hammer}
	defender.Position = models.Position{X: 1}
	defender.CurrentLife = 30
	defender.AbilityUses = map[string]int{jump.ID: 1}
	s.Orbs = []models.Orb{{ID: "orb-4", Position: models.Position{X: 2}}}
	if err := s.Attack(a, "a", defender.Position); err != nil {
		t.Fatal(err)
	}
	if !s.Orbs[0].Triggered || s.Orbs[0].CollectedBy != "b" || len(defender.AbilityUses) != 0 {
		t.Fatal("forced movement did not collect orb")
	}
}

func TestGrenadeCrossesBothOrbMilestonesInOneAction(t *testing.T) {
	s, a, b := roundSession(t)
	positions := []models.Position{{X: 5, Y: 5}, {X: 5, Y: 4}, {X: 6, Y: 4}, {X: 7, Y: 4}, {X: 7, Y: 5}, {X: 8, Y: 8}}
	index := 0
	for _, owner := range s.TurnOrder {
		for i := range s.Players[owner].Units {
			unit := &s.Players[owner].Units[i]
			unit.Position = positions[index]
			unit.CurrentLife = 2
			index++
		}
	}
	actor := &s.Players[a].Units[0]
	actor.CurrentLife = 3
	s.Players[b].Units[2].CurrentLife = 3
	grenade, _ := models.AbilityByID(models.AbilityGrenadeID)
	actor.Unit.Abilities = []models.Ability{grenade}
	s.CurrentTurn = &a
	s.CurrentUnitID = actor.Unit.ID
	if err := s.UseAbility(a, actor.Unit.ID, grenade.ID, models.Position{X: 6, Y: 5}); err != nil {
		t.Fatal(err)
	}
	if s.Status != models.GameStatusPlaying || len(s.Orbs) != 2 || s.Orbs[0].Position == s.Orbs[1].Position {
		t.Fatal("multi-kill skipped milestones")
	}
}
