// Shared terrain catalog, map types and rendering dimensions.
// The painted diamond occupies 224 x 112 pixels inside the 256 x 352 image.
export const TILE_WIDTH = 56;
export const TILE_HEIGHT = 28;
export const SPRITE_SCALE = 0.25;
export const LEVEL_HEIGHT = 32;
export const MAX_UNBRIDGED_WATER_RATIO = 0.2;

export const TERRAIN_ASSETS = {
  water: "/assets/tiles/sketchTown/water_center_E.png",
  dirt: "/assets/tiles/sketchTown/dirt_center_E.png",
  grass: "/assets/tiles/sketchTown/grass_center_E.png",
  tree: "/assets/tiles/sketchTown/tree_single_E.png",
  rock: "/assets/tiles/sketchTown/rocks_grass_E.png",
  rockDirt: "/assets/tiles/sketchTown/rocks_dirt_E.png",
  forest: "/assets/tiles/sketchTown/tree_multiple_E.png",
  house: "/assets/tiles/sketchTown/building_doorWindowsBeige_E.png",
  roofGable: "/assets/tiles/sketchTown/roof_gablePurple_E.png",
  roofGreen: "/assets/tiles/sketchTown/roof_roundGreen_E.png",
  roofPurple: "/assets/tiles/sketchTown/roof_pointPurple_E.png",
  waterfall: "/assets/tiles/sketchTown/water_fall_S.png",
  ruin: "/assets/tiles/sketchTown/structure_low_E.png",
  bridgeX: "/assets/tiles/sketchTown/bridge_E.png",
  bridgeY: "/assets/tiles/sketchTown/bridge_S.png",
} as const;

export type Asset = keyof typeof TERRAIN_ASSETS;
type Terrain = "water" | "dirt" | "grass";
type MapObject = "tree" | "rock" | "rockDirt" | "forest" | "house" | "ruin" | "bridgeX" | "bridgeY";
type Roof = "roofGable" | "roofGreen" | "roofPurple";

// Objects rest on the top face of a cube, rather than replacing the cube.
export const OBJECT_RULES: Record<
  MapObject,
  {
    allowedTerrain: readonly Terrain[];
    surfaceOffset: number;
  }
> = {
  tree: { allowedTerrain: ["grass", "dirt"], surfaceOffset: 30 },
  rock: { allowedTerrain: ["grass", "dirt"], surfaceOffset: 35 },
  rockDirt: { allowedTerrain: ["dirt"], surfaceOffset: 35 },
  forest: { allowedTerrain: ["grass", "dirt"], surfaceOffset: 35 },
  house: { allowedTerrain: ["grass", "dirt"], surfaceOffset: LEVEL_HEIGHT },
  ruin: { allowedTerrain: ["grass", "dirt"], surfaceOffset: LEVEL_HEIGHT },
  bridgeX: { allowedTerrain: ["water"], surfaceOffset: 20 },
  bridgeY: { allowedTerrain: ["water"], surfaceOffset: 20 },
};

export type MapTile = {
  x: number;
  y: number;
  terrain: Terrain;
  object?: MapObject;
  elevation: number;
  height: number;
  roof?: Roof;
  waterfall?: boolean;
};

export const TERRAIN_LABELS: Array<{
  terrain: Asset;
  label: string;
  color: string;
}> = [
  { terrain: "water", label: "Agua", color: "#53a9c7" },
  { terrain: "dirt", label: "Terra", color: "#c98259" },
  { terrain: "grass", label: "Grama", color: "#9ed356" },
  { terrain: "tree", label: "Floresta", color: "#397649" },
  { terrain: "rock", label: "Rochas", color: "#7c847d" },
  { terrain: "forest", label: "Mata densa", color: "#28573b" },
  { terrain: "house", label: "Casas", color: "#d5ac79" },
  { terrain: "waterfall", label: "Cachoeiras", color: "#a5e7f1" },
  { terrain: "ruin", label: "Ruínas", color: "#b67e62" },
  { terrain: "bridgeX", label: "Pontes (tiles)", color: "#d39a6e" },
];

export function hasBridge(tile: MapTile) {
  return tile.object === "bridgeX" || tile.object === "bridgeY";
}
