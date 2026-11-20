package models

type GameStatus string

const (
	GameStatusWaiting  GameStatus = "waiting"
	GameStatusPlaying  GameStatus = "playing"
	GameStatusPaused   GameStatus = "paused"
	GameStatusFinished GameStatus = "finished"
)
