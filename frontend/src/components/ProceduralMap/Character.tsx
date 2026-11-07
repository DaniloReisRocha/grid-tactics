import { useCallback, useEffect, useRef, useState } from "react";
import { useTick } from "@pixi/react";
import { Assets, Rectangle, Sprite, Texture } from "pixi.js";
import sheetUrl from "@/assets/Small-8-Direction-Characters_by_AxulArt/Small-8-Direction-Characters_by_AxulArt.png";
import { LEVEL_HEIGHT, TILE_HEIGHT, TILE_WIDTH, hasBridge, type MapTile } from "./generator";
import type { CharacterModel } from "./types";

interface CharacterProps {
  spawn: MapTile;
  path: readonly MapTile[];
  onError: () => void;
  model?: CharacterModel;
  onSelect?: () => void;
  onHover?: (hovered: boolean) => void;
  onSprite?: (sprite: Sprite | null) => void;
}

const FRAME_WIDTH = 16;
const FRAME_HEIGHT = 24;
const STEP_MS = 180;

function project(tile: MapTile) {
  return {
    x: (tile.x - tile.y) * TILE_WIDTH / 2,
    // Ground's top-face center is 42 pixels above its image's bottom anchor.
    y: (tile.x + tile.y) * TILE_HEIGHT / 2 - tile.height * LEVEL_HEIGHT - 42
      + (hasBridge(tile) ? -8 : 0),
  };
}

export default function Character({ spawn, path, onError, model = 2, onSelect, onHover, onSprite }: CharacterProps) {
  const spriteRef = useRef<Sprite>(null);
  const registerSprite = useCallback((sprite: Sprite | null) => {
    spriteRef.current = sprite;
    onSprite?.(sprite);
  }, [onSprite]);
  const motion = useRef({ from: spawn, to: spawn, elapsed: STEP_MS, direction: 4, clock: 0 });
  const command = useRef<readonly MapTile[]>(path);
  const pending = useRef<MapTile[]>([]);
  const [frames, setFrames] = useState<Texture[][] | null>(null);

  useEffect(() => {
    let active = true;
    let ownedFrames: Texture[][] = [];
    Assets.load<Texture>(sheetUrl).then((sheet) => {
      if (!active) return;
      sheet.source.scaleMode = "nearest";
      ownedFrames = Array.from({ length: 8 }, (_, direction) =>
        Array.from({ length: 3 }, (_, pose) => new Texture({
          source: sheet.source,
          // The first 24px row of each block contains arrows, not character art.
          frame: new Rectangle(direction * FRAME_WIDTH, model * 96 + (pose + 1) * FRAME_HEIGHT, FRAME_WIDTH, FRAME_HEIGHT),
        })),
      );
      setFrames(ownedFrames);
    }).catch(() => { if (active) onError(); });
    return () => {
      active = false;
      for (const direction of ownedFrames) for (const texture of direction) texture.destroy(false);
    };
  }, [onError, model]);

  useTick((ticker) => {
    const sprite = spriteRef.current;
    if (!sprite || !frames) return;
    const state = motion.current;
    const delta = Math.min(ticker.deltaMS, 50);
    if (command.current !== path) {
      command.current = path;
      // A new click replaces future steps; finish the tile transition in progress.
      const source = path[0];
      if (!source || (source.x === state.to.x && source.y === state.to.y)) {
        pending.current = path.slice(1);
      } else {
        // Several server-confirmed moves can arrive before the previous animation ends.
        const anchor = pending.current.findIndex((tile) => tile.x === source.x && tile.y === source.y);
        if (anchor >= 0) pending.current = [...pending.current.slice(0, anchor + 1), ...path.slice(1)];
        else { state.from = source; state.to = source; state.elapsed = STEP_MS; pending.current = path.slice(1); }
      }
    }
    state.clock += delta;
    state.elapsed = Math.min(STEP_MS, state.elapsed + delta);
    if (state.elapsed >= STEP_MS) {
      const next = pending.current.shift();
      if (next) {
        const dx = next.x - state.to.x;
        const dy = next.y - state.to.y;
        const sx = dx - dy;
        const sy = dx + dy;
        // Columns are N, NE, E, SE, S, SW, W, NW, matching screen direction.
        state.direction = (Math.round(Math.atan2(sx, -sy) / (Math.PI / 4)) + 8) % 8;
        state.from = state.to;
        state.to = next;
        state.elapsed = 0;
      }
    }
    const progress = state.elapsed / STEP_MS;
    const from = project(state.from);
    const to = project(state.to);
    sprite.position.set(from.x + (to.x - from.x) * progress, from.y + (to.y - from.y) * progress);
    // Interleave the character with terrain columns and their objects.
    const depth = (tile: MapTile) => (tile.x + tile.y) * 10 + tile.y * 0.01;
    sprite.zIndex = depth(state.from) + (depth(state.to) - depth(state.from)) * progress + 5;
    const pose = progress < 1 ? [1, 0, 2, 0][Math.floor(state.clock / 110) % 4] : 0;
    sprite.texture = frames[state.direction][pose];
  });

  const initial = project(spawn);
  return frames ? (
    <pixiSprite
      ref={registerSprite}
      texture={frames[4][0]}
      x={initial.x} y={initial.y}
      anchor={{ x: 0.5, y: 0.875 }}
      scale={2}
      roundPixels
      eventMode={onSelect ? "static" : "none"}
      cursor={onSelect ? "pointer" : "default"}
      onPointerTap={onSelect}
      onPointerOver={() => onHover?.(true)}
      onPointerOut={() => onHover?.(false)}
    />
  ) : null;
}
