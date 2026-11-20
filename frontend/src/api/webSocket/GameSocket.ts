import type { GameStatePayload } from "@/api/models";
import type { UnitLoadout } from "@/api/team";
import { ConnectionStatus } from "../status";
export { ConnectionStatus } from "../status";

export const MessageType = {
  echo: "echo", joinGame: "joinGame", movePiece: "movePiece",
  attackPiece: "attackPiece", endTurn: "endTurn", leaveGame: "leaveGame",
  actionFailed: "actionFailed", gameStateUpdate: "gameStateUpdate",
  useAbility: "useAbility",
} as const;

export type ServerMessage =
  | { type: "gameStateUpdate"; payload: GameStatePayload }
  | { type: "actionFailed"; payload: { action: string; message: string } }
  | { type: "echo"; payload: unknown };

class GameConnection {
  private socket: WebSocket | null = null;
  private desired: { sessionID: string; accoID: string; team?: UnitLoadout[] } | null = null;
  private retry: ReturnType<typeof setTimeout> | undefined;
  private attempts = 0;
  private messages = new Set<(message: ServerMessage) => void>();
  private statuses = new Set<(status: ConnectionStatus) => void>();

  subscribe(listener: (message: ServerMessage) => void) {
    this.messages.add(listener);
    return () => { this.messages.delete(listener); };
  }
  onStatus(listener: (status: ConnectionStatus) => void) {
    this.statuses.add(listener);
    return () => { this.statuses.delete(listener); };
  }
  private status(status: ConnectionStatus) { this.statuses.forEach((listener) => listener(status)); }
  connect(sessionID: string, accoID: string, team?: UnitLoadout[]) {
    if (this.desired?.sessionID === sessionID && this.desired.accoID === accoID
      && this.socket && this.socket.readyState <= WebSocket.OPEN) return;
    this.disconnect();
    this.desired = { sessionID, accoID, ...(team ? { team: team.map((unit) => ({ weaponID: unit.weaponID, abilityIDs: [...unit.abilityIDs] })) } : {}) };
    this.open();
  }
  private open() {
    if (!this.desired) return;
    this.status(this.attempts ? ConnectionStatus.reconnecting : ConnectionStatus.connecting);
    const socket = new WebSocket(import.meta.env.VITE_WS_URL);
    this.socket = socket;
    socket.onopen = () => {
      if (this.socket !== socket || !this.desired) return;
      this.attempts = 0;
      this.status(ConnectionStatus.connected);
      this.sendAction(MessageType.joinGame, this.desired);
    };
    socket.onmessage = (event) => {
      if (this.socket !== socket) return;
      try {
        const message = JSON.parse(String(event.data)) as ServerMessage;
        if (message && typeof message.type === "string") this.messages.forEach((listener) => listener(message));
      } catch { this.notifyError("Resposta inválida do servidor."); }
    };
    socket.onerror = () => { /* onclose handles retry, including failed handshakes. */ };
    socket.onclose = (event) => {
      if (this.socket !== socket) return;
      this.socket = null;
      if (event.code === 4001) {
        this.desired = null;
        this.status(ConnectionStatus.disconnected);
        this.notifyError("Esta conta entrou em outra conexão. Volte ao início para entrar novamente.");
        return;
      }
      if (!this.desired) { this.status(ConnectionStatus.disconnected); return; }
      this.status(ConnectionStatus.reconnecting);
      const delay = Math.min(10000, 500 * 2 ** this.attempts++);
      this.retry = setTimeout(() => this.open(), delay);
    };
  }
  private notifyError(message: string) {
    this.messages.forEach((listener) => listener({ type: "actionFailed", payload: { action: "connection", message } }));
  }
  sendAction(type: string, payload: unknown = {}) {
    if (this.socket?.readyState !== WebSocket.OPEN) return false;
    this.socket.send(JSON.stringify({ type, payload }));
    return true;
  }
  disconnect() {
    this.desired = null;
    clearTimeout(this.retry);
    this.retry = undefined;
    const socket = this.socket;
    this.socket = null;
    socket?.close();
    this.attempts = 0;
    this.status(ConnectionStatus.disconnected);
  }
}

export const GameSocket = new GameConnection();
