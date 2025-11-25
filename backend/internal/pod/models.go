package pod

import "time"

// EventInfo represents pod event information as a DTO
type EventInfo struct {
	Type      string    `json:"type"`
	Reason    string    `json:"reason"`
	Message   string    `json:"message"`
	Count     int32     `json:"count"`
	FirstSeen time.Time `json:"firstSeen"`
	LastSeen  time.Time `json:"lastSeen"`
	Source    string    `json:"source,omitempty"`
}
