import type { Ability, Cell, Position, Weapon } from "./models";
import type { CharacterModel } from "@/components/ProceduralMap/types";
import { backendURL, backendHeaders } from "./backendURL";

export interface SkillToken extends Position {
  turnCount?: number;
  id: string;
  team: number;
  model: CharacterModel;
  life: number;
  shield: number;
  maxLife: number;
  maxShield: number;
  weaponID: string;
  abilityIDs: string[];
  staminaSpent: number;
  abilityUses?: Record<string, number>;
  movementMode?: "jump" | "";
  path: Position[];
}

export interface LabConfig { board: Cell[]; weapons: Weapon[]; abilities: Ability[]; baseStamina: number }
export interface LabResult { board: Cell[]; tokens: SkillToken[]; message: string }
export interface LabAbilityResult extends LabResult { board: Cell[]; matchAbilityUses: Record<string, number> }
export type LabLayout = "arena" | "procedural";

function labURL() {
  return backendURL("/api/skills-lab");
}

async function request<T>(url: URL, init: RequestInit): Promise<T> {
  let response: Response;
  try { response = await fetch(url, { ...init, headers: backendHeaders(url, init.headers) }); }
  catch (error) {
    if (init.signal?.aborted) throw error;
    throw new Error("Não foi possível acessar o laboratório. Verifique se o backend está atualizado e em execução.");
  }
  if (response.headers.get("Content-Type")?.includes("text/html")) throw new Error("A API devolveu uma página HTML. Confira o túnel do backend e reinicie o servidor atualizado.");
  const data = await response.json().catch(() => null);
  if (!response.ok) throw new Error(data?.error ?? "Não foi possível executar o teste.");
  return data as T;
}

export function loadLab(seed: string, layout: LabLayout, signal: AbortSignal) {
  const url = labURL();
  url.search = new URLSearchParams({ seed, layout }).toString();
  return request<LabConfig>(url, { signal });
}

export function attackInLab(board: Cell[], tokens: SkillToken[], attackerID: string, targetID: string | undefined, signal: AbortSignal, target?: Position) {
  const url = labURL();
  url.pathname += "/attack";
  return request<LabResult>(url, {
    method: "POST", signal, headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ board, tokens, attackerID, targetID, target }),
  });
}

export function activateAbilityInLab(board: Cell[], tokens: SkillToken[], actorID: string, abilityID: string, target: Position, matchAbilityUses: Record<string, number>, signal: AbortSignal) {
  const url = labURL(); url.pathname += "/ability";
  return request<LabAbilityResult>(url, { method: "POST", signal, headers: { "Content-Type": "application/json" },
    body: JSON.stringify({board,tokens,actorID,abilityID,target,matchAbilityUses}) });
}
