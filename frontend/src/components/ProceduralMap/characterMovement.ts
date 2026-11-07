import { hasBridge, type MapTile } from "./generator";

export const tileKey = (x: number, y: number) => `${x},${y}`;
const PASSABLE_DECORATIONS = new Set<MapTile["object"]>(["tree", "forest", "rock", "rockDirt"]);

export function isWalkable(tile: MapTile) {
  return hasBridge(tile) || (tile.terrain !== "water" && (!tile.object || PASSABLE_DECORATIONS.has(tile.object)));
}

export function movementCost(from: Pick<MapTile, "height">, to: Pick<MapTile, "height">) {
  return 1 + Math.min(2, Math.max(0, to.height - from.height));
}

export function canStep(from: MapTile, to: MapTile | undefined) {
  if (!to || !isWalkable(from) || !isWalkable(to)) return false;
  const dx = Math.abs(to.x - from.x);
  const dy = Math.abs(to.y - from.y);
  if (Math.max(dx, dy) !== 1) return false;
  for (const tile of [from, to]) {
    if (tile.object === "bridgeX" && dy !== 0) return false;
    if (tile.object === "bridgeY" && dx !== 0) return false;
  }
  return true;
}

export function getDestination(from: MapTile, dx: number, dy: number, lookup: Map<string, MapTile>) {
  const to = lookup.get(tileKey(from.x + dx, from.y + dy));
  if (!canStep(from, to)) return undefined;
  // Diagonal movement cannot cut through the corner of an obstacle.
  if (dx && dy && (
    !canStep(from, lookup.get(tileKey(from.x + dx, from.y))) ||
    !canStep(from, lookup.get(tileKey(from.x, from.y + dy)))
  )) return undefined;
  return to;
}
