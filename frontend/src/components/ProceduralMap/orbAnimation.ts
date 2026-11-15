const IDLE_FRAMES = [2, 3, 4, 3];
const IDLE_FRAME_MS = 180;
const BURST_FRAME_MS = 120;

export function orbFrame(elapsed: number, triggered: boolean): number | null {
  if (!triggered) return IDLE_FRAMES[Math.floor(elapsed / IDLE_FRAME_MS) % IDLE_FRAMES.length];
  const index = Math.floor(elapsed / BURST_FRAME_MS);
  return index < 5 ? 5 + index : null;
}
