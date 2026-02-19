package model

type SessionResponse struct {
	Category Category `json:"category"`
	Session  Session  `json:"session"`
}

type SessionList struct {
	Sessions   []Session `json:"sessions"`
	MaxRAM     int64     `json:"max_ram,omitempty"`
	MinRAM     int64     `json:"min_ram,omitempty"`
	MinStorage int64     `json:"min_storage,omitempty"`
	MaxStorage int64     `json:"max_storage,omitempty"`
}
