import { useEffect, useRef, useState, type RefObject } from "react";
import { useTick } from "@pixi/react";
import { Assets, Rectangle, Sprite, Texture } from "pixi.js";
import arrowsUrl from "@/assets/arrows/direction_arrows_set_01.png";

const BOB_PERIOD_MS = 1400;
const APPEARANCES = {
  turn: { frame: [377, 51, 253, 307], width: 26, height: 32, bob: 6 },
  selection: { frame: [377, 700, 263, 297], width: 16, height: 18, bob: 3 },
} as const;

export default function UnitArrow({ character, onError, variant }: {
  character: RefObject<Sprite | null>;
  onError: () => void;
  variant: keyof typeof APPEARANCES;
}) {
  const appearance = APPEARANCES[variant];
  const sprite = useRef<Sprite>(null);
  const elapsed = useRef(0);
  const [texture, setTexture] = useState<Texture | null>(null);

  useEffect(() => {
    let active = true;
    let owned: Texture | undefined;
    Assets.load<Texture>(arrowsUrl).then((sheet) => {
      if (!active) return;
      sheet.source.scaleMode = "nearest";
      owned = new Texture({ source: sheet.source, frame: new Rectangle(...appearance.frame) });
      setTexture(owned);
    }).catch(() => { if (active) onError(); });
    return () => { active = false; owned?.destroy(false); };
  }, [onError, appearance]);

  useTick((ticker) => {
    elapsed.current = (elapsed.current + ticker.deltaMS) % BOB_PERIOD_MS;
    const marker = sprite.current;
    const unit = character.current;
    if (!marker) return;
    marker.visible = Boolean(unit && !unit.destroyed && unit.visible);
    if (unit && !unit.destroyed) {
      const bob = (1 - Math.cos(elapsed.current / BOB_PERIOD_MS * Math.PI * 2)) * appearance.bob / 2;
      marker.position.set(unit.x, unit.y - unit.height * unit.anchor.y - 6 - bob);
    }
  });

  return texture ? <pixiSprite
    ref={sprite}
    texture={texture}
    anchor={{ x: 0.5, y: 1 }}
    width={appearance.width} height={appearance.height}
    zIndex={1000000}
    visible={false}
    roundPixels
    eventMode="none"
  /> : null;
}
