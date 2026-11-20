import type { Cell, Position, Weapon } from "@/api/models";
import { weaponEffectProblem } from "./effects";

export function blocksDirectAttack(tile: Cell) {
  return tile.object === "house" || tile.object === "ruin" || tile.object === "barricade";
}

function terrainBlocksAttack(tile: Cell, source: Cell, target: Cell, entry: number, exit: number) {
  const start = source.height + 0.5;
  const delta = target.height - source.height;
  return tile.height + 1e-9 >= Math.min(start + delta * entry, start + delta * exit);
}

export function directAttackDistance(from: Position, target: Position): number | null {
  const dx = Math.abs(target.x-from.x), dy = Math.abs(target.y-from.y);
  if (dx===0 && dy===0) return null;
  if (Math.max(dx,dy)===1) return 1;
  return dx===0 || dy===0 ? dx+dy : null;
}

export function directAttackProblem(from: Position, target: Position, weapon: Weapon, board: readonly Cell[]) {
  const dx=target.x-from.x,dy=target.y-from.y;
  const distance=directAttackDistance(from,target);
  if (distance===null) return "Escolha uma das oito casas vizinhas ou um alvo na mesma linha ou coluna.";
  if (distance>weapon.range) return "Alvo fora do alcance da arma.";
  const cells=new Map(board.map((tile)=>[`${tile.x},${tile.y}`,tile]));
  const source=cells.get(`${from.x},${from.y}`),destination=cells.get(`${target.x},${target.y}`);
  const effectProblem=weaponEffectProblem(weapon,distance,source?.height,destination?.height);
  if (effectProblem) return effectProblem;
  // Retain geometry-only checks for callers without a map; game actions supply it.
  if (!board.length) return null;
  if (!source || !destination) return "Terreno fora do mapa.";
  if (dx!==0 && dy!==0) {
    for (const [x,y] of [[target.x,from.y],[from.x,target.y]]) {
      const side=cells.get(`${x},${y}`);
      if (!side) return "Terreno fora do mapa.";
      if (blocksDirectAttack(side)) return "Uma estrutura bloqueia a passagem diagonal do ataque.";
      if (terrainBlocksAttack(side,source,destination,0.5,0.5)) return "O relevo bloqueia a passagem diagonal do ataque.";
    }
  }
  const sx=Math.sign(dx),sy=Math.sign(dy);
  for (let x=from.x+sx,y=from.y+sy,step=1;;x+=sx,y+=sy,step++) {
    const tile=cells.get(`${x},${y}`);
    if (!tile) return "Terreno fora do mapa.";
    if (x===target.x&&y===target.y) return !blocksDirectAttack(tile)||tile.object==="barricade" ? null : "Uma estrutura bloqueia a linha de ataque.";
    if (blocksDirectAttack(tile)) return "Uma estrutura bloqueia a linha de ataque.";
    if (terrainBlocksAttack(tile,source,destination,(step-0.5)/distance,(step+0.5)/distance)) return "O relevo bloqueia a linha de ataque.";
  }
}

export function attackableArea(from: Position, weapon: Weapon, board: readonly Cell[]) {
  return new Set(board.filter((tile) => !directAttackProblem(from, tile, weapon, board))
    .map((tile) => `${tile.x},${tile.y}`));
}
