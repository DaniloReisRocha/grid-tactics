import { useCallback, useEffect, useMemo, type RefObject } from "react";
import { useTick } from "@pixi/react";
import { Bounds, type Sprite, type Ticker } from "pixi.js";
import { approachOpacity, OCCLUDED_OPACITY, occludesUnit } from "./occlusion";

interface OcclusionProps {
  structures: RefObject<Map<string, Sprite>>;
  selectedSprite: RefObject<Sprite | null>;
}

/** Update after character movement so fading tracks the actual animated position. */
export default function StructureOcclusion({ structures, selectedSprite }: OcclusionProps) {
  const bounds = useMemo(() => ({ unit: new Bounds(), structure: new Bounds() }), []);
  const update = useCallback((ticker: Ticker) => {
    const unit = selectedSprite.current;
    const selectedBounds = unit && !unit.destroyed ? unit.getBounds(false, bounds.unit) : null;
    for (const structure of structures.current.values()) {
      if (structure.destroyed) continue;
      const depth = structure.parent?.zIndex ?? -Infinity;
      const blocksUnit = selectedBounds && unit && depth > unit.zIndex
        && occludesUnit(selectedBounds, unit.zIndex, structure.getBounds(false, bounds.structure), depth);
      structure.alpha = approachOpacity(structure.alpha, blocksUnit ? OCCLUDED_OPACITY : 1, Math.min(ticker.deltaMS, 50));
    }
  }, [bounds, structures, selectedSprite]);

  useTick({ callback: update, priority: -10 });
  useEffect(() => () => {
    for (const sprite of structures.current.values()) {
      if (!sprite.destroyed) sprite.alpha = 1;
    }
  }, [structures]);
  return null;
}
