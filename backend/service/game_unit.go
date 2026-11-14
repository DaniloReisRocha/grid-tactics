package service

import (
	"fmt"

	"github.com/DaniloReisRocha/grid-tactics/backend/models"
	"github.com/gofrs/uuid"
	"github.com/jmoiron/sqlx"
)

// Starter roster for the prototype. Account persistence can implement this service later.
type GameUnitsGetService struct{ DB *sqlx.DB }

func (g *GameUnitsGetService) Run(accoID uuid.UUID) ([]models.Unit, error) {
	pistol, ok := models.WeaponByID(models.WeaponPistolID)
	if !ok {
		return nil, fmt.Errorf("arma inicial não encontrada: %s", models.WeaponPistolID)
	}
	icons := []string{"🦸‍♂️", "🦸‍♀️", "🧙‍♂️"}
	units := make([]models.Unit, len(icons))
	for i, icon := range icons {
		units[i] = models.Unit{
			ID:   fmt.Sprintf("%s-%d", accoID.String(), i+1),
			Icon: icon, LP: 3, Shield: 3, Stamina: 4,
			Weapons: []models.Weapon{pistol},
			Tools:   []string{},
		}
	}
	return units, nil
}
