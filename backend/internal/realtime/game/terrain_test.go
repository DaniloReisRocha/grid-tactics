package game

import (
	"crypto/sha256"
	"fmt"
	"strings"
	"testing"

	"github.com/DaniloReisRocha/grid-tactics/backend/models"
)

func TestGeneratorMatchesReferenceMaps(t *testing.T) {
	// Golden hashes detect unintended changes to the current generation rules.
	for seed, expected := range map[string]string{
		"grid-tactics":          "60582b1ce9c77d3546e357709e327c151e6333c0ff44c431e9cb3a9227a767ed",
		"integration":           "fe93b43e0375d096374c3ef0d981657cffe0c496c580f0f2539843a2f6a469fc",
		"water-check-7":         "0f7567668cb8d1a2bb9be0a419f50de60ac3527a5ebfb2990b676bed2201c51c",
		"cidade-a\u00e7\u00e3o": "2b309dbe1c307633078e6ef11c6dc7f533ff96d17dac7fd6f2a74e44327bc488",
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
			t.Errorf("%s map differs from reference: %s", seed, actual)
		}
	}
}

func TestBridgeRequiresTwoCardinalWaterNeighbors(t *testing.T) {
	for _, test := range []struct {
		name      string
		neighbors []models.Position
		want      bool
	}{
		{"isolated puddle", nil, false},
		{"one water side", []models.Position{{X: 1}}, false},
		{"diagonals do not count", []models.Position{{X: 1, Y: 1}, {X: -1, Y: -1}}, false},
		{"opposite water sides", []models.Position{{X: 1}, {X: -1}}, true},
		{"adjacent water sides", []models.Position{{X: 1}, {Y: 1}}, true},
	} {
		t.Run(test.name, func(t *testing.T) {
			center := &terrainTile{Cell: models.Cell{Terrain: models.TerrainWater}}
			cells := map[models.Position]*terrainTile{{}: center}
			for _, position := range test.neighbors {
				cells[position] = &terrainTile{Cell: models.Cell{X: position.X, Y: position.Y, Terrain: models.TerrainWater}}
			}
			at := func(x, y int) *terrainTile { return cells[models.Position{X: x, Y: y}] }
			if got := hasBridgeWaterNeighbors(center, at); got != test.want {
				t.Fatalf("got %v, want %v", got, test.want)
			}
		})
	}
}

func TestGeneratedBridgesKeepWaterNeighborsAndBudget(t *testing.T) {
	bridges := 0
	for seed := 0; seed < 500; seed++ {
		board := generateBoard(fmt.Sprintf("bridge-check-%d", seed), 15, 15)
		cells := make(map[models.Position]models.Cell, len(board))
		for _, cell := range board {
			cells[models.Position{X: cell.X, Y: cell.Y}] = cell
		}
		unbridgedWater := 0
		for _, cell := range board {
			if !isBridge(cell) {
				if cell.Terrain == models.TerrainWater {
					unbridgedWater++
				}
				continue
			}
			bridges++
			waterNeighbors := 0
			for _, d := range [][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}} {
				if cells[models.Position{X: cell.X + d[0], Y: cell.Y + d[1]}].Terrain == models.TerrainWater {
					waterNeighbors++
				}
			}
			if cell.Terrain != models.TerrainWater || waterNeighbors < 2 {
				t.Fatalf("seed %d: invalid bridge at %d,%d with %d water neighbors", seed, cell.X, cell.Y, waterNeighbors)
			}
		}
		if unbridgedWater > 45 {
			t.Fatalf("seed %d: water budget exceeded: %d", seed, unbridgedWater)
		}
	}
	if bridges == 0 {
		t.Fatal("no bridges generated across 500 seeds")
	}
}

func TestTerrainTraversalAndSpawn(t *testing.T) {
	grass := models.Cell{X: 0, Y: 0, Terrain: models.TerrainGrass}
	water := models.Cell{X: 1, Y: 0, Terrain: models.TerrainWater}
	if CanTraverse(grass, water) {
		t.Fatal("water accepted")
	}
	water.Object = models.ObjectBridgeX
	if !CanTraverse(grass, water) {
		t.Fatal("bridge rejected")
	}
	water.Object = models.ObjectBridgeY
	if CanTraverse(grass, water) {
		t.Fatal("bridge entered sideways")
	}
	cliff := models.Cell{X: 1, Y: 0, Terrain: models.TerrainGrass, Height: 2}
	if !CanTraverse(grass, cliff) {
		t.Fatal("multi-level climb rejected despite stamina-based movement")
	}
	cliff.Height = 1
	cliff.Object = models.ObjectForest
	if !CanTraverse(grass, cliff) {
		t.Fatal("forest should be traversable")
	}
}

func TestDecorationsArePassableButBuildingsAndWaterStillBlock(t *testing.T) {
	from := models.Cell{X: 0, Y: 0, Terrain: models.TerrainGrass}
	for _, object := range []string{models.ObjectTree, models.ObjectForest, models.ObjectRock, models.ObjectRockDirt} {
		to := models.Cell{X: 1, Y: 0, Terrain: models.TerrainGrass, Object: object}
		if !CanTraverse(from, to) || !CanTraverse(to, from) {
			t.Errorf("cannot enter/leave %s", object)
		}
		to.Terrain = models.TerrainWater
		if CanTraverse(from, to) {
			t.Errorf("decoration %s made water traversable", object)
		}
	}
	for _, object := range []string{models.ObjectHouse, models.ObjectRuin, "unknown"} {
		if CanTraverse(from, models.Cell{X: 1, Y: 0, Terrain: models.TerrainGrass, Object: object}) {
			t.Errorf("blocking structure %s became passable", object)
		}
	}
	s, a, b := startedSession(t)
	s.Board = nil
	for x, object := range []string{"", models.ObjectTree, models.ObjectForest, models.ObjectRock, models.ObjectRockDirt, "", ""} {
		s.Board = append(s.Board, models.Cell{X: x, Y: 0, Terrain: models.TerrainGrass, Object: object})
	}
	s.Players[a].Units[0].Position = models.Position{}
	s.Players[b].Units[0].Position = models.Position{X: 6}
	if err := s.MovePiece(a, "a", models.Position{X: 5}); err != nil {
		t.Fatal(err)
	}
	if unit := s.Players[a].Units[0]; unit.X != 5 || unit.CurrentStamina != 1 {
		t.Fatal("crossing decorations did not use normal movement cost")
	}
}
