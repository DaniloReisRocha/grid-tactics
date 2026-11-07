import { useCallback, useEffect, useMemo, useRef, useState, type CSSProperties } from "react";
import { Application, extend } from "@pixi/react";
import { Assets, Container, Graphics, Polygon, Sprite, type FederatedPointerEvent, type Texture } from "pixi.js";
import Minimap from "./Minimap";
import Character from "./Character";
import { findPath, findSpawn, isWalkable, tileKey } from "./characterMovement";
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

extend({ Container, Graphics, Sprite });

const EMPTY_PATH: readonly MapTile[] = [];

export interface ProceduralMapProps {
  /** Number of columns, in tiles. */
  width: number;
  /** Number of rows, in tiles. */
  height: number;
  /** The same seed and dimensions produce the same map. */
  seed?: string;
  /** Show the top-down overview in the corner. */
  showMinimap?: boolean;
  /** Enable the character controlled by clicking tiles. */
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
  const lookup = useMemo(() => new Map(tiles.map((tile) => [tileKey(tile.x, tile.y), tile])), [tiles]);
  const [hover, setHover] = useState<{ tiles: MapTile[]; tile: MapTile } | null>(null);
  const [selection, setSelection] = useState<{ tiles: MapTile[]; path: MapTile[] } | null>(null);
  const hoveredTile = hover?.tiles === tiles ? hover.tile : undefined;
  const selectedPath = selection?.tiles === tiles ? selection.path : EMPTY_PATH;
  const preview = useMemo(() => (
    showCharacter && !characterError && hoveredTile && characterPosition
      ? findPath(characterPosition, hoveredTile, lookup) : null
  ), [showCharacter, characterError, hoveredTile, characterPosition, lookup]);
  const remainingPath = useMemo(() => {
    const index = selectedPath.findIndex((tile) => tile === characterPosition);
    return index >= 0 ? selectedPath.slice(index) : EMPTY_PATH;
  }, [selectedPath, characterPosition]);
  const highlightedPath = hoveredTile ? preview ?? EMPTY_PATH : remainingPath;
  const highlightedKeys = useMemo(() => new Set(highlightedPath.map((tile) => tileKey(tile.x, tile.y))), [highlightedPath]);
  const destination = highlightedPath[highlightedPath.length - 1];
  const selectTile = (tile: MapTile) => {
    if (!showCharacter || characterError || !characterPosition) return;
    const path = findPath(characterPosition, tile, lookup);
    if (path) setSelection({ tiles, path });
  };

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
      role="group"
      aria-label="Mapa isométrico. Passe o mouse nos tiles para visualizar o caminho e clique para mover o personagem."
      onPointerLeave={() => setHover(null)}
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
              const markerY = surfaceY - 42 + (hasBridge(tile) ? -8 : 0);
              const isHovered = hoveredTile === tile;
              const isOnPath = highlightedKeys.has(tileKey(tile.x, tile.y));
              const unreachable = isHovered && !preview;
              const isDestination = destination === tile;

              return (
                <pixiContainer
                  key={`${tile.x}-${tile.y}`} x={isoX} y={isoY}
                  zIndex={(tile.x + tile.y) * 10 + tile.y * 0.01}
                  eventMode={showCharacter && !characterError ? "static" : "none"}
                  interactiveChildren={false}
                  hitArea={new Polygon([
                    0, markerY - TILE_HEIGHT / 2,
                    TILE_WIDTH / 2, markerY,
                    0, markerY + TILE_HEIGHT / 2,
                    -TILE_WIDTH / 2, markerY,
                  ])}
                  cursor={isWalkable(tile) && !(isHovered && !preview) ? "pointer" : "not-allowed"}
                  onPointerOver={() => setHover({ tiles, tile })}
                  onPointerOut={() => setHover((previous) => previous?.tile === tile ? null : previous)}
                  onPointerTap={(event: FederatedPointerEvent) => { if (event.button === 0) selectTile(tile); }}
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
                  {showCharacter && (isOnPath || isHovered) ? (
                    <pixiGraphics
                      eventMode="none"
                      draw={(graphics) => {
                        const color = unreachable ? 0xef6c67 : isDestination ? 0xffdc62 : 0x78dcb1;
                        graphics.clear();
                        graphics.poly([
                          0, markerY - TILE_HEIGHT / 2,
                          TILE_WIDTH / 2, markerY,
                          0, markerY + TILE_HEIGHT / 2,
                          -TILE_WIDTH / 2, markerY,
                        ]).fill({ color, alpha: unreachable ? 0.45 : 0.38 })
                          .stroke({ color, width: isDestination || unreachable ? 2 : 1 });
                      }}
                    />
                  ) : null}
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
                spawn={spawn}
                path={selectedPath}
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
            : hoveredTile && !preview ? "Destino inacessível"
            : "Passe o mouse para ver o caminho · Clique para mover"}</div>
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
