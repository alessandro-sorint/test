package dto

import "time"

type TriggerDto struct {
	IsStarted  bool       `json:"isStarted"`
	IsRunning  *bool      `json:"isRunning"`
	IsStopping *bool      `json:"isStopping"`
	LastRun    *time.Time `json:"lastRun"`
	TimeLeft   *uint      `json:"timeLeft"`
}

type TriggersStatusDto struct {
	OrganizationStatus      TriggerDto `json:"organizationStatus"`
	DiscoveryRunFailsStatus TriggerDto `json:"discoveryRunFailsStatus"`
	UserSynkStatus          TriggerDto `json:"userSynkStatus"`
}
