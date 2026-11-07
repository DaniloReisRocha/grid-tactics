import { useEffect } from "react";
import { useGameSocket } from "@/store/GameSocket";

export function useStartGame(sessionID: string, accountID: string) {
  const joinGame = useGameSocket((state) => state.joinGame);
  const leaveGame = useGameSocket((state) => state.leaveGame);
  useEffect(() => {
    joinGame(sessionID, accountID);
    return () => leaveGame();
  }, [sessionID, accountID, joinGame, leaveGame]);
}
