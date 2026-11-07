// src/services/GameSocket.ts
const GameSocket = new WebSocket("ws://localhost:8080/api/game-connect");

const MessageType = {
  echo: "echo",
  joinGame: "joinGame",
  movePiece: "movePiece",
  actionFailed: "actionFailed",
  gameStateUpdate: "gameStateUpdate",
} as const;

type Message = {
  type: keyof typeof MessageType;
  payload: any;
};

GameSocket.onmessage = (event) => {
  const { type, payload }: Message = JSON.parse(event.data);

  console.log("type => ", type);
  console.log("payload => ", payload);

  switch (type) {
    case MessageType.gameStateUpdate:
      // TODO: atualizar Zustand ou Context com o novo estado
      break;
    case MessageType.joinGame:
      break;
    case MessageType.movePiece:
      break;
    case MessageType.actionFailed:
      console.error(payload.message);
      break;
  }
};

const joinGame = (sessionID: string, accoID: string, pieces: any[]) => {
  GameSocket.send(
    JSON.stringify({
      type: MessageType.joinGame,
      payload: {
        sessionID,
        accoID,
        pieces,
      },
    })
  );
};

export { MessageType, GameSocket, joinGame };
