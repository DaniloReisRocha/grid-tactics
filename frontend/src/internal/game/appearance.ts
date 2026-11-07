import type { CharacterModel } from "@/components/ProceduralMap/types";

const TEAM_COLORS = ["#ffdc62", "#6ccfff", "#ef8acb"] as const;

/** Stable across clients/reconnects; each of up to three players gets a distinct sheet block. */
export function playerAppearance(sessionID: string, playerIndex: number) {
  let hash = 2166136261;
  for (let index = 0; index < sessionID.length; index++) {
    hash = Math.imul(hash ^ sessionID.charCodeAt(index), 16777619);
  }
  return {
    model: (((hash >>> 0) % 3 + playerIndex) % 3) as CharacterModel,
    color: TEAM_COLORS[playerIndex % 3],
  };
}
