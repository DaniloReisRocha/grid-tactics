import React from "react";
import { useGameSocket } from "@/store/GameSocket";

export const TurnControl: React.FC = () => {
  const { currentTurn, players } = useGameSocket();
  const player = players?.[currentTurn ?? ""];
  const endTurn = () => {};

  return (
    <div style={{ marginTop: "10px" }}>
      <div>Turno do: Player {player?.accoID}</div>
      <button onClick={endTurn} style={{ marginTop: "6px" }}>
        Finalizar Turno
      </button>
    </div>
  );
};
