import { useEffect, useRef, useState } from "react";
import { useTick } from "@pixi/react";
import { Assets, Rectangle, Texture, type Container } from "pixi.js";
import explosionSheetUrl from "@/assets/damage/hit_explosions_set_08.png";
import numberSheetUrl from "@/assets/damage/damage_numbers_style_01.png";
import { hasBridge, LEVEL_HEIGHT, type MapTile } from "./generator";
import { projectPosition, viewDepth, type MapView } from "./projection";

const SHEET_SIZE = 1024;
const GRID_SIZE = 3;
const DURATION_MS = 720;

function frameAt(source: Texture["source"], index: number) {
  const column = index % GRID_SIZE;
  const row = Math.floor(index / GRID_SIZE);
  const left = Math.floor(column * SHEET_SIZE / GRID_SIZE);
  const top = Math.floor(row * SHEET_SIZE / GRID_SIZE);
  const right = Math.floor((column + 1) * SHEET_SIZE / GRID_SIZE);
  const bottom = Math.floor((row + 1) * SHEET_SIZE / GRID_SIZE);
  return new Texture({ source, frame: new Rectangle(left, top, right - left, bottom - top) });
}

export default function DamageIndicator({ amount, tile, view, onError, onComplete }: {
  amount: number;
  tile: MapTile;
  view: MapView;
  onError: () => void;
  onComplete: () => void;
}) {
  const container = useRef<Container>(null);
  const elapsed = useRef(0);
  const completed = useRef(false);
  const [textures, setTextures] = useState<{ explosion: Texture; number: Texture } | null>(null);
  const position = projectPosition(tile, view);
  const groundY = position.y - tile.height * LEVEL_HEIGHT - 42 - (hasBridge(tile) ? 8 : 0);

  useEffect(() => {
    let active = true;
    let owned: Texture[] = [];
    Promise.all([
      Assets.load<Texture>(explosionSheetUrl),
      Assets.load<Texture>(numberSheetUrl),
    ]).then(([explosions, numbers]) => {
      if (!active) return;
      explosions.source.scaleMode = "nearest";
      numbers.source.scaleMode = "nearest";
      const digitIndex = Math.min(9, Math.max(1, Math.round(amount))) - 1;
      // Pick a consistent variation for each damage value while preserving the sheet's 3x3 layout.
      owned = [frameAt(explosions.source, (digitIndex + 1) % 9), frameAt(numbers.source, digitIndex)];
      setTextures({ explosion: owned[0], number: owned[1] });
    }).catch(() => { if (active) onError(); });
    return () => {
      active = false;
      for (const texture of owned) texture.destroy(false);
    };
  }, [amount, onError]);

  useTick((ticker) => {
    const node = container.current;
    if (!node || !textures || completed.current) return;
    elapsed.current = Math.min(DURATION_MS, elapsed.current + Math.min(ticker.deltaMS, 50));
    const progress = elapsed.current / DURATION_MS;
    const impact = Math.min(1, progress / 0.22);
    const fade = progress < 0.58 ? 1 : 1 - (progress - 0.58) / 0.42;
    node.alpha = Math.max(0, fade);
    node.y = groundY - 28 - Math.sin(progress * Math.PI / 2) * 24;
    node.scale.set(0.55 + Math.sin(impact * Math.PI / 2) * 0.45);
    if (progress >= 1) {
      completed.current = true;
      onComplete();
    }
  });

  if (!textures) return null;
  return (
    <pixiContainer
      ref={container}
      x={position.x}
      y={groundY - 28}
      zIndex={viewDepth(tile, view) + 20}
      eventMode="none"
    >
      <pixiSprite
        texture={textures.explosion}
        anchor={0.5}
        scale={0.16}
        roundPixels
      />
      <pixiSprite
        texture={textures.number}
        y={-22}
        anchor={0.5}
        scale={0.095}
        roundPixels
      />
    </pixiContainer>
  );
}
