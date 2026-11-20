import { GameStatus } from "@/api/status";


const TURN_SECONDS = 45;

export function turnSecondsLeft(status: GameStatus, deadlineMS: number, pausedRemainingMS: number, serverOffsetMS: number, nowMS: number) {
  const remaining = status === GameStatus.paused ? pausedRemainingMS
    : status === GameStatus.playing && deadlineMS > 0 ? deadlineMS - (nowMS + serverOffsetMS) : 0;
  return Math.min(TURN_SECONDS, Math.max(0, Math.ceil(remaining / 1000)));
}
