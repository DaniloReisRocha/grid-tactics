import { useCallback, useEffect, useMemo, useRef, useState, type CSSProperties } from "react";
import { Application, extend } from "@pixi/react";
import { Assets, Container, Sprite, type Texture } from "pixi.js";
import Minimap from "./Minimap";
import Character from "./Character";
import { findSpawn } from "./characterMovement";
import {
  generateSeededMap,
  hasBridge,
  LEVEL_HEIGHT,
  OBJECT_RULES,
  SPRITE_SCALE,
  TERRAIN_ASSETS,
  TERRAIN_LABELS,
  TILE_HEIGHT,
  TILE_WIDTH,
  type Asset,
  type MapTile,
} from "./generator";

extend({ Container, Sprite });

export interface ProceduralMapProps {
  /** Number of columns, in tiles. */
  width: number;
  /** Number of rows, in tiles. */
  height: number;
  /** The same seed and dimensions produce the same map. */
  seed?: string;
  /** Show the top-down overview in the corner. */
  showMinimap?: boolean;
  /** Enable the keyboard-controlled character. */
  showCharacter?: boolean;
  className?: string;
  /** The parent must provide a height, or set one here (for example, 600). */
  style?: CSSProperties;
}

type TextureMap = Record<Asset, Texture>;

export default function ProceduralMap({
  width,
  height,
  seed = "grid-tactics",
  showMinimap = true,
  showCharacter = true,
  className,
  style,
}: ProceduralMapProps) {
  const containerRef = useRef<HTMLDivElement>(null);
  const [viewport, setViewport] = useState({ width: 0, height: 0 });
  const [textures, setTextures] = useState<TextureMap | null>(null);
  const [loadError, setLoadError] = useState(false);
  const tiles = useMemo(() => generateSeededMap(seed, width, height), [seed, width, height]);
  const spawn = useMemo(() => findSpawn(tiles), [tiles]);
  const [characterLocation, setCharacterLocation] = useState<{ tiles: MapTile[]; tile: MapTile } | null>(null);
  const [characterError, setCharacterError] = useState(false);
  const reportCharacterError = useCallback(() => setCharacterError(true), []);
  const moveCharacter = useCallback((tile: MapTile) => setCharacterLocation({ tiles, tile }), [tiles]);
  const characterPosition = characterLocation?.tiles === tiles ? characterLocation.tile : spawn;

  const terrainCount = useMemo(() => {
    const count = Object.fromEntries(
      Object.keys(TERRAIN_ASSETS).map((asset) => [asset, 0]),
    ) as Record<Asset, number>;
    for (const tile of tiles) {
      count[tile.terrain] += 1;
      if (tile.object) {
        const asset = hasBridge(tile) ? "bridgeX"
          : tile.object === "rockDirt" ? "rock" : tile.object;
        count[asset] += 1;
      }
      if (tile.waterfall) count.waterfall += 1;
    }
    return count;
  }, [tiles]);

  // Include transparent sprite margins and room above the highest roof.
  const minX = -(height - 1) * (TILE_WIDTH / 2) - 32;
  const minY = -240;
  const projectedWidth = (width + height - 2) * (TILE_WIDTH / 2) + 64;
  const projectedHeight = (width + height - 2) * (TILE_HEIGHT / 2) - minY;
  const mapScale = Math.max(0.001, Math.min(
    1.1,
    (viewport.width - 32) / projectedWidth,
    (viewport.height - 68) / projectedHeight,
  ));
  const mapPosition = {
    x: (viewport.width - projectedWidth * mapScale) / 2 - minX * mapScale,
    y: (viewport.height - 36 - projectedHeight * mapScale) / 2 - minY * mapScale,
  };

  useEffect(() => {
    const element = containerRef.current;
    if (!element) return;
    const observer = new ResizeObserver(([entry]) => {
      setViewport({
        width: Math.floor(entry.contentRect.width),
        height: Math.floor(entry.contentRect.height),
      });
    });
    observer.observe(element);
    return () => observer.disconnect();
  }, []);

  useEffect(() => {
    let isMounted = true;
    Promise.all(Object.entries(TERRAIN_ASSETS).map(async ([asset, source]) => {
      const texture = await Assets.load<Texture>(source);
      return [asset, texture] as const;
    })).then((loadedTextures) => {
      if (isMounted) setTextures(Object.fromEntries(loadedTextures) as TextureMap);
    }).catch(() => {
      if (isMounted) setLoadError(true);
    });
    return () => { isMounted = false; };
  }, []);

  return (
    <div
      ref={containerRef}
      className={className}
      tabIndex={showCharacter ? 0 : undefined}
      role="group"
      aria-label="Mapa isométrico. Clique no mapa ou use Tab para focar, e mova o personagem com as setas ou WASD."
      onPointerDown={(event) => {
        if (showCharacter && !(event.target instanceof Element && event.target.closest("a, button, input"))) {
          event.currentTarget.focus({ preventScroll: true });
        }
      }}
      style={{ position: "relative", width: "100%", height: "100%", overflow: "hidden", background: "#17231f", ...style }}
    >
      {viewport.width > 0 && viewport.height > 0 ? (
      <Application
        width={viewport.width}
        height={viewport.height}
        backgroundColor={0x17231f}
        antialias
      >
        {textures ? (
          <pixiContainer
            x={mapPosition.x}
            y={mapPosition.y}
            scale={mapScale}
            sortableChildren
          >
            {/* Draw full columns back to front, so cliffs occlude objects behind them. */}
            {tiles.map((tile) => {
              const isoX = (tile.x - tile.y) * (TILE_WIDTH / 2);
              const isoY = (tile.x + tile.y) * (TILE_HEIGHT / 2);
              const surfaceY = -tile.height * LEVEL_HEIGHT;

              return (
                <pixiContainer
                  key={`${tile.x}-${tile.y}`} x={isoX} y={isoY}
                  zIndex={(tile.x + tile.y) * 10 + tile.y * 0.01}
                >
                  {Array.from({ length: tile.height + 1 }, (_, level) => {
                    const asset =
                      tile.waterfall && level > 0
                        ? "waterfall"
                        : level === tile.height
                          ? tile.terrain
                          : "dirt";
                    return (
                      <pixiSprite
                        key={`ground-${level}`}
                        texture={textures[asset]}
                        y={-level * LEVEL_HEIGHT}
                        scale={SPRITE_SCALE}
                        anchor={{ x: 0.5, y: 1 }}
                      />
                    );
                  })}
                  {tile.object ? (
                    <pixiSprite
                      texture={textures[tile.object]}
                      y={surfaceY - OBJECT_RULES[tile.object].surfaceOffset}
                      scale={SPRITE_SCALE}
                      anchor={{ x: 0.5, y: 1 }}
                    />
                  ) : null}
                  {tile.object === "house" && tile.roof ? (
                    <pixiSprite
                      texture={textures[tile.roof]}
                      y={surfaceY - LEVEL_HEIGHT * 2}
                      scale={SPRITE_SCALE}
                      anchor={{ x: 0.5, y: 1 }}
                    />
                  ) : null}
                </pixiContainer>
              );
            })}
            {showCharacter && spawn ? (
              <Character
                key={JSON.stringify([seed, width, height])}
                tiles={tiles}
                spawn={spawn}
                controls={containerRef}
                onMove={moveCharacter}
                onError={reportCharacterError}
              />
            ) : null}
          </pixiContainer>
        ) : null}
      </Application>
      ) : null}

      {showMinimap ? (
        <Minimap tiles={tiles} width={width} height={height} character={showCharacter && !characterError ? characterPosition : undefined} />
      ) : null}

      {showCharacter ? (
        <div style={styles.characterInfo}>
          <div>{characterError ? "Não foi possível carregar o personagem."
            : !spawn ? "Sem espaço livre para o personagem nesta seed."
            : "Clique no mapa · Setas / WASD para mover"}</div>
          <a href="https://axulart.itch.io/small-8-direction-characters" target="_blank" rel="noreferrer" style={{ color: "#a7c875" }}>
            Personagem: AxulArt · CC BY 4.0
          </a>
        </div>
      ) : null}

      <div style={styles.legend} aria-label="Legenda do mapa">
        {TERRAIN_LABELS.map(({ terrain, label, color }) => (
          <span key={terrain} style={styles.legendItem}>
            <span style={{ ...styles.swatch, background: color }} />
            {label} {terrainCount[terrain]}
          </span>
        ))}
      </div>

      {!textures ? (
        <div style={styles.loading} role="status">
          {loadError ? "Não foi possível carregar os assets do mapa." : "Carregando terreno..."}
        </div>
      ) : null}
    </div>
  );
}

const styles: Record<string, CSSProperties> = {
  characterInfo: {
    position: "absolute", bottom: 76, left: 16, maxWidth: "55%",
    padding: "7px 10px", borderRadius: 8, color: "#cbd4c8", fontSize: 11,
    background: "rgba(19, 31, 27, 0.9)",
  },
  legend: {
    position: "absolute", bottom: 12, left: 16, right: 16,
    display: "flex", flexWrap: "wrap", justifyContent: "center", gap: "7px 12px",
    pointerEvents: "none",
  },
  legendItem: {
    display: "inline-flex", alignItems: "center", gap: 5, color: "#cbd4c8",
    fontSize: 11, background: "rgba(19, 31, 27, 0.9)", borderRadius: 4, padding: "2px 5px",
  },
  swatch: { width: 8, height: 8, borderRadius: 2 },
  loading: {
    position: "absolute", top: "50%", left: "50%", transform: "translate(-50%, -50%)",
    color: "#cbd4c8", fontSize: 14,
  },
};
