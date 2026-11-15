package service

import (
	"fmt"

	"github.com/DaniloReisRocha/grid-tactics/backend/models"
)

func TeamOptions() models.TeamOptions {
	return models.TeamOptions{Weapons: models.WeaponOptions(), Abilities: models.AbilityOptions(), UnitCount: models.TeamUnitCount, BaseLife: models.UnitBaseLife, ShieldCapacity: models.UnitShieldCapacity, BaseStamina: models.UnitBaseStamina, ItemSlots: models.UnitItemSlots, AbilitySlots: models.UnitAbilitySlots}
}

func ApplyTeamLoadout(units []models.Unit, team []models.UnitLoadout) ([]models.Unit, error) {
	if len(team) != models.TeamUnitCount || len(units) != models.TeamUnitCount {
		return nil, fmt.Errorf("a equipe deve possuir exatamente %d unidades", models.TeamUnitCount)
	}
	result := append([]models.Unit(nil), units...)
	for i, loadout := range team {
		weapon, ok := models.WeaponByID(loadout.WeaponID)
		if !ok {
			return nil, fmt.Errorf("unidade %d: arma inválida", i+1)
		}
		unit := &result[i]
		unit.Weapons = []models.Weapon{weapon}
		unit.Abilities = []models.Ability{}
		if len(loadout.AbilityIDs) > models.UnitAbilitySlots {
			return nil, fmt.Errorf("unidade %d: quantidade de habilidades inválida", i+1)
		}
		seen := map[string]bool{}
		for _, id := range loadout.AbilityIDs {
			if id == "" {
				continue
			}
			ability, ok := models.AbilityByID(id)
			if !ok || seen[id] {
				return nil, fmt.Errorf("unidade %d: habilidade inválida ou repetida", i+1)
			}
			seen[id] = true
			unit.Abilities = append(unit.Abilities, ability)
		}
		if len(unit.Abilities) > unit.EffectiveAbilitySlots() {
			return nil, fmt.Errorf("unidade %d: esta arma permite até %d habilidade(s)", i+1, unit.EffectiveAbilitySlots())
		}
	}
	return result, nil
}
