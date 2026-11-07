import { useEffect } from "react";
import { Board } from "@/containers/game/Board";
import { PieceInfoPanel } from "@/containers/game/PieceInfoPanel";
import WaitingScreen from "@/containers/game/WaitingScreen";
import { TurnControl } from "@/containers/game/TurnControl";
import { useGameSocket } from "@/store/GameSocket";

type GameProps = {
  sessionID: string;
  accoID: string;
};

export const Game = (props: GameProps) => {
  const { sessionID, accoID } = props;
  const { joinGame, players } = useGameSocket();
  const playerCount = Object.keys(players).length;

  useEffect(() => {
    joinGame(sessionID, accoID);
  }, [joinGame, sessionID, accoID]);

  return (
    <div
      style={{
        height: "calc(100vh - 34px)",
        width: "100vw",
        display: "flex",
        alignItems: "center",
        justifyContent: "center",
      }}
    >
      {playerCount < 2 ? (
        <WaitingScreen count={playerCount} />
      ) : (
        <div style={{ display: "flex", gap: "20px" }}>
          <TurnControl />
          <Board />
          <PieceInfoPanel />
        </div>
      )}
    </div>
  );
};
