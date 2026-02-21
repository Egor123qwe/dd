package template

type Configuration struct {
	UseGPU bool

	Ports   []Port
	Volumes []string
	Envs    []string

	// Минимальные объёмы (для пропорционального распределения storage поставщика)
	MinStorageBytes       uint64   // минимум на контейнер (rootfs)
	MinVolumeStorageBytes []uint64 // минимум на каждый том, в порядке Volumes
}

type Template struct {
	ID      string
	Type    string
	Version string

	ImageName string
	ImageTag  string

	Configuration Configuration

	// info about template usages
	Usages map[int32]struct{}
	// debug
	Data string
}

type Info struct {
	Template Template

	ImageUsage    float64
	LocalMemUsage float64
	RentMemUsage  float64
}
