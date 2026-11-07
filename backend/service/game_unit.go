package service

import (
	"github.com/DaniloReisRocha/grid-tactics/backend/models"
	"github.com/gofrs/uuid"
	"github.com/jmoiron/sqlx"
)

type GameUnitsGetService struct {
	DB *sqlx.DB
}

func (g *GameUnitsGetService) Run(accoID uuid.UUID) ([]models.Unit, error) {
	return getGameUnits(g.DB, accoID)
}

func getGameUnits(db *sqlx.DB, accoID uuid.UUID) ([]models.Unit, error) {

	gamePlayers := map[string][]models.Unit{
		"98cad047-c065-4011-bce0-bf7558bb54b9": {
			{
				ID:      "11",
				Icon:    "🦸‍♂️",
				LP:      3,
				Shield:  3,
				Stamina: 4,
				Weapons: []string{"pistol"},
				Tools:   []string{},
			},
			{
				ID:      "12",
				Icon:    "🦸‍♀️",
				LP:      3,
				Shield:  3,
				Stamina: 4,
				Weapons: []string{"pistol"},
				Tools:   []string{},
			},
			{
				ID:      "13",
				Icon:    "🧙‍♂️",
				LP:      3,
				Shield:  3,
				Stamina: 4,
				Weapons: []string{"pistol"},
				Tools:   []string{},
			},
		},
		"fcfc0336-c317-41e7-b7d5-6b2377938bd3": {
			{
				ID:      "21",
				Icon:    "🦹‍♀️",
				LP:      3,
				Shield:  3,
				Stamina: 4,
				Weapons: []string{"pistol"},
				Tools:   []string{},
			},
			{
				ID:      "22",
				Icon:    "🦹‍♂️",
				LP:      3,
				Shield:  3,
				Stamina: 4,
				Weapons: []string{"pistol"},
				Tools:   []string{},
			},
			{
				ID:      "23",
				Icon:    "🧛",
				LP:      3,
				Shield:  3,
				Stamina: 4,
				Weapons: []string{"pistol"},
				Tools:   []string{},
			},
		},
	}
	if units, ok := gamePlayers[accoID.String()]; ok {
		return units, nil
	}
	return []models.Unit{}, nil
}
