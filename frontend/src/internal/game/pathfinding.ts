import type { Cell, Position, UnitInstance } from "@/api/models";
import { getDestination, movementCost } from "@/components/ProceduralMap/characterMovement";

export const positionKey = ({ x, y }: Position) => `${x},${y}`;

/** Matches the server's Dijkstra search, terrain costs, occupancy and diagonals. */
export function movementOptions(board: Cell[], unit: UnitInstance, units: UnitInstance[]) {
  const cells = new Map(board.map((cell) => [positionKey(cell), cell]));
  const blocked = new Set(units.filter((other) => other.currentLife > 0 && other.unit.id !== unit.unit.id).map(positionKey));
  const origin = positionKey(unit);
  const costs = new Map<string, number>([[origin, 0]]);
  const paths = new Map<string, Position[]>([[origin, []]]);
  const queue = [{ x: unit.x, y: unit.y, cost: 0 }];
  while (queue.length) {
    queue.sort((a, b) => a.cost - b.cost);
    const current = queue.shift()!;
    const key = positionKey(current);
    if (current.cost !== costs.get(key)) continue;
    for (const [dx, dy] of [[1, 0], [0, 1], [-1, 0], [0, -1], [1, 1], [-1, 1], [1, -1], [-1, -1]]) {
      const next = { x: current.x + dx, y: current.y + dy };
      const nextKey = positionKey(next);
      const from = cells.get(key);
      if (!from) continue;
      const cell = getDestination(from, dx, dy, cells);
      if (!cell || blocked.has(nextKey)) continue;
      if (dx && dy) {
        const sideX = positionKey({ x: next.x, y: current.y });
        const sideY = positionKey({ x: current.x, y: next.y });
        if (!cells.has(sideX) || !cells.has(sideY) || blocked.has(sideX) || blocked.has(sideY)) continue;
      }
      const cost = current.cost + movementCost(from, cell);
      if (cost > unit.currentStamina || cost >= (costs.get(nextKey) ?? Infinity)) continue;
      costs.set(nextKey, cost);
      paths.set(nextKey, [...paths.get(key)!, next]);
      queue.push({ ...next, cost });
    }
  }
  return { paths, costs };
}
