import { GameStatus } from "@/api/status";
import { useGameSocket } from "@/store/GameSocket";
import { useTurnSeconds } from "./useTurnSeconds";

export default function TurnTimer() {
  const status = useGameSocket((state) => state.status);
  const yourTurn = useGameSocket((state) => state.currentTurn !== null && state.currentTurn === state.accountID);
  const seconds = useTurnSeconds();
  return <div style={{ display: "flex", alignItems: "center", gap: 16 }}>
    <div role="timer" aria-label="Tempo restante do turno"
    style={{ display: "grid", gap: 2, color: seconds <= 10 ? "#ff7777" : "#e9eddf" }}>
    <span style={{ fontSize: 12 }}>Tempo do turno</span>
    <strong style={{ fontSize: 26, lineHeight: 1.2, fontVariantNumeric: "tabular-nums" }}>{seconds}s</strong>
    {status === GameStatus.paused ? <small>Pausado</small> : null}
    </div>
    <span style={{ color: yourTurn ? "#b1de73" : "#e9eddf", fontWeight: 600 }}>
      {yourTurn ? "seu turno" : "Turno do oponente"}
    </span>
  </div>;
}
