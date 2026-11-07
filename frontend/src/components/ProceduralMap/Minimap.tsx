import type { CSSProperties } from "react";
import { hasBridge, TERRAIN_LABELS, type Asset, type MapTile } from "./generator";

interface MinimapProps {
  tiles: readonly MapTile[];
  width: number;
  height: number;
  character?: Pick<MapTile, "x" | "y">;
}

const colors = Object.fromEntries(
  TERRAIN_LABELS.map(({ terrain, color }) => [terrain, color]),
) as Partial<Record<Asset, string>>;

export default function Minimap({ tiles, width, height, character }: MinimapProps) {
  return (
    <figure style={styles.panel}>
      <figcaption style={styles.caption}>
        <span>Minimapa</span>
        <span style={{ color: "#9ea9a1" }}>{width} × {height}</span>
      </figcaption>
      <svg
        viewBox={`0 0 ${width} ${height}`}
        preserveAspectRatio="xMidYMid meet"
        role="img"
        aria-label={`Vista superior do mapa: ${width} colunas e ${height} linhas. Cores seguem a legenda do terreno.`}
        style={styles.map}
      >
        {tiles.map((tile) => {
          const marker = tile.waterfall ? "waterfall"
            : hasBridge(tile) ? "bridgeX"
            : tile.object === "rockDirt" ? "rock" : tile.object;
          return (
            <g key={`${tile.x}-${tile.y}`}>
              <rect
                x={tile.x} y={tile.y} width={1} height={1}
                fill={colors[tile.terrain]}
                shapeRendering="crispEdges"
              />
              {tile.height > 0 ? (
                <rect
                  x={tile.x} y={tile.y} width={1} height={1}
                  fill="#ffffff" opacity={tile.height * 0.07}
                  shapeRendering="crispEdges"
                />
              ) : null}
              {marker ? (
                <rect
                  x={tile.x + (tile.object === "bridgeX" ? 0 : 0.18)}
                  y={tile.y + (tile.object === "bridgeY" ? 0 : 0.18)}
                  width={tile.object === "bridgeX" ? 1 : 0.64}
                  height={tile.object === "bridgeY" ? 1 : 0.64}
                  fill={colors[marker]}
                />
              ) : null}
            </g>
          );
        })}
        {character ? (
          <circle
            cx={character.x + 0.5} cy={character.y + 0.5} r={0.42}
            fill="#ffdc62" stroke="#17231f" strokeWidth={0.12}
          >
            <title>Personagem</title>
          </circle>
        ) : null}
      </svg>
    </figure>
  );
}

const styles: Record<string, CSSProperties> = {
  panel: {
    position: "absolute", right: 16, bottom: 76, margin: 0,
    width: "clamp(96px, 25%, 180px)", padding: 10, boxSizing: "border-box",
    border: "1px solid rgba(224, 237, 198, 0.24)", borderRadius: 12,
    background: "rgba(19, 31, 27, 0.94)",
    boxShadow: "0 8px 24px rgba(0, 0, 0, 0.25)", pointerEvents: "none",
  },
  caption: {
    display: "flex", flexWrap: "wrap", justifyContent: "space-between", gap: 4,
    marginBottom: 8, color: "#e9eddf", fontSize: 11, fontWeight: 700,
  },
  map: { display: "block", width: "100%", maxHeight: 160 },
};
