package service

import (
	"encoding/json"
	"github.com/DaniloReisRocha/grid-tactics/backend/models"
	"github.com/gofrs/uuid"
	"testing"
)

func testTeam() []models.UnitLoadout {
	return []models.UnitLoadout{
		{WeaponID: models.WeaponPistolID, AbilityIDs: []string{models.AbilityShieldID, models.AbilityTrampolineID}},
		{WeaponID: models.WeaponSniperID, AbilityIDs: []string{models.AbilityGrenadeID}},
		{WeaponID: models.WeaponWarHammerID, AbilityIDs: []string{models.AbilityBarricadeID}},
	}
}

func TestTeamUsesServerCatalogAndDoesNotMutateStarterRoster(t *testing.T) {
	units, err := (&GameUnitsGetService{}).Run(uuid.Must(uuid.NewV4()))
	if err != nil {
		t.Fatal(err)
	}
	before, _ := json.Marshal(units)
	team := testTeam()
	result, err := ApplyTeamLoadout(units, team)
	if err != nil {
		t.Fatal(err)
	}
	for i, unit := range result {
		if unit.ID != units[i].ID || unit.LP != 3 || unit.Shield != 3 || unit.Stamina != models.UnitBaseStamina || unit.Weapons[0].ID != team[i].WeaponID || len(unit.Abilities) != len(team[i].AbilityIDs) {
			t.Fatalf("invalid loadout: %+v", unit)
		}
		weapon, _ := models.WeaponByID(team[i].WeaponID)
		if unit.Weapons[0].Damage != weapon.Damage || unit.EffectiveStamina() != models.UnitBaseStamina-weapon.StaminaPenalty {
			t.Fatal("weapon data did not come from catalog")
		}
	}
	after, _ := json.Marshal(units)
	if string(before) != string(after) {
		t.Fatal("team selection mutated starter roster")
	}
}

func TestInvalidTeamsAreRejected(t *testing.T) {
	units, _ := (&GameUnitsGetService{}).Run(uuid.Must(uuid.NewV4()))
	for _, name := range []string{"too few", "too many", "unknown weapon", "unknown ability", "duplicate ability", "weapon penalty", "too many skills"} {
		t.Run(name, func(t *testing.T) {
			team := testTeam()
			switch name {
			case "too few":
				team = team[:2]
			case "too many":
				team = append(team, team[0])
			case "unknown weapon":
				team[0].WeaponID = "invalid"
			case "unknown ability":
				team[0].AbilityIDs = []string{"invalid"}
			case "duplicate ability":
				team[0].AbilityIDs = []string{models.AbilityShieldID, models.AbilityShieldID}
			case "weapon penalty":
				team[1].AbilityIDs = []string{models.AbilityShieldID, models.AbilityGrenadeID}
			case "too many skills":
				team[0].AbilityIDs = []string{models.AbilityShieldID, models.AbilityGrenadeID, models.AbilityBarricadeID}
			}
			if _, err := ApplyTeamLoadout(units, team); err == nil {
				t.Fatal("invalid loadout accepted")
			}
		})
	}
}
