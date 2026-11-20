import { useEffect, useRef, useState } from "react";
import { useTick } from "@pixi/react";
import { Assets, type Sprite, type Texture } from "pixi.js";
import frame1 from "@/assets/Explosion_3/Explosion_1.png";
import frame2 from "@/assets/Explosion_3/Explosion_2.png";
import frame3 from "@/assets/Explosion_3/Explosion_3.png";
import frame4 from "@/assets/Explosion_3/Explosion_4.png";
import frame5 from "@/assets/Explosion_3/Explosion_5.png";
import frame6 from "@/assets/Explosion_3/Explosion_6.png";
import frame7 from "@/assets/Explosion_3/Explosion_7.png";
import frame8 from "@/assets/Explosion_3/Explosion_8.png";
import frame9 from "@/assets/Explosion_3/Explosion_9.png";
import frame10 from "@/assets/Explosion_3/Explosion_10.png";
import { hasBridge, LEVEL_HEIGHT, TILE_WIDTH, type MapTile } from "./generator";
import { projectPosition, type MapView } from "./projection";
import { grenadeFrame, GRENADE_FRAME_COUNT, GRENADE_FRAME_MS } from "./grenadeAnimation";

const SOURCES = [frame1, frame2, frame3, frame4, frame5, frame6, frame7, frame8, frame9, frame10];

export default function GrenadeExplosion({ tile, view, onComplete, onError }: {
  tile: MapTile; view: MapView; onComplete: () => void; onError: () => void;
}) {
  const sprite = useRef<Sprite>(null);
  const clock = useRef({elapsed: 0, completed: false});
  const [frames, setFrames] = useState<Texture[] | null>(null);
  const complete = useRef(onComplete);
  useEffect(() => { complete.current = onComplete; }, [onComplete]);

  useEffect(() => {
    let active = true;
    Promise.all(SOURCES.map((source) => Assets.load<Texture>(source))).then((textures) => {
      if (active) setFrames(textures);
    }).catch(() => { if (active) { onError(); complete.current(); } });
    return () => { active = false; }; // Shared cached textures belong to Assets.
  }, [onError]);

  useTick((ticker) => {
    if (!sprite.current || !frames || clock.current.completed) return;
    const state = clock.current;
    const frame = grenadeFrame(state.elapsed);
    if (frame === null) {
      state.completed = true;
      sprite.current.visible = false;
      onComplete();
      return;
    }
    sprite.current.texture = frames[frame];
    const duration = GRENADE_FRAME_COUNT * GRENADE_FRAME_MS;
    sprite.current.alpha = Math.min(1, (duration - state.elapsed) / (GRENADE_FRAME_MS * 2));
    state.elapsed += Math.min(ticker.deltaMS, 50);
  });

  const position = projectPosition(tile, view);
  return frames ? <pixiSprite ref={sprite} texture={frames[0]}
    x={position.x} y={position.y - tile.height * LEVEL_HEIGHT - 42 - (hasBridge(tile) ? 8 : 0)}
    anchor={{x:0.5,y:0.82}} width={TILE_WIDTH*3} height={TILE_WIDTH*3}
    zIndex={100000} eventMode="none" /> : null;
}
