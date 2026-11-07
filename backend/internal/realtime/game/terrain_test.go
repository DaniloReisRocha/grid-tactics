package game

import (
	"crypto/sha256"
	"fmt"
	"strings"
	"testing"

	"github.com/DaniloReisRocha/grid-tactics/backend/models"
)

func TestGeneratorMatchesReferenceMaps(t *testing.T) {
	// Golden hashes preserve compatibility with the original procedural generator.
	for seed, expected := range map[string]string{
		"grid-tactics":          "5223e6f053fedd338ddbfbd6d7dde04bcf32cd229a43b760f2e5825c572b5873",
		"integration":           "7dec4a3541602c0dad2319b9b0578e8a1ff731747f7aa5dcfb8f63422308be25",
		"water-check-7":         "1899291ffff3e7a061fee7dddf75c9d7558de329bbc592ea0d7a8771303e24fd",
		"cidade-a\u00e7\u00e3o": "dbbcede79589fea0b04c5ef0b5ade1e3b3abd60c04a99bbf56bbf1f4004e5f1f",
	} {
		var data strings.Builder
		for _, cell := range generateBoard(seed, 15, 15) {
			fall := 0
			if cell.Waterfall {
				fall = 1
			}
			fmt.Fprintf(&data, "%d,%d:%s:%d:%s:%s:%d\n", cell.X, cell.Y, cell.Terrain, cell.Height, cell.Object, cell.Roof, fall)
		}
		if actual := fmt.Sprintf("%x", sha256.Sum256([]byte(data.String()))); actual != expected {
			t.Errorf("%s map differs from example: %s", seed, actual)
		}
	}
}

func TestTerrainTraversalAndSpawn(t *testing.T) {
	grass := models.Cell{X: 0, Y: 0, Terrain: "grass"}
	water := models.Cell{X: 1, Y: 0, Terrain: "water"}
	if CanTraverse(grass, water) {
		t.Fatal("water accepted")
	}
	water.Object = "bridgeX"
	if !CanTraverse(grass, water) {
		t.Fatal("bridge rejected")
	}
	water.Object = "bridgeY"
	if CanTraverse(grass, water) {
		t.Fatal("bridge entered sideways")
	}
	cliff := models.Cell{X: 1, Y: 0, Terrain: "grass", Height: 2}
	if !CanTraverse(grass, cliff) {
		t.Fatal("multi-level climb rejected despite stamina-based movement")
	}
	cliff.Height = 1
	cliff.Object = "forest"
	if CanTraverse(grass, cliff) {
		t.Fatal("forest accepted")
	}
}
