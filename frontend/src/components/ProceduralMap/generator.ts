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
  reserved?: boolean;
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

function hashSeed(seed: string) {
  let hash = 2166136261;

  for (let index = 0; index < seed.length; index += 1) {
    hash ^= seed.charCodeAt(index);
    hash = Math.imul(hash, 16777619);
  }

  return hash >>> 0;
}

function randomAt(seed: number, x: number, y: number) {
  let value = seed;
  value ^= Math.imul(x, 374761393);
  value ^= Math.imul(y, 668265263);
  value = Math.imul(value ^ (value >>> 13), 1274126177);
  return ((value ^ (value >>> 16)) >>> 0) / 4294967295;
}

function smoothstep(value: number) {
  return value * value * (3 - 2 * value);
}

function valueNoise(seed: number, x: number, y: number) {
  const left = Math.floor(x);
  const top = Math.floor(y);
  const offsetX = smoothstep(x - left);
  const offsetY = smoothstep(y - top);

  const topLeft = randomAt(seed, left, top);
  const topRight = randomAt(seed, left + 1, top);
  const bottomLeft = randomAt(seed, left, top + 1);
  const bottomRight = randomAt(seed, left + 1, top + 1);

  const topValue = topLeft + (topRight - topLeft) * offsetX;
  const bottomValue = bottomLeft + (bottomRight - bottomLeft) * offsetX;

  return topValue + (bottomValue - topValue) * offsetY;
}

function fractalNoise(seed: number, x: number, y: number) {
  let frequency = 0.12;
  let amplitude = 1;
  let value = 0;
  let amplitudeSum = 0;

  for (let octave = 0; octave < 4; octave += 1) {
    value +=
      valueNoise(seed + octave * 1013, x * frequency, y * frequency) *
      amplitude;
    amplitudeSum += amplitude;
    frequency *= 2;
    amplitude *= 0.5;
  }

  return value / amplitudeSum;
}

function getTerrain(elevation: number): Terrain {
  if (elevation < 0.36) return "water";
  if (elevation < 0.45) return "dirt";
  return "grass";
}

export function generateSeededMap(seed: string, width: number, height: number) {
  if (!Number.isSafeInteger(width) || !Number.isSafeInteger(height) || width < 1 || height < 1) {
    throw new RangeError("Map width and height must be positive integers (tile counts).");
  }
  const numericSeed = hashSeed(seed || "grid-tactics");
  const tiles: MapTile[] = [];
  const maxEdgeDistance = Math.min(width, height) / 2;

  for (let y = 0; y < height; y += 1) {
    for (let x = 0; x < width; x += 1) {
      const distanceFromEdge = Math.min(x, y, width - 1 - x, height - 1 - y);
      const islandMask = Math.min(1, distanceFromEdge / maxEdgeDistance);
      const baseElevation = fractalNoise(numericSeed, x, y);
      const elevation = baseElevation - (1 - islandMask) * 0.2;
      const terrain = getTerrain(elevation);

      tiles.push({
        x,
        y,
        elevation,
        terrain,
        height:
          terrain === "water"
            ? 0
            : Math.max(0, Math.min(3, Math.floor((elevation - 0.48) * 10))),
      });
    }
  }

  const tileAt = (x: number, y: number) =>
    x >= 0 && y >= 0 && x < width && y < height
      ? tiles[y * width + x]
      : undefined;
  const featureRoll = (feature: number) =>
    randomAt(numericSeed ^ 0x85ebca6b, feature, 0);

  // A waterfall is a complete geographic feature, never an isolated decoration.
  if (width >= 12 && height >= 10 && featureRoll(1) < 0.58) {
    const fallX = 5 + Math.floor(featureRoll(2) * (width - 10));
    const fallY = 4 + Math.floor(featureRoll(3) * (height - 8));

    // A broad, solid mound with a lower terrace supports the upper reservoir.
    for (let dy = -3; dy <= 3; dy += 1) {
      for (let dx = -5; dx <= 0; dx += 1) {
        const tile = tileAt(fallX + dx, fallY + dy);
        if (!tile) continue;
        const isPlateau = Math.abs(dy) <= 2 && dx >= -4;
        tile.terrain = Math.abs(dy) === 3 || dx === -5 ? "dirt" : "grass";
        tile.height = isPlateau ? 2 : 1;
        tile.reserved = true;
      }
    }

    // Water comes from the upper basin and flows toward +x (the S-facing edge).
    for (let dx = -3; dx <= 0; dx += 1) {
      const tile = tileAt(fallX + dx, fallY)!;
      tile.terrain = "water";
      tile.height = 2;
      if (dx === 0) tile.waterfall = true;
    }

    // The receiving basin is strictly below the cliff and never climbs uphill.
    for (let dx = 1; dx <= 4; dx += 1) {
      for (let dy = -2; dy <= 2; dy += 1) {
        const tile = tileAt(fallX + dx, fallY + dy);
        if (!tile) continue;
        tile.terrain = Math.abs(dy) <= 1 && dx <= 3 ? "water" : "dirt";
        tile.height = 0;
        tile.reserved = true;
      }
    }
  }

  const waterBudget = Math.floor(tiles.length * MAX_UNBRIDGED_WATER_RATIO);
  const hasExcessWater = tiles.filter((tile) => tile.terrain === "water").length > waterBudget;
  const bridgeLimit = 1 + Math.floor(featureRoll(8) * Math.min(4, Math.ceil(tiles.length / 150)));

  const addBridges = () => {
    const candidates: Array<{
      start: MapTile;
      end: MapTile;
      span: MapTile[];
      object: "bridgeX" | "bridgeY";
    }> = [];

    // A bridge must join two dry banks at the same height, across 1–4 water tiles.
    for (const start of tiles) {
      if (start.terrain === "water" || start.reserved) continue;
      for (const [dx, dy] of [[1, 0], [0, 1]]) {
        const span: MapTile[] = [];
        let next = tileAt(start.x + dx, start.y + dy);
        while (next && next.terrain === "water" && !next.reserved
          && next.height === start.height && span.length < 4) {
          span.push(next);
          next = tileAt(next.x + dx, next.y + dy);
        }
        if (span.length && next && next.terrain !== "water" && !next.reserved
          && next.height === start.height) {
          candidates.push({ start, end: next, span, object: dx ? "bridgeX" : "bridgeY" });
        }
      }
    }

    candidates.sort((a, b) =>
      randomAt(numericSeed ^ 1103, a.start.x, a.start.y)
      - randomAt(numericSeed ^ 1103, b.start.x, b.start.y));
    let count = 0;
    for (const { start, end, span, object } of candidates) {
      if (count >= bridgeLimit) break;
      if ([start, end, ...span].some((tile) => tile.reserved)) continue;
      for (const tile of span) {
        tile.object = object;
        tile.reserved = true;
      }
      // Keep approaches clear of houses, trees and rocks.
      start.terrain = end.terrain = "dirt";
      start.reserved = end.reserved = true;
      count += 1;
    }
    return count;
  };

  const bridgesAdded = hasExcessWater ? addBridges() : 0;
  const uncoveredWater = tiles.filter((tile) => tile.terrain === "water" && !hasBridge(tile));
  let excessWater = uncoveredWater.length - waterBudget;

  // Reclaim the shallowest water first, preserving waterfall sources and basins.
  const reclaimableWater = uncoveredWater.filter((tile) => !tile.reserved)
    .sort((a, b) => b.elevation - a.elevation || a.y - b.y || a.x - b.x);
  for (const tile of reclaimableWater) {
    if (excessWater <= 0) break;
    tile.terrain = "dirt";
    tile.height = 0;
    excessWater -= 1;
  }

  // Reclaiming land can create safe banks that did not exist in the raw map.
  if (hasExcessWater && bridgesAdded === 0) addBridges();

  // Houses need a dry, level 3 x 3 clearing, raised at least one cube.
  const housesEnabled = featureRoll(4) < 0.65;
  const houseLimit = 1 + Math.floor(featureRoll(5) * 4);
  const houses: MapTile[] = [];
  const roofs: Roof[] = ["roofGable", "roofGreen", "roofPurple"];
  if (housesEnabled) {
    const candidates = tiles
      .filter((tile) => tile.terrain !== "water" && !tile.reserved)
      .sort(
        (a, b) =>
          randomAt(numericSeed ^ 901, a.x, a.y) -
          randomAt(numericSeed ^ 901, b.x, b.y),
      );

    for (const candidate of candidates) {
      if (houses.length >= houseLimit) break;
      if (
        houses.some(
          (house) =>
            Math.hypot(house.x - candidate.x, house.y - candidate.y) < 5,
        )
      )
        continue;
      const clearing: MapTile[] = [];
      for (let dy = -1; dy <= 1; dy += 1) {
        for (let dx = -1; dx <= 1; dx += 1) {
          const tile = tileAt(candidate.x + dx, candidate.y + dy);
          if (tile && !tile.reserved && tile.terrain !== "water")
            clearing.push(tile);
        }
      }
      if (clearing.length !== 9) continue;
      const height = Math.max(1, ...clearing.map((tile) => tile.height));
      for (const tile of clearing) {
        tile.height = height;
        tile.terrain = "grass";
        tile.reserved = true;
      }
      candidate.terrain = "dirt";
      candidate.object = "house";
      candidate.roof =
        roofs[
          Math.floor(
            randomAt(numericSeed ^ 902, candidate.x, candidate.y) *
              roofs.length,
          ) % roofs.length
        ];
      houses.push(candidate);
    }
  }

  const denseForestsEnabled = featureRoll(6) < 0.72;
  const ruinsEnabled = featureRoll(7) < 0.4;
  for (const tile of tiles) {
    if (tile.reserved || tile.terrain === "water") continue;
    const detail = randomAt(numericSeed ^ 0x9e3779b9, tile.x, tile.y);
    const forestNoise = fractalNoise(numericSeed ^ 0x27d4eb2d, tile.x, tile.y);
    let object: MapObject | undefined;
    if (denseForestsEnabled && forestNoise > 0.53 && detail > 0.18)
      object = "forest";
    else if (detail > 0.91 || (tile.height >= 2 && detail > 0.65)) {
      object = tile.terrain === "dirt" ? "rockDirt" : "rock";
    } else if (ruinsEnabled && detail < 0.018 && tile.height > 0)
      object = "ruin";
    else if (detail > 0.77) object = "tree";

    if (object && OBJECT_RULES[object].allowedTerrain.includes(tile.terrain))
      tile.object = object;
  }

  // Painter's algorithm: tiles farther north are drawn before tiles farther south.
  return tiles.sort((a, b) => a.x + a.y - (b.x + b.y) || a.y - b.y);
}
