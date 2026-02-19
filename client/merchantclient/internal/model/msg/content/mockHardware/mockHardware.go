package mockHardware

type Hardware struct {
	CPUs     []CPU     `json:"cpus"`
	GPUs     []GPU     `json:"gpus"`
	Storages []Storage `json:"storages"`
	Network  *Network  `json:"network,omitempty"`
	RAM      *RAM      `json:"ram,omitempty"`
}

type CPU struct {
	Name      string `json:"name"`
	Total     int    `json:"total"`
	Available int    `json:"available"`
}

type GPU struct {
	Name      string  `json:"name"`
	Total     int64   `json:"total"`
	Available int64   `json:"available"`
	Used      int64   `json:"used"`
	Dlperf    float64 `json:"dlperf"`
}

type Storage struct {
	Name      string  `json:"name"`
	Type      string  `json:"type"`
	Total     int64   `json:"total"`
	Available int64   `json:"available"`
	Used      int64   `json:"used"`
	Bandwidth float64 `json:"bandwidth"`
}

type Network struct {
	LoadSpeed   float64 `json:"load_speed"`
	UploadSpeed float64 `json:"upload_speed"`
	Ping        int64   `json:"ping"`
}

type RAM struct {
	TotalRAM     int64 `json:"total_ram"`
	AvailableRAM int64 `json:"available_ram"`
	UsedRAM      int64 `json:"used_ram"`
}
