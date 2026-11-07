package game

import (
	"github.com/DaniloReisRocha/grid-tactics/backend/internal/utils/array"
	"github.com/DaniloReisRocha/grid-tactics/backend/models"
	"github.com/gofrs/uuid"
	"github.com/pkg/errors"
)

func moveUnitByID(gs *GameSession, accoID uuid.UUID, unitID string, to models.Position) error {
	player, ok := gs.Players[accoID]
	if !ok {
		return errors.New("player not found")
	}
	unitIndex, ok := array.FindIndex(player.Units, func(p models.UnitInstance) bool {
		return p.Unit.ID == unitID
	})
	if !ok {
		return errors.New("unit not found")
	}
	unit := player.Units[unitIndex]
	path := findPathToTarget(gs.Board, unit.X, unit.Y, to.X, to.Y)
	if path == nil {
		return errors.New("invalid path")
	}
	totalCost := 0
	for _, step := range path {
		cell, ok := findCell(gs.Board, step.X, step.Y)
		if !ok {
			return errors.New("invalid cell in path")
		}
		totalCost += 1 + int(cell.Elevation)
	}

	if totalCost > unit.CurrentStamina {
		return errors.New("not enough stamina")
	}

	unit.X = to.X
	unit.Y = to.Y
	unit.CurrentStamina -= totalCost
	gs.Players[accoID].Units[unitIndex] = unit

	return nil
}
