import { useEffect, useState } from "react";
import { GameStatus } from "@/api/status";
import { useGameSocket } from "@/store/GameSocket";
import { turnSecondsLeft } from "@/internal/game/turnClock";

export function useTurnSeconds() {
  const { status, turnDeadlineMS, turnRemainingMS, serverClockOffsetMS } = useGameSocket();
  const [now, setNow] = useState(Date.now);
  useEffect(() => {
    if (status !== GameStatus.playing) return;
    const timer = window.setInterval(() => setNow(Date.now()), 100);
    return () => window.clearInterval(timer);
  }, [status]);
  return turnSecondsLeft(status, turnDeadlineMS, turnRemainingMS, serverClockOffsetMS, now);
}
