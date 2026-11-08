import { TILE_HEIGHT, TILE_WIDTH } from "./generator";

export interface MapView {
  width: number;
  height: number;
  reversed: boolean;
}

// Transform only the view. Tiles and server commands retain their world coordinates.
export function viewPosition(point: { x: number; y: number }, view: MapView) {
  return view.reversed
    ? { x: view.width - 1 - point.x, y: view.height - 1 - point.y }
    : { x: point.x, y: point.y };
}

export function projectPosition(point: { x: number; y: number }, view: MapView) {
  const { x, y } = viewPosition(point, view);
  return { x: (x - y) * TILE_WIDTH / 2, y: (x + y) * TILE_HEIGHT / 2 };
}

export function viewDepth(point: { x: number; y: number }, view: MapView) {
  const { x, y } = viewPosition(point, view);
  return (x + y) * 10 + y * 0.01;
}
