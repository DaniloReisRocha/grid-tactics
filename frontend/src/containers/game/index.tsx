import { useNavigate } from "react-router-dom";
import { Board } from "./Board";
import { PieceInfoPanel } from "./PieceInfoPanel";
import WaitingScreen from "./WaitingScreen";
import { TurnControl } from "./TurnControl";
import { useGameSocket } from "@/store/GameSocket";
import { useStartGame } from "@/hooks/game";

type GameProps = { sessionID: string; accoID: string };

export const Game = ({ sessionID, accoID }: GameProps) => {
  const navigate = useNavigate();
  const { players, status, connection, error, clearError, leaveGame, winner } = useGameSocket();
  useStartGame(sessionID, accoID);
  const connected = Object.values(players).filter((player) => player.connected).length;
  return (
    <main style={{ height: "100vh", width: "100vw", overflow: "auto", boxSizing: "border-box", padding: 16 }}>
      <header style={{ display: "flex", flexWrap: "wrap", alignItems: "center", gap: 16 }}>
        <strong>Sessão: {sessionID}</strong>
        <span>{connection === "connected" ? "Conectado" : connection === "reconnecting" ? "Reconectando..." : connection === "connecting" ? "Conectando..." : "Desconectado"}</span>
        <button type="button" onClick={() => { leaveGame(true); navigate("/home"); }}>Sair da partida</button>
      </header>
      {error ? <div role="alert">{error} <button onClick={clearError}>Fechar</button></div> : null}
      {status === "paused" ? <p role="status">Partida pausada. Aguardando reconexão do outro jogador.</p> : null}
      {status === "finished" ? <p role="status">Partida encerrada. Vencedor: {winner ? players[winner]?.name : "nenhum"}.</p> : null}
      {status === "waiting" ? <WaitingScreen count={connected} /> : (
        <div style={{ display: "flex", flexWrap: "wrap", gap: 20, marginTop: 16, alignItems: "flex-start" }}>
          <TurnControl />
          <Board />
          <PieceInfoPanel />
        </div>
      )}
    </main>
  );
};
