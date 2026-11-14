import type { Cell, Position, Weapon } from "./models";
import type { CharacterModel } from "@/components/ProceduralMap/types";

export interface SkillToken extends Position {
  id: string;
  team: number;
  model: CharacterModel;
  life: number;
  shield: number;
  maxLife: number;
  maxShield: number;
  weaponID: string;
  path: Position[];
}

export interface LabConfig { board: Cell[]; weapons: Weapon[] }
export interface LabResult { tokens: SkillToken[]; message: string }
export type LabLayout = "arena" | "procedural";

function labURL() {
  const socketURL = import.meta.env.VITE_WS_URL as string | undefined;
  const base = socketURL ? new URL(socketURL, window.location.href)
    : new URL(`${window.location.protocol === "https:" ? "https:" : "http:"}//${window.location.hostname}:8080`);
  base.protocol = base.protocol === "wss:" || base.protocol === "https:" ? "https:" : "http:";
  base.pathname = "/api/skills-lab";
  base.search = "";
  base.hash = "";
  return base;
}

async function request<T>(url: URL, init: RequestInit): Promise<T> {
  let response: Response;
  try { response = await fetch(url, init); }
  catch (error) {
    if (init.signal?.aborted) throw error;
    throw new Error("Não foi possível acessar o laboratório. Verifique se o backend está atualizado e em execução.");
  }
  const data = await response.json().catch(() => null);
  if (!response.ok) throw new Error(data?.error ?? "Não foi possível executar o teste.");
  return data as T;
}

export function loadLab(seed: string, layout: LabLayout, signal: AbortSignal) {
  const url = labURL();
  url.search = new URLSearchParams({ seed, layout }).toString();
  return request<LabConfig>(url, { signal });
}

export function attackInLab(board: Cell[], tokens: SkillToken[], attackerID: string, targetID: string, signal: AbortSignal) {
  const url = labURL();
  url.pathname += "/attack";
  return request<LabResult>(url, {
    method: "POST", signal, headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ board, tokens, attackerID, targetID }),
  });
}
