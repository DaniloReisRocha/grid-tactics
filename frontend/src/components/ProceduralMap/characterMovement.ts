import { hasBridge, type MapTile } from "./generator";

export const tileKey = (x: number, y: number) => `${x},${y}`;

export function isWalkable(tile: MapTile) {
  return hasBridge(tile) || (tile.terrain !== "water" && !tile.object);
}

export function canStep(from: MapTile, to: MapTile | undefined) {
  if (!to || !isWalkable(from) || !isWalkable(to)) return false;
  const dx = Math.abs(to.x - from.x);
  const dy = Math.abs(to.y - from.y);
  if (Math.max(dx, dy) !== 1 || Math.abs(to.height - from.height) > 1) return false;
  for (const tile of [from, to]) {
    if (tile.object === "bridgeX" && dy !== 0) return false;
    if (tile.object === "bridgeY" && dx !== 0) return false;
  }
  return true;
}

export function getDestination(from: MapTile, dx: number, dy: number, lookup: Map<string, MapTile>) {
  const to = lookup.get(tileKey(from.x + dx, from.y + dy));
  if (!canStep(from, to)) return undefined;
  // Diagonal movement cannot cut through the corner of an obstacle or cliff.
  if (dx && dy && (
    !canStep(from, lookup.get(tileKey(from.x + dx, from.y))) ||
    !canStep(from, lookup.get(tileKey(from.x, from.y + dy)))
  )) return undefined;
  return to;
}

/** Shortest route in tile steps, including start and destination; null if unreachable. */
export function findPath(from: MapTile, to: MapTile, lookup: Map<string, MapTile>): MapTile[] | null {
  if (!isWalkable(from) || !isWalkable(to)) return null;
  const destinationKey = tileKey(to.x, to.y);
  const originKey = tileKey(from.x, from.y);
  const parents = new Map<string, MapTile | null>([[originKey, null]]);
  const queue = [from];
  for (let index = 0; index < queue.length; index += 1) {
    const current = queue[index];
    if (tileKey(current.x, current.y) === destinationKey) {
      const route: MapTile[] = [];
      let step: MapTile | null = current;
      while (step) {
        route.push(step);
        step = parents.get(tileKey(step.x, step.y)) ?? null;
      }
      return route.reverse();
    }
    for (const [dx, dy] of [[1, 0], [0, 1], [-1, 0], [0, -1], [1, 1], [-1, 1], [1, -1], [-1, -1]]) {
      const next = getDestination(current, dx, dy, lookup);
      if (!next) continue;
      const key = tileKey(next.x, next.y);
      if (parents.has(key)) continue;
      parents.set(key, current);
      queue.push(next);
    }
  }
  return null;
}

export function findSpawn(tiles: readonly MapTile[]) {
  const lookup = new Map(tiles.map((tile) => [tileKey(tile.x, tile.y), tile]));
  const visited = new Set<string>();
  let largest: MapTile[] = [];
  for (const tile of tiles) {
    const key = tileKey(tile.x, tile.y);
    if (visited.has(key) || !isWalkable(tile)) continue;
    const region = [tile];
    visited.add(key);
    for (let index = 0; index < region.length; index += 1) {
      const current = region[index];
      for (const [dx, dy] of [[1, 0], [-1, 0], [0, 1], [0, -1]]) {
        const next = getDestination(current, dx, dy, lookup);
        if (next && !visited.has(tileKey(next.x, next.y))) {
          visited.add(tileKey(next.x, next.y));
          region.push(next);
        }
      }
    }
    if (region.length > largest.length) largest = region;
  }
  return largest[Math.floor(largest.length / 2)];
}
