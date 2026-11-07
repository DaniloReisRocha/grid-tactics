import React from "react";
import { useGameSocket } from "@/store/GameSocket";
import type { UnitInstance } from "@/api/models";
import "@/styles/game/PieceInfoPanel.css";

export const PieceInfoPanel: React.FC = () => {
  const {
    // units,
    // pieceID,
    // selectPiece,
    // currentPlayer,
    // setAttackMode,
    // attackMode,
    players,
  } = useGameSocket();
  const units = Object.values(players).reduce((p1, p2) => {
    return p1.concat(p2?.units ?? []);
  }, [] as UnitInstance[]);
  // const selectedPiece = units.find((p) => p.unit.id === pieceID);

  const selectPiece = (id: string) => {};

  return (
    <div className="unit-info-panel">
      {/* {selectedPiece ? (
        <div className="selected-unit">
          <div className="icon">{selectedPiece.unit.icon}</div>
          <div className="player">Player: {selectedPiece.player}</div>
          <div className="stats">
            <div>LP: {selectedPiece.currentLife}</div>
            <div>Shield: {selectedPiece.currentShield}</div>
            <div>Stamina: {selectedPiece.currentStamina}</div>
            {selectedPiece.player === currentPlayer && (
              <button
                className={`attack-btn ${attackMode ? "pressed" : ""}`}
                onClick={() => setAttackMode(!attackMode)}
              >
                Atacar
              </button>
            )}
          </div>
        </div>
      ) : (
        <div className="no-selection">Nenhuma peça selecionada</div>
      )} */}

      <div className="other-units-list">
        {units.map((p) => (
          <div
            key={p.unit.id}
            className={`other-unit ${
              ""
              // selectedPiece?.unit.id === p.unit.id ? "selected" : ""
            }`}
            onClick={() => selectPiece(p.unit.id)}
          >
            <span>{p.unit.icon}</span> | Player {p.player}
          </div>
        ))}
      </div>
    </div>
  );
};
