import { useCallback, useEffect, useRef, useState } from "react";
import type { GameStatePayload, UnitInstance } from "@/api/models";
import type { MapTile } from "@/components/ProceduralMap/generator";

const EMPTY_ORBS: GameStatePayload["orbs"] = [];

export function useGameOrbs(orbs: GameStatePayload["orbs"] = EMPTY_ORBS, units: readonly UnitInstance[]) {
  const seen = useRef(new Set<string>());
  const [hidden, setHidden] = useState<ReadonlySet<string>>(() => new Set());
  const [bursting, setBursting] = useState<ReadonlySet<string>>(() => new Set());

  useEffect(() => {
    const alreadyCollected = orbs.filter((orb) => !seen.current.has(orb.id) && orb.triggered).map((orb) => orb.id);
    for (const orb of orbs) seen.current.add(orb.id);
    // Reconnecting must not replay historical collections or revive old orbs.
    if (alreadyCollected.length) setHidden((ids) => new Set([...ids, ...alreadyCollected]));
    const defeatedCollectors = orbs.filter((orb) => orb.triggered && units.some((unit) => unit.unit.id === orb.collectedBy && unit.currentLife <= 0)).map((orb) => orb.id);
    if (defeatedCollectors.length) setBursting((ids) => {
      const missing = defeatedCollectors.filter((id) => !ids.has(id));
      return missing.length ? new Set([...ids, ...missing]) : ids;
    });
  }, [orbs, units]);

  const onActorTileEnter = useCallback((tile: MapTile, actorID: string) => {
    const collected = orbs.filter((orb) => orb.triggered && orb.collectedBy === actorID && orb.x === tile.x && orb.y === tile.y);
    if (collected.length) setBursting((ids) => new Set([...ids, ...collected.map((orb) => orb.id)]));
  }, [orbs]);
  const onOrbComplete = useCallback((id: string) => setHidden((ids) => new Set([...ids, id])), []);

  return {
    orbs: orbs.filter((orb) => !hidden.has(orb.id) && (!orb.triggered || seen.current.has(orb.id)))
      .map((orb) => ({...orb, triggered: bursting.has(orb.id)})),
    onActorTileEnter, onOrbComplete,
  };
}
