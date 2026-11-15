import { useEffect, useRef, useState, type RefObject } from "react";
import { useTick } from "@pixi/react";
import { Assets, Rectangle, Sprite, Texture } from "pixi.js";
import arrowsUrl from "@/assets/arrows/direction_arrows_set_01.png";

const BOB_PERIOD_MS = 1400;
const BOB_DISTANCE = 6;

export default function ActiveTurnArrow({ character, onError }: {
  character: RefObject<Sprite | null>;
  onError: () => void;
}) {
  const sprite = useRef<Sprite>(null);
  const elapsed = useRef(0);
  const [texture, setTexture] = useState<Texture | null>(null);

  useEffect(() => {
    let active = true;
    let owned: Texture | undefined;
    Assets.load<Texture>(arrowsUrl).then((sheet) => {
      if (!active) return;
      sheet.source.scaleMode = "nearest";
      // Bounds of the downward arrow in the first row of the supplied sheet.
      owned = new Texture({ source: sheet.source, frame: new Rectangle(377, 51, 253, 307) });
      setTexture(owned);
    }).catch(() => { if (active) onError(); });
    return () => { active = false; owned?.destroy(false); };
  }, [onError]);

  useTick((ticker) => {
    elapsed.current = (elapsed.current + ticker.deltaMS) % BOB_PERIOD_MS;
    const marker = sprite.current;
    const unit = character.current;
    if (!marker) return;
    marker.visible = Boolean(unit && !unit.destroyed && unit.visible);
    if (unit && !unit.destroyed) {
      const bob = (1 - Math.cos(elapsed.current / BOB_PERIOD_MS * Math.PI * 2)) * BOB_DISTANCE / 2;
      marker.position.set(unit.x, unit.y - unit.height * unit.anchor.y - 6 - bob);
    }
  });

  return texture ? <pixiSprite
    ref={sprite}
    texture={texture}
    anchor={{ x: 0.5, y: 1 }}
    width={26} height={32}
    zIndex={1000000}
    visible={false}
    roundPixels
    eventMode="none"
  /> : null;
}
