package game

import (
	"testing"

	"github.com/DaniloReisRocha/grid-tactics/backend/models"
)

func combinedEffectWeapon() models.Weapon {
	// Deliberately reuse all effects on a pistol: behavior must follow effects,
	// rather than require the original shotgun/sniper/hammer weapon IDs.
	weapon := models.Weapon{ID: models.WeaponPistolID, Damage: 1, Range: 3}
	for _, id := range []string{models.EffectKnockbackID, models.EffectShieldBreakID, models.EffectNoAdjacentID, models.EffectSameLevelID} {
		effect, _ := models.EffectByID(id)
		weapon.Effect = append(weapon.Effect, effect)
	}
	return weapon
}

func TestEffectsCanBeCombinedAndReusedByAnotherWeapon(t *testing.T) {
	weapon := combinedEffectWeapon()
	if ValidateWeaponEffects(weapon, 1, 0, 0) == nil {
		t.Fatal("negative adjacency effect ignored")
	}
	if ValidateWeaponEffects(weapon, 2, 0, 1) == nil {
		t.Fatal("negative height effect ignored")
	}
	if err := ValidateWeaponEffects(weapon, 2, 1, 1); err != nil {
		t.Fatal(err)
	}
	life, shield := ApplyWeaponDamage(weapon, 6, 3)
	if life != 6 || shield != 0 {
		t.Fatal("reused shield break not applied")
	}
	withoutEffects := weapon
	withoutEffects.ID = models.WeaponShotgunID
	withoutEffects.Effect = nil
	_, shield = ApplyWeaponDamage(withoutEffects, 6, 3)
	if shield != 2 {
		t.Fatal("shield break still tied to shotgun ID")
	}
	if err := ValidateWeaponEffects(withoutEffects, 1, 0, 1); err != nil {
		t.Fatal("restrictions applied without effects")
	}
	cells := map[models.Position]models.Cell{
		{X: 2}: {X: 2, Terrain: models.TerrainGrass}, {X: 3}: {X: 3, Terrain: models.TerrainGrass},
	}
	if next, ok := WeaponKnockback(weapon, models.Position{}, models.Position{X: 2}, cells, nil); !ok || next.X != 3 {
		t.Fatal("reused knockback not applied")
	}
	if _, ok := WeaponKnockback(withoutEffects, models.Position{}, models.Position{X: 2}, cells, nil); ok {
		t.Fatal("knockback applied without effect")
	}
}

func TestMultiplayerUsesReusableEffects(t *testing.T) {
	s, a, b := startedSession(t)
	attacker, defender := &s.Players[a].Units[0], &s.Players[b].Units[0]
	attacker.Position = models.Position{}
	attacker.Unit.Weapons = []models.Weapon{combinedEffectWeapon()}
	attacker.CurrentStamina = 0
	defender.Position, defender.CurrentLife, defender.CurrentShield = models.Position{X: 1}, 6, 3
	if s.Attack(a, "a", defender.Position) == nil || attacker.HasAttacked || defender.CurrentShield != 3 {
		t.Fatal("negative effect did not reject attack cleanly")
	}
	defender.Position = models.Position{X: 2}
	if err := s.Attack(a, "a", defender.Position); err != nil {
		t.Fatal(err)
	}
	if defender.X != 3 || defender.CurrentShield != 0 || defender.CurrentLife != 6 || len(defender.Path) != 2 {
		t.Fatalf("effects not applied to multiplayer: %+v", defender)
	}
	if !attacker.HasAttacked || attacker.CurrentStamina != 0 {
		t.Fatal("effects changed attack or stamina rules")
	}
}
