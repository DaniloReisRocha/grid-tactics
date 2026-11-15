import { useEffect, useRef, useState } from "react";
import { useTick } from "@pixi/react";
import { Assets, type Sprite, type Texture } from "pixi.js";
import frame2 from "@/assets/Explosion_5/Explosion_2.png";
import frame3 from "@/assets/Explosion_5/Explosion_3.png";
import frame4 from "@/assets/Explosion_5/Explosion_4.png";
import frame5 from "@/assets/Explosion_5/Explosion_5.png";
import frame6 from "@/assets/Explosion_5/Explosion_6.png";
import frame7 from "@/assets/Explosion_5/Explosion_7.png";
import frame8 from "@/assets/Explosion_5/Explosion_8.png";
import frame9 from "@/assets/Explosion_5/Explosion_9.png";
import { hasBridge, LEVEL_HEIGHT, type MapTile } from "./generator";
import { projectPosition, viewDepth, type MapView } from "./projection";
import { orbFrame } from "./orbAnimation";

const SOURCES = [frame2, frame3, frame4, frame5, frame6, frame7, frame8, frame9];

export default function Orb({ tile, triggered, view, onComplete, onError }: {
  tile: MapTile; triggered: boolean; view: MapView;
  onComplete: () => void; onError: () => void;
}) {
  const sprite = useRef<Sprite>(null);
  const clock = useRef({ elapsed: 0, triggered, completed: false });
  const [frames, setFrames] = useState<Texture[] | null>(null);

  useEffect(() => {
    let active = true;
    Promise.all(SOURCES.map((source) => Assets.load<Texture>(source))).then((textures) => {
      if (active) setFrames(textures);
    }).catch(() => { if (active) onError(); });
    return () => { active = false; }; // Assets owns the shared frame textures.
  }, [onError]);

  useTick((ticker) => {
    if (!sprite.current || !frames || clock.current.completed) return;
    const state = clock.current;
    if (state.triggered !== triggered) { state.triggered = triggered; state.elapsed = 0; }
    const frame = orbFrame(state.elapsed, triggered);
    if (frame === null) {
      state.completed = true;
      sprite.current.visible = false;
      onComplete();
      return;
    }
    sprite.current.texture = frames[frame - 2];
    state.elapsed += Math.min(ticker.deltaMS, 50);
  });

  const position = projectPosition(tile, view);
  return frames ? <pixiSprite ref={sprite} texture={frames[triggered ? 3 : 0]}
    x={position.x} y={position.y - tile.height * LEVEL_HEIGHT - 58 - (hasBridge(tile) ? 8 : 0)}
    zIndex={viewDepth(tile, view) + (triggered ? 15 : 4)}
    anchor={0.5} scale={0.15} eventMode="none" /> : null;
}
