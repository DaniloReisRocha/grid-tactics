import { Fragment, useEffect, useRef, useState } from "react";
import Character from "./Character";
import DamageIndicator from "./DamageIndicator";
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
  const durability = actor.life + actor.shield;
  const previousDurability = useRef(durability);
  const [hit, setHit] = useState<{ id: number; amount: number } | null>(null);
  // Unrelated broadcasts (turns/combat) must not restart an existing walk animation.
  const signature = actor.path.map((tile) => `${tile.x},${tile.y}`).join(";");
  const stablePath = useRef({ signature, path: actor.path });
  if (stablePath.current.signature !== signature) stablePath.current = { signature, path: actor.path };
  const path = stablePath.current.path;

  useEffect(() => {
    const amount = previousDurability.current - durability;
    previousDurability.current = durability;
    if (amount > 0) setHit((current) => ({ id: (current?.id ?? 0) + 1, amount }));
  }, [durability]);

  return (
    <Fragment>
      {actor.defeated ? (
        <DefeatedUnit tile={actor.tile} view={view} onError={onError} />
      ) : (
        <Character
          view={view}
          jump={actor.movementMode === "jump"}
          spawn={initialTile.current} path={path} model={actor.model}
          onError={onError}
          onSprite={actor.selected ? onSelectedSprite : undefined}
          onElevation={actor.selected ? onSelectedElevation : undefined}
          onSelect={onSelect ? () => onSelect(actor.tile) : undefined}
          onHover={(hovered) => onHover?.(hovered ? actor.tile : null)}
        />
      )}
      {hit ? (
        <DamageIndicator
          key={hit.id}
          amount={hit.amount}
          tile={actor.tile}
          view={view}
          onError={onError}
          onComplete={() => setHit((current) => current?.id === hit.id ? null : current)}
        />
      ) : null}
    </Fragment>
  );
}
