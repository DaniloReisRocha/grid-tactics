package game

import (
	"github.com/DaniloReisRocha/grid-tactics/backend/models"
	"math/rand"
)

func generateBoard(sid string, elevationPercent float64) []models.Cell {
	if elevationPercent < 0 || elevationPercent > 1 {
		elevationPercent = 0.3
	}

	source := rand.NewSource(hashSid(sid))
	r := rand.New(source)

	cells := make([]models.Cell, 100)
	for y := 0; y < 10; y++ {
		for x := 0; x < 10; x++ {
			cells[y*10+x] = models.Cell{X: x, Y: y, Elevation: models.ElevationLow}
		}
	}
	totalElevated := int(float64(100) * elevationPercent)
	mediumCount := r.Intn(totalElevated + 1)
	highCount := totalElevated - mediumCount

	fillElevation(cells, r, models.ElevationMedium, mediumCount)
	fillElevation(cells, r, models.ElevationHigh, highCount)

	return cells
}

func fillElevation(cells []models.Cell, r *rand.Rand, elevation models.ElevationLevel, count int) {
	if count <= 0 {
		return
	}
	positions := r.Perm(len(cells))
	placed := 0

	for _, pos := range positions {
		if cells[pos].Elevation == models.ElevationLow {
			cells[pos].Elevation = elevation
			placed++
			if placed == count {
				return
			}
		}
	}
}

func hashSid(sid string) int64 {
	var hash int64
	for _, c := range sid {
		hash = hash*31 + int64(c)
	}
	return hash
}
