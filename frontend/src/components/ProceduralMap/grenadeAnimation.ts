export const GRENADE_FRAME_MS = 90;
export const GRENADE_FRAME_COUNT = 10;

export function grenadeFrame(elapsedMS: number) {
  const index = Math.floor(Math.max(0, elapsedMS) / GRENADE_FRAME_MS);
  return index < GRENADE_FRAME_COUNT ? index : null;
}
