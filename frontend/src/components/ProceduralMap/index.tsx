import { useCallback, useEffect, useMemo, useRef, useState, type CSSProperties } from "react";
import { Application, extend } from "@pixi/react";
import { Assets, Container, Graphics, Polygon, Sprite, type FederatedPointerEvent, type Texture } from "pixi.js";
import Minimap from "./Minimap";
import MapActorRenderer from "./MapActor";
import StructureOcclusion from "./StructureOcclusion";
import type { MapActor, MapInteraction } from "./types";
import { tileKey } from "./characterMovement";
import {
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

const EMPTY_ACTORS: readonly MapActor[] = [];

export interface ProceduralMapProps {
  /** Number of columns, in tiles. */
  width: number;
  /** Number of rows, in tiles. */
  height: number;
  /** Authoritative terrain generated and validated by the server. */
  tiles: readonly MapTile[];
  actors?: readonly MapActor[];
  interaction?: MapInteraction;
  /** Show the top-down overview in the corner. */
  showMinimap?: boolean;
  className?: string;
  /** The parent must provide a height, or set one here (for example, 600). */
  style?: CSSProperties;
}

type TextureMap = Record<Asset, Texture>;

export default function ProceduralMap({
  width,
  height,
  tiles: providedTiles,
  actors = EMPTY_ACTORS,
  interaction,
  showMinimap = true,
  className,
  style,
}: ProceduralMapProps) {
  const containerRef = useRef<HTMLDivElement>(null);
  const structures = useRef(new Map<string, Sprite>());
  const structureRefs = useRef(new Map<string, (sprite: Sprite | null) => void>());
  const selectedSprite = useRef<Sprite | null>(null);
  const registerSelectedSprite = useCallback((sprite: Sprite | null) => { selectedSprite.current = sprite; }, []);
  const structureRef = (key: string) => {
    let callback = structureRefs.current.get(key);
    if (!callback) {
      callback = (sprite: Sprite | null) => {
        const previous = structures.current.get(key);
        if (previous && previous !== sprite && !previous.destroyed) previous.alpha = 1;
        if (sprite) structures.current.set(key, sprite);
        else structures.current.delete(key);
      };
      structureRefs.current.set(key, callback);
    }
    return callback;
  };
  const [viewport, setViewport] = useState({ width: 0, height: 0 });
  const [textures, setTextures] = useState<TextureMap | null>(null);
  const [loadError, setLoadError] = useState(false);
  const tiles = useMemo(() => [...providedTiles]
    .sort((a, b) => a.x + a.y - b.x - b.y || a.y - b.y), [providedTiles]);
  const [characterError, setCharacterError] = useState(false);
  const reportCharacterError = useCallback(() => setCharacterError(true), []);
  const [hover, setHover] = useState<{ tiles: MapTile[]; tile: MapTile } | null>(null);
  const hoveredTile = hover?.tiles === tiles ? hover.tile : undefined;
  const highlightedPath = interaction?.path ?? [];
  const highlightedKeys = new Set(highlightedPath.map((tile) => tileKey(tile.x, tile.y)));
  const destination = highlightedPath[highlightedPath.length - 1];

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
      onPointerLeave={() => { setHover(null); interaction?.onHover(null); }}
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
              const selectable = interaction?.selectable.has(tileKey(tile.x, tile.y)) ?? false;
              const unreachable = isHovered && !selectable;
              const isDestination = destination?.x === tile.x && destination.y === tile.y;
              const selectedActor = actors?.find((actor) => actor.selected && actor.tile.x === tile.x && actor.tile.y === tile.y);
              const attackTarget = interaction?.attackTargets.has(tileKey(tile.x, tile.y));

              return (
                <pixiContainer
                  key={`${tile.x}-${tile.y}`} x={isoX} y={isoY}
                  zIndex={(tile.x + tile.y) * 10 + tile.y * 0.01}
                  eventMode={interaction ? "static" : "none"}
                  interactiveChildren={false}
                  hitArea={new Polygon([
                    0, markerY - TILE_HEIGHT / 2,
                    TILE_WIDTH / 2, markerY,
                    0, markerY + TILE_HEIGHT / 2,
                    -TILE_WIDTH / 2, markerY,
                  ])}
                  cursor={selectable ? "pointer" : "not-allowed"}
                  onPointerOver={() => { setHover({ tiles, tile }); interaction?.onHover(tile); }}
                  onPointerOut={() => { setHover((previous) => previous?.tile === tile ? null : previous); interaction?.onHover(null); }}
                  onPointerTap={(event: FederatedPointerEvent) => { if (event.button === 0) interaction?.onSelect(tile); }}
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
                        ref={level > 0 ? structureRef(`${tile.x},${tile.y}:ground:${level}`) : undefined}
                        texture={textures[asset]}
                        y={-level * LEVEL_HEIGHT}
                        scale={SPRITE_SCALE}
                        anchor={{ x: 0.5, y: 1 }}
                      />
                    );
                  })}
                  {interaction && (isOnPath || isHovered || selectedActor || attackTarget) ? (
                    <pixiGraphics
                      eventMode="none"
                      draw={(graphics) => {
                        const color = unreachable || attackTarget ? 0xef6c67 : isDestination ? 0xffdc62 : selectedActor ? Number.parseInt(selectedActor.color.slice(1), 16) : 0x78dcb1;
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
                      ref={structureRef(`${tile.x},${tile.y}:object`)}
                      texture={textures[tile.object]}
                      y={surfaceY - OBJECT_RULES[tile.object].surfaceOffset}
                      scale={SPRITE_SCALE}
                      anchor={{ x: 0.5, y: 1 }}
                    />
                  ) : null}
                  {tile.object === "house" && tile.roof ? (
                    <pixiSprite
                      ref={structureRef(`${tile.x},${tile.y}:roof`)}
                      texture={textures[tile.roof]}
                      y={surfaceY - LEVEL_HEIGHT * 2}
                      scale={SPRITE_SCALE}
                      anchor={{ x: 0.5, y: 1 }}
                    />
                  ) : null}
                </pixiContainer>
              );
            })}
            {actors.map((actor) => (
              <MapActorRenderer key={actor.id} actor={actor}
                onSelect={interaction?.onSelect}
                onHover={(tile) => { setHover(tile ? { tiles, tile } : null); interaction?.onHover(tile); }}
                onSelectedSprite={registerSelectedSprite}
                onError={reportCharacterError} />
            ))}
            <StructureOcclusion structures={structures} selectedSprite={selectedSprite} />
          </pixiContainer>
        ) : null}
      </Application>
      ) : null}

      {showMinimap ? (
        <Minimap tiles={tiles} width={width} height={height} actors={actors} />
      ) : null}

      {actors.length > 0 ? (
        <div style={styles.characterInfo}>
          <div>{characterError ? "Não foi possível carregar o personagem."
            : hoveredTile && !interaction?.selectable.has(tileKey(hoveredTile.x, hoveredTile.y)) ? "Destino inacessível"
            : "Selecione uma unidade · Passe o mouse para ver a rota · Clique para agir"}</div>
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
