package hardware

type DiskType int32

const (
	Unknown DiskType = iota
	HDD
	FDD
	ODD
	SSD
	Virtual
)

type Info struct {
	GPU     GPU
	Storage Storage
	RAM     RAM
	CPU     CPU
	Network Network
}

type GPU struct {
	NvidiaDriver NvidiaDriver
	Cards        []Card
}

type NvidiaDriver struct {
	Installed bool
	Version   string
}

type Card struct {
	IsNvidia bool
	Name     string
	VMem     *VMem
}

type VMem struct {
	Total float64
	Used  float64
	Free  float64
}

type Storage struct {
	Types    []DiskType
	TotalMem float64
	UsedMem  float64
	FreeMem  float64
}

type RAM struct {
	TotalMem float64
	UsedMem  float64
	FreeMem  float64
}

type CPU struct {
	Name       string
	CoresCount uint32
}

type Network struct {
	Ping     float64
	Download float64
	Upload   float64
}
