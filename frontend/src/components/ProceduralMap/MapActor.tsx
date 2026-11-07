import { useMemo, useRef } from "react";
import Character from "./Character";
import type { MapTile } from "./generator";
import type { MapActor as Actor } from "./types";
import type { Sprite } from "pixi.js";

export default function MapActor({ actor, onSelect, onHover, onError, onSelectedSprite }: {
  actor: Actor;
  onSelect?: (tile: MapTile) => void;
  onHover?: (tile: MapTile | null) => void;
  onError: () => void;
  onSelectedSprite: (sprite: Sprite | null) => void;
}) {
  const initialTile = useRef(actor.tile);
  // Unrelated broadcasts (turns/combat) must not restart an existing walk animation.
  const signature = actor.path.map((tile) => `${tile.x},${tile.y}`).join(";");
  const path = useMemo(() => actor.path, [signature]);
  return <Character
    spawn={initialTile.current} path={path} model={actor.model}
    onError={onError}
    onSprite={actor.selected ? onSelectedSprite : undefined}
    onSelect={onSelect ? () => onSelect(actor.tile) : undefined}
    onHover={(hovered) => onHover?.(hovered ? actor.tile : null)}
  />;
}
