package models

// Terrain identifiers shared by map generation, movement and spawn rules.
// Keep their string values stable: they are also consumed by the frontend.
const (
	TerrainGrass = "grass"
	TerrainDirt  = "dirt"
	TerrainWater = "water"
)

const (
	ObjectNone     = ""
	ObjectTree     = "tree"
	ObjectForest   = "forest"
	ObjectRock     = "rock"
	ObjectRockDirt = "rockDirt"
	ObjectHouse    = "house"
	ObjectRuin     = "ruin"
	ObjectBridgeX  = "bridgeX"
	ObjectBridgeY  = "bridgeY"
)

const (
	RoofGable  = "roofGable"
	RoofGreen  = "roofGreen"
	RoofPurple = "roofPurple"
)
