package game

import (
	"testing"

	"github.com/DaniloReisRocha/grid-tactics/backend/models"
)

func TestReloadEffectAlternatesOwnTurnsAndSurvivesReconnect(t *testing.T) {
	for _, weaponID := range []string{models.WeaponSniperID, "reused-effect"} {
		t.Run(weaponID, func(t *testing.T) {
			s, a, b := startedSession(t)
			attacker, defender := &s.Players[a].Units[0], &s.Players[b].Units[0]
			weapon, _ := models.WeaponByID(models.WeaponSniperID)
			weapon.ID = weaponID // Rules follow the effect, not the weapon ID.
			attacker.Unit.Weapons = []models.Weapon{weapon}
			attacker.TurnCount = 1
			attacker.Position = models.Position{}
			defender.Position, defender.CurrentLife = models.Position{X: 4}, 30
			for turn := 1; turn <= 6; turn++ {
				if attacker.TurnCount != turn {
					t.Fatal("incorrect own-turn counter")
				}
				life, revision := defender.CurrentLife, s.Revision
				err := s.Attack(a, "a", defender.Position)
				if turn%2 == 1 {
					if err == nil || attacker.HasAttacked || defender.CurrentLife != life || s.Revision != revision {
						t.Fatal("reload did not reject attack atomically")
					}
					if turn == 1 {
						if err := s.MovePiece(a, "a", models.Position{Y: 1}); err != nil {
							t.Fatal("reloading blocked movement", err)
						}
						attacker.Position = models.Position{}
						barricade, _ := models.AbilityByID(models.AbilityBarricadeID)
						attacker.Unit.Abilities = []models.Ability{barricade}
						if err := s.UseAbility(a, "a", barricade.ID, models.Position{Y: 1}); err != nil {
							t.Fatal("reloading blocked ability", err)
						}
					}
				} else {
					if err != nil || defender.CurrentLife != life-weapon.Damage || !attacker.HasAttacked {
						t.Fatal("ready turn could not attack", err)
					}
					if s.Attack(a, "a", defender.Position) == nil {
						t.Fatal("attacked twice on ready turn")
					}
				}
				if err := s.AddPlayer(nil, a, []models.Unit{testUnit("replacement")}); err != nil {
					t.Fatal(err)
				}
				if attacker.TurnCount != turn || attacker.Unit.Weapons[0].Reloading(attacker.TurnCount) != (turn%2 == 1) {
					t.Fatal("rejoin reset reload")
				}
				if err := s.EndTurn(a); err != nil {
					t.Fatal(err)
				}
				if attacker.TurnCount != turn {
					t.Fatal("opponent turn advanced reload")
				}
				if err := s.EndTurn(b); err != nil {
					t.Fatal(err)
				}
			}
		})
	}
}

func TestSkippingReadyAttackStillAdvancesReloadCycle(t *testing.T) {
	s, a, b := startedSession(t)
	unit := &s.Players[a].Units[0]
	weapon, _ := models.WeaponByID(models.WeaponSniperID)
	unit.Unit.Weapons, unit.TurnCount = []models.Weapon{weapon}, 1
	for i := 0; i < 2; i++ {
		if err := s.EndTurn(a); err != nil {
			t.Fatal(err)
		}
		if err := s.EndTurn(b); err != nil {
			t.Fatal(err)
		}
	}
	if unit.TurnCount != 3 || !weapon.Reloading(unit.TurnCount) {
		t.Fatal("unused attack bypassed reload turn")
	}
	weapon.Effect = nil
	if weapon.Reloading(1) {
		t.Fatal("weapon without reload effect was restricted")
	}
}
