package model

type GPU struct {
	ID            string  `json:"id"`
	Name          string  `json:"name"`
	GPUDict       GPUDict `json:"-"`
	TotalVRAM     int64   `json:"total_vram"`
	AvailableVRAM int     `json:"available_vram"`
	UsedVRAM      int     `json:"used_vram"`
	Dlperf        float32 `json:"avg_dlperf"`
	Price         float32 `json:"price,omitempty"`
}
type CPU struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	Total     int     `json:"total"`
	Available int     `json:"available"`
	Price     float32 `json:"price"`
}

type Storage struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	Type      string  `json:"type"`
	Total     int64   `json:"total"`
	Available int     `json:"available"`
	Used      int     `json:"used"`
	Bandwidth float32 `json:"bandwidth"`
	Price     float32 `json:"price"`
}

type PrePull struct {
	ID         string `json:"id"`
	TemplateId string `json:"template_id"`
}

type Session struct {
	ID            string    `json:"session_id"`
	TotalRam      int64     `json:"total_ram"`
	AvailableRam  int64     `json:"available_ram"`
	UsedRam       int64     `json:"used_ram"`
	PriceRam      float32   `json:"price_ram"`
	LoadSpeed     float32   `json:"load_speed"`
	UploadSpeed   float32   `json:"upload_speed"`
	Ping          float32   `json:"ping"`
	PriceInternet float32   `json:"price_internet"`
	CreatedAt     string    `json:"created_at,omitempty"`
	DeletedAt     string    `json:"deleted_at,omitempty"`
	GPUs          []GPU     `json:"gpus"`
	CPUs          []CPU     `json:"cpus"`
	Storage       []Storage `json:"storages"`
	PrePull       []PrePull `json:"prepull"`
	TotalPrice    float32   `json:"total_price"`
}

type GPUDict struct {
	ID            string  `json:"id,omitempty"`
	Name          string  `json:"name,omitempty"`
	TotalVRAM     float32 `json:"total_vram" db:"total_vram"`
	TotalSessions int     `json:"total_sessions,omitempty" db:"total_sessions"`
	Price         float32 `json:"price,omitempty" db:"price"`
}
