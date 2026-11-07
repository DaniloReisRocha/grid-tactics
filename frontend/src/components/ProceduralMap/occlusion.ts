export const OCCLUDED_OPACITY = 0.2;

export interface ScreenBounds {
  minX: number;
  minY: number;
  maxX: number;
  maxY: number;
}

export function occludesUnit(
  unit: ScreenBounds,
  unitDepth: number,
  structure: ScreenBounds,
  structureDepth: number,
) {
  return structureDepth > unitDepth
    && structure.minX < unit.maxX && structure.maxX > unit.minX
    && structure.minY < unit.maxY && structure.maxY > unit.minY;
}

export function approachOpacity(current: number, target: number, elapsedMS: number) {
  const next = current + (target - current) * (1 - Math.exp(-Math.max(0, elapsedMS) / 85));
  return Math.abs(next - target) < 0.002 ? target : next;
}
