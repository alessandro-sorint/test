package dto

import "time"

type TriggerRunTimeDto struct {
	Chan        chan TriggerMessage
	LastRun     time.Time
	IsRunning   bool
	IsStopping  bool
	TriggerTime uint
}

type TriggerMessage string

const (
	Restart TriggerMessage = "RESTART"
	Stop    TriggerMessage = "STOP"
)
