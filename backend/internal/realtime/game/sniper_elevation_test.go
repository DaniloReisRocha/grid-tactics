package game

import (
	"testing"

	"github.com/DaniloReisRocha/grid-tactics/backend/models"
)

func TestSniperOnlyDamagesTargetsAtSameOrLowerHeight(t *testing.T) {
	for sourceHeight := 0; sourceHeight <= 3; sourceHeight++ {
		for targetHeight := 0; targetHeight <= 3; targetHeight++ {
			s, a, b := startedSession(t)
			attacker, defender := &s.Players[a].Units[0], &s.Players[b].Units[0]
			attacker.Position = models.Position{X: 4, Y: 4}
			defender.Position = models.Position{X: 6, Y: 4}
			defender.CurrentLife = 30
			sniper, _ := models.WeaponByID(models.WeaponSniperID)
			attacker.Unit.Weapons = []models.Weapon{sniper}
			attacker.TurnCount = 2
			s.Board[44].Height = sourceHeight
			s.Board[46].Height = targetHeight
			revision := s.Revision
			err := s.Attack(a, "a", defender.Position)
			if targetHeight > sourceHeight {
				if err == nil || attacker.HasAttacked || defender.CurrentLife != 30 || s.Revision != revision {
					t.Fatal("higher target was damaged or consumed attack")
				}
			} else if err != nil || defender.CurrentLife != 30-sniper.Damage {
				t.Fatalf("sniper rejected height %d -> %d: %v", sourceHeight, targetHeight, err)
			}
		}
	}
	// The height restriction follows its reusable effect, independently of IDs.
	effect, _ := models.EffectByID(models.EffectNoHigherTargetsID)
	if !effect.Negative || ValidateWeaponEffects(models.Weapon{ID: models.WeaponPistolID, Effect: []models.Effect{effect}}, 2, 0, 1) == nil {
		t.Fatal("height restriction is not reusable")
	}
	if err := ValidateWeaponEffects(models.Weapon{ID: models.WeaponSniperID}, 2, 0, 1); err != nil {
		t.Fatal("restriction applied without its effect")
	}
}
