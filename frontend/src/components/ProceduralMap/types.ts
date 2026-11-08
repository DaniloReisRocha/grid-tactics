import type { MapTile } from "./generator";

export type CharacterModel = 0 | 1 | 2;
export interface MapActor {
  id: string;
  tile: MapTile;
  path: readonly MapTile[];
  model: CharacterModel;
  color: string;
  selected: boolean;
  defeated: boolean;
}
export interface MapInteraction {
  path: readonly { x: number; y: number }[];
  selectable: ReadonlySet<string>;
  attackTargets: ReadonlySet<string>;
  onSelect: (tile: MapTile) => void;
  onHover: (tile: MapTile | null) => void;
}
