package hardware

type CPU struct {
	Name       string  `json:"name"`
	Total      int     `json:"total"`
	Available  int     `json:"available"`
	Price      float64 `json:"price"`
	HardwareID string  `json:"hardware_id"`
}

type GPU struct {
	Name       string  `json:"name"`
	Total      int64   `json:"total"`
	Available  int64   `json:"available"`
	Used       int64   `json:"used"`
	Dlperf     float64 `json:"dlperf"`
	HardwareID string  `json:"hardware_id"`
	Price      float64 `json:"price"`
}

type Storage struct {
	Type       string  `json:"type"`
	Name       string  `json:"name"`
	Total      int64   `json:"total"`
	Available  int64   `json:"available"`
	Used       int64   `json:"used"`
	Bandwidth  float64 `json:"bandwidth"`
	HardwareID string  `json:"hardware_id"`
	Price      float64 `json:"price"`
}

type Prepull struct {
	ID         int64  `json:"id"`
	TemplateID string `json:"template_id"`
	HardwareID string `json:"hardware_id"`
}

type Session struct {
	ID             string    `json:"id"`
	CPUs           []CPU     `json:"cpus"`
	GPUs           []GPU     `json:"gpus"`
	StorageDevices []Storage `json:"storages"`
	TotalRAM       int64     `json:"total_ram"`
	AvailableRAM   int64     `json:"available_ram"`
	UsedRAM        int64     `json:"used_ram"`
	PriceRAM       float64   `json:"price_ram"`
	LoadSpeed      float64   `json:"load_speed"`
	UploadSpeed    float64   `json:"upload_speed"`
	Ping           int64     `json:"ping"`
	PriceInternet  float64   `json:"price_internet"`
}
