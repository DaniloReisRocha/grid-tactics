import { useEffect, useRef } from "react";
import dirtCenter from "@/assets/tiles/sketchTown/dirt_center_E.png";

type TileType = "grass" | "road" | "house";
export type MapGrid = TileType[][];

interface Props {
  map: MapGrid;
  tileWidth?: number; // largura da tile (ex: 64)
  tileHeight?: number; // altura da tile (ex: 32)
}

const tileImages: Record<TileType, string> = {
  grass: dirtCenter,
  road: dirtCenter,
  house: dirtCenter,
};

export default function IsometricMap({
  map,
  tileWidth = 64,
  tileHeight = 32,
}: Props) {
  const canvasRef = useRef<HTMLCanvasElement>(null);

  useEffect(() => {
    const canvas = canvasRef.current;
    if (!canvas) return;
    const ctx = canvas.getContext("2d");
    if (!ctx) return;

    const rows = map.length;
    const cols = map[0]?.length || 0;
    canvas.width = (cols + rows) * (tileWidth / 2);
    canvas.height = (cols + rows) * (tileHeight / 2);

    const images: Partial<Record<TileType, HTMLImageElement>> = {};
    let loaded = 0;

    const loadImages = () => {
      for (const [type, src] of Object.entries(tileImages)) {
        const img = new Image();
        img.src = src;
        img.onload = () => {
          loaded++;
          images[type as TileType] = img;
          if (loaded === Object.keys(tileImages).length) {
            drawMap();
          }
        };
      }
    };

    const drawMap = () => {
      ctx.clearRect(0, 0, canvas.width, canvas.height);
      for (let y = 0; y < rows; y++) {
        for (let x = 0; x < cols; x++) {
          const tile = map[y][x];
          const img = images[tile];
          if (!img) continue;

          // Cálculo da posição isométrica
          const isoX =
            (x - y) * (tileWidth / 2) + canvas.width / 2 - tileWidth / 2;
          const isoY = (x + y) * (tileHeight / 2);
          ctx.drawImage(img, isoX, isoY, tileWidth, tileHeight);
        }
      }
    };

    loadImages();
  }, [map, tileWidth, tileHeight]);

  return <canvas ref={canvasRef} style={{ border: "1px solid #ccc" }} />;
}
