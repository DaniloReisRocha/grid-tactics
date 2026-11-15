import type { MapTile } from "./generator";

export type CharacterModel = 0 | 1 | 2;
export interface MapActor {
  id: string;
  tile: MapTile;
  path: readonly MapTile[];
  life: number;
  shield: number;
  model: CharacterModel;
  color: string;
  selected: boolean;
  activeTurn?: boolean;
  defeated: boolean;
  movementMode?: "jump" | "";
}
export interface MapInteraction {
  path: readonly { x: number; y: number }[];
  area?: readonly { x: number; y: number }[];
  selectable: ReadonlySet<string>;
  attackTargets: ReadonlySet<string>;
  onSelect: (tile: MapTile) => void;
  onHover: (tile: MapTile | null) => void;
}
