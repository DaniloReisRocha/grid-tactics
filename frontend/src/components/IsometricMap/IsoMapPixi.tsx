import { Application, extend } from "@pixi/react";
import { Sprite, Container, Texture } from "pixi.js";
import { useMemo } from "react";

extend({ Sprite, Container });

const TILE_WIDTH = 128;
const TILE_HEIGHT = 64;

type Tile = {
  name: string;
  z?: number;
};

interface Props {
  map: Tile[][];
}

export default function IsoMapPixi({ map }: Props) {
  const textures = useMemo(() => {
    const tex: Record<string, Texture> = {};
    for (const row of map) {
      for (const tile of row) {
        if (!tex[tile.name]) {
          tex[tile.name] = Texture.from(
            "/assets/tiles/sketchTown/dirt_center_E.png"
          );
        }
      }
    }
    return tex;
  }, [map]);

  const rows = map.length;
  const cols = map[0]?.length || 0;

  return (
    <Application width={800} height={600} background={0xc6bfa7}>
      <pixiContainer x={400} y={100}>
        {map.flatMap((row, y) =>
          row.map((tile, x) => {
            const z = tile.z || 0;
            const isoX = (x - y) * (TILE_WIDTH / 2);
            const isoY = (x + y) * (TILE_HEIGHT / 2) - z * (TILE_HEIGHT / 2);

            return (
              <pixiSprite
                key={`${x}-${y}`}
                texture={textures[tile.name]}
                x={isoX}
                y={isoY}
                anchor={{ x: 0.5, y: 1 }}
              />
            );
          })
        )}
      </pixiContainer>
    </Application>
  );
}
