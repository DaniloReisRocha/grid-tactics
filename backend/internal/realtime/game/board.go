package game

import (
	"github.com/DaniloReisRocha/grid-tactics/backend/models"
	"math"
	"sort"
	"unicode/utf16"
)

// Authoritative procedural generation. Reference hashes in terrain_test.go
// preserve the terrain distribution and features across refactors.
type terrainTile struct {
	models.Cell
	noise    float64
	reserved bool
}

func hasBridgeWaterNeighbors(tile *terrainTile, at func(int, int) *terrainTile) bool {
	if tile == nil || tile.Terrain != models.TerrainWater {
		return false
	}
	waterNeighbors := 0
	for _, direction := range [][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}} {
		neighbor := at(tile.X+direction[0], tile.Y+direction[1])
		if neighbor != nil && neighbor.Terrain == models.TerrainWater {
			waterNeighbors++
		}
	}
	return waterNeighbors >= 2
}

func seedHash(seed string) uint32 {
	if seed == "" {
		seed = "grid-tactics"
	}
	hash := uint32(2166136261)
	for _, code := range utf16.Encode([]rune(seed)) {
		hash ^= uint32(code)
		hash *= 16777619
	}
	return hash
}
func randomAt(seed uint32, x, y int) float64 {
	value := seed ^ uint32(x)*374761393 ^ uint32(y)*668265263
	value = (value ^ (value >> 13)) * 1274126177
	return float64(value^(value>>16)) / 4294967295
}
func valueNoise(seed uint32, x, y float64) float64 {
	left, top := int(math.Floor(x)), int(math.Floor(y))
	smooth := func(v float64) float64 { return v * v * (3 - 2*v) }
	ox, oy := smooth(x-float64(left)), smooth(y-float64(top))
	a, b := randomAt(seed, left, top), randomAt(seed, left+1, top)
	c, d := randomAt(seed, left, top+1), randomAt(seed, left+1, top+1)
	upper, lower := a+(b-a)*ox, c+(d-c)*ox
	return upper + (lower-upper)*oy
}
func fractalNoise(seed uint32, x, y int) float64 {
	frequency, amplitude, value, total := 0.12, 1.0, 0.0, 0.0
	for octave := 0; octave < 4; octave++ {
		value += valueNoise(seed+uint32(octave*1013), float64(x)*frequency, float64(y)*frequency) * amplitude
		total += amplitude
		frequency *= 2
		amplitude *= 0.5
	}
	return value / total
}
func generateBoard(sid string, width, height int) []models.Cell {
	seed := seedHash(sid)
	tiles := make([]terrainTile, width*height)
	at := func(x, y int) *terrainTile {
		if x < 0 || y < 0 || x >= width || y >= height {
			return nil
		}
		return &tiles[y*width+x]
	}
	roll := func(feature int) float64 { return randomAt(seed^0x85ebca6b, feature, 0) }
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			edge := min(x, y, width-1-x, height-1-y)
			mask := math.Min(1, float64(edge)/(float64(min(width, height))/2))
			noise := fractalNoise(seed, x, y) - (1-mask)*0.2
			terrain := models.TerrainGrass
			level := max(0, min(3, int(math.Floor((noise-0.48)*10))))
			if noise < 0.36 {
				terrain = models.TerrainWater
				level = 0
			} else if noise < 0.45 {
				terrain = models.TerrainDirt
			}
			tiles[y*width+x] = terrainTile{Cell: models.Cell{X: x, Y: y, Terrain: terrain, Height: level}, noise: noise}
		}
	}
	if width >= 12 && height >= 10 && roll(1) < 0.58 {
		fx := 5 + int(math.Floor(roll(2)*float64(width-10)))
		fy := 4 + int(math.Floor(roll(3)*float64(height-8)))
		for dy := -3; dy <= 3; dy++ {
			for dx := -5; dx <= 0; dx++ {
				tile := at(fx+dx, fy+dy)
				if tile == nil {
					continue
				}
				tile.Terrain = models.TerrainGrass
				if abs(dy) == 3 || dx == -5 {
					tile.Terrain = models.TerrainDirt
				}
				tile.Height = 1
				if abs(dy) <= 2 && dx >= -4 {
					tile.Height = 2
				}
				tile.reserved = true
			}
		}
		for dx := -3; dx <= 0; dx++ {
			tile := at(fx+dx, fy)
			tile.Terrain = models.TerrainWater
			tile.Height = 2
			tile.Waterfall = dx == 0
		}
		for dx := 1; dx <= 4; dx++ {
			for dy := -2; dy <= 2; dy++ {
				tile := at(fx+dx, fy+dy)
				if tile == nil {
					continue
				}
				tile.Terrain = models.TerrainDirt
				if abs(dy) <= 1 && dx <= 3 {
					tile.Terrain = models.TerrainWater
				}
				tile.Height = 0
				tile.reserved = true
			}
		}
	}
	waterBudget := int(math.Floor(float64(len(tiles)) * 0.2))
	water := 0
	for _, tile := range tiles {
		if tile.Terrain == models.TerrainWater {
			water++
		}
	}
	excess := water > waterBudget
	bridgeLimit := 1 + int(math.Floor(roll(8)*math.Min(4, math.Ceil(float64(len(tiles))/150))))
	addBridges := func() {
		type candidate struct {
			start, end *terrainTile
			span       []*terrainTile
			object     string
		}
		candidates := []candidate{}
		for i := range tiles {
			start := &tiles[i]
			if start.Terrain == models.TerrainWater || start.reserved {
				continue
			}
			for _, direction := range [][2]int{{1, 0}, {0, 1}} {
				dx, dy := direction[0], direction[1]
				span := []*terrainTile{}
				next := at(start.X+dx, start.Y+dy)
				for next != nil && next.Terrain == models.TerrainWater && !next.reserved && next.Height == start.Height && len(span) < 4 {
					span = append(span, next)
					next = at(next.X+dx, next.Y+dy)
				}
				if len(span) > 0 && next != nil && next.Terrain != models.TerrainWater && !next.reserved && next.Height == start.Height {
					valid := true
					for _, tile := range span {
						if !hasBridgeWaterNeighbors(tile, at) {
							valid = false
							break
						}
					}
					if !valid {
						continue
					}
					object := models.ObjectBridgeY
					if dx != 0 {
						object = models.ObjectBridgeX
					}
					candidates = append(candidates, candidate{start, next, span, object})
				}
			}
		}
		sort.SliceStable(candidates, func(i, j int) bool {
			a, b := candidates[i].start, candidates[j].start
			return randomAt(seed^1103, a.X, a.Y) < randomAt(seed^1103, b.X, b.Y)
		})
		count := 0
		for _, c := range candidates {
			if count >= bridgeLimit {
				break
			}
			blocked := c.start.reserved || c.end.reserved
			for _, tile := range c.span {
				blocked = blocked || tile.reserved
			}
			if blocked {
				continue
			}
			for _, tile := range c.span {
				tile.Object = c.object
				tile.reserved = true
			}
			c.start.Terrain = models.TerrainDirt
			c.end.Terrain = models.TerrainDirt
			c.start.reserved = true
			c.end.reserved = true
			count++
		}
	}
	uncovered := 0
	reclaim := []*terrainTile{}
	for i := range tiles {
		tile := &tiles[i]
		if tile.Terrain == models.TerrainWater && !isBridge(tile.Cell) {
			uncovered++
			if !tile.reserved {
				reclaim = append(reclaim, tile)
			}
		}
	}
	sort.SliceStable(reclaim, func(i, j int) bool {
		a, b := reclaim[i], reclaim[j]
		if a.noise != b.noise {
			return a.noise > b.noise
		}
		if a.Y != b.Y {
			return a.Y < b.Y
		}
		return a.X < b.X
	})
	for _, tile := range reclaim {
		if uncovered <= waterBudget {
			break
		}
		tile.Terrain = models.TerrainDirt
		tile.Height = 0
		uncovered--
	}
	// Place bridges only after reclaiming excess water, so their water neighbors
	// cannot subsequently become dirt and leave an isolated bridge tile.
	if excess {
		addBridges()
	}
	houses := []*terrainTile{}
	limit := 1 + int(math.Floor(roll(5)*4))
	if roll(4) < 0.65 {
		candidates := []*terrainTile{}
		for i := range tiles {
			tile := &tiles[i]
			if tile.Terrain != models.TerrainWater && !tile.reserved {
				candidates = append(candidates, tile)
			}
		}
		sort.SliceStable(candidates, func(i, j int) bool {
			a, b := candidates[i], candidates[j]
			return randomAt(seed^901, a.X, a.Y) < randomAt(seed^901, b.X, b.Y)
		})
		for _, candidate := range candidates {
			if len(houses) >= limit {
				break
			}
			near := false
			for _, house := range houses {
				if math.Hypot(float64(house.X-candidate.X), float64(house.Y-candidate.Y)) < 5 {
					near = true
				}
			}
			if near {
				continue
			}
			clearing := []*terrainTile{}
			for dy := -1; dy <= 1; dy++ {
				for dx := -1; dx <= 1; dx++ {
					tile := at(candidate.X+dx, candidate.Y+dy)
					if tile != nil && !tile.reserved && tile.Terrain != models.TerrainWater {
						clearing = append(clearing, tile)
					}
				}
			}
			if len(clearing) != 9 {
				continue
			}
			level := 1
			for _, tile := range clearing {
				level = max(level, tile.Height)
			}
			for _, tile := range clearing {
				tile.Height = level
				tile.Terrain = models.TerrainGrass
				tile.reserved = true
			}
			candidate.Terrain = models.TerrainDirt
			candidate.Object = models.ObjectHouse
			candidate.Roof = []string{models.RoofGable, models.RoofGreen, models.RoofPurple}[int(math.Floor(randomAt(seed^902, candidate.X, candidate.Y)*3))%3]
			houses = append(houses, candidate)
		}
	}
	for i := range tiles {
		tile := &tiles[i]
		if tile.reserved || tile.Terrain == models.TerrainWater {
			continue
		}
		detail := randomAt(seed^0x9e3779b9, tile.X, tile.Y)
		forest := fractalNoise(seed^0x27d4eb2d, tile.X, tile.Y)
		if roll(6) < 0.72 && forest > 0.53 && detail > 0.18 {
			tile.Object = models.ObjectForest
		} else if detail > 0.91 || (tile.Height >= 2 && detail > 0.65) {
			tile.Object = models.ObjectRock
			if tile.Terrain == models.TerrainDirt {
				tile.Object = models.ObjectRockDirt
			}
		} else if roll(7) < 0.4 && detail < 0.018 && tile.Height > 0 {
			tile.Object = models.ObjectRuin
		} else if detail > 0.77 {
			tile.Object = models.ObjectTree
		}
	}
	board := make([]models.Cell, len(tiles))
	for i, tile := range tiles {
		tile.Elevation = models.ElevationLevel(tile.Height)
		board[i] = tile.Cell
	}
	// Small disconnected pockets must not strand either team at spawn.
	if len(largestRegion(board)) < 6 {
		for y := 0; y < height-2; y++ {
			for x := 0; x < width-2; x++ {
				safe := true
				for dy := 0; dy < 3; dy++ {
					for dx := 0; dx < 3; dx++ {
						if at(x+dx, y+dy).reserved {
							safe = false
						}
					}
				}
				if !safe {
					continue
				}
				for dy := 0; dy < 3; dy++ {
					for dx := 0; dx < 3; dx++ {
						index := (y+dy)*width + x + dx
						board[index] = models.Cell{X: x + dx, Y: y + dy, Terrain: models.TerrainGrass}
					}
				}
				return board
			}
		}
	}
	return board
}
