import { useMemo, useRef } from "react";
import Character from "./Character";
import DefeatedUnit from "./DefeatedUnit";
import type { MapTile } from "./generator";
import type { MapActor as Actor } from "./types";
import type { Sprite } from "pixi.js";
import type { MapView } from "./projection";

export default function MapActor({ actor, view, onSelect, onHover, onError, onSelectedSprite, onSelectedElevation }: {
  actor: Actor;
  view: MapView;
  onSelect?: (tile: MapTile) => void;
  onHover?: (tile: MapTile | null) => void;
  onError: () => void;
  onSelectedSprite: (sprite: Sprite | null) => void;
  onSelectedElevation: (height: number) => void;
}) {
  const initialTile = useRef(actor.tile);
  // Unrelated broadcasts (turns/combat) must not restart an existing walk animation.
  const signature = actor.path.map((tile) => `${tile.x},${tile.y}`).join(";");
  const path = useMemo(() => actor.path, [signature]);
  if (actor.defeated) return <DefeatedUnit tile={actor.tile} view={view} onError={onError} />;
  return <Character
    view={view}
    spawn={initialTile.current} path={path} model={actor.model}
    onError={onError}
    onSprite={actor.selected ? onSelectedSprite : undefined}
    onElevation={actor.selected ? onSelectedElevation : undefined}
    onSelect={onSelect ? () => onSelect(actor.tile) : undefined}
    onHover={(hovered) => onHover?.(hovered ? actor.tile : null)}
  />;
}
