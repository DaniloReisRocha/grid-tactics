import { useEffect, useRef, useState } from "react";
import { useTick } from "@pixi/react";
import { Assets, type Sprite, type Texture } from "pixi.js";
import frontUrl from "@/assets/skulls/light outline skulls1.png";
import sideUrl from "@/assets/skulls/light outline skulls2.png";
import backUrl from "@/assets/skulls/light outline skulls3.png";
import otherSideUrl from "@/assets/skulls/light outline skulls4.png";
import { hasBridge, LEVEL_HEIGHT, type MapTile } from "./generator";
import { projectPosition, viewDepth, type MapView } from "./projection";

const FRAME_MS = 300;
const SOURCES = [frontUrl, sideUrl, backUrl, otherSideUrl];

export default function DefeatedUnit({ tile, view, onError }: {
  tile: MapTile;
  view: MapView;
  onError: () => void;
}) {
  const sprite = useRef<Sprite>(null);
  const elapsed = useRef(0);
  const [frames, setFrames] = useState<Texture[] | null>(null);

  useEffect(() => {
    let active = true;
    Promise.all(SOURCES.map((source) => Assets.load<Texture>(source))).then((textures) => {
      if (!active) return;
      for (const texture of textures) texture.source.scaleMode = "nearest";
      setFrames(textures);
    }).catch(() => { if (active) onError(); });
    // Assets owns these shared textures, including when several units die.
    return () => { active = false; };
  }, [onError]);

  useTick((ticker) => {
    if (!sprite.current || !frames) return;
    elapsed.current = (elapsed.current + ticker.deltaMS) % (frames.length * FRAME_MS);
    const offset = view.reversed ? 2 : 0;
    sprite.current.texture = frames[(Math.floor(elapsed.current / FRAME_MS) + offset) % frames.length];
  });

  const position = projectPosition(tile, view);
  return frames ? (
    <pixiSprite
      ref={sprite}
      texture={frames[view.reversed ? 2 : 0]}
      x={position.x}
      y={position.y - tile.height * LEVEL_HEIGHT - 42 - (hasBridge(tile) ? 8 : 0)}
      zIndex={viewDepth(tile, view) + 4}
      anchor={{ x: 0.5, y: 1 }}
      scale={2}
      roundPixels
      eventMode="none"
    />
  ) : null;
}
