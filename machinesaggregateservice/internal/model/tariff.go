package model

type TariffName string

const (
	Effective TariffName = "effective"
	Optimal   TariffName = "optimal"
)

type Tariff struct {
	ID            string   `json:"id"`
	Name          string   `json:"type,omitempty" db:"type"`
	Price         float32  `json:"price,omitempty"`
	GPUMeta       GPUDict  `json:"gpu"`
	TotalSessions int      `json:"total_sessions" db:"total_sessions"`
	ListSessionID []string `json:"sessions"`
}
