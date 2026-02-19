package rent

import (
	"encoding/json"
	"os"

	hardwareModel "gitlab.roy9.ru/roy9/backend/clientside/merchantclient/internal/model/msg/content/hardware"
	"gitlab.roy9.ru/roy9/backend/clientside/merchantclient/internal/model/msg/content/mockHardware"
	rdModel "gitlab.roy9.ru/roy9/backend/clientside/merchantclient/pkg/proto/runtimedaemon/generate"
)

const (
	configurationPath = "./runtime.conf"
)

func (u usecase) newConfiguration(hardware *rdModel.SystemInfo, useMock bool) (hardwareModel.Session, error) {
	var result hardwareModel.Session

	if hardware != nil {
		result = u.daemonConfiguration(hardware)
	}

	if useMock {
		mock, err := u.mockConfiguration()
		if err != nil {
			return result, err
		}

		result = u.mergeConfigurations(result, mock)
	}

	return result, nil
}

func (u usecase) mockConfiguration() (mockHardware.Hardware, error) {
	var result mockHardware.Hardware

	data, err := os.ReadFile(configurationPath)
	if err != nil {
		return mockHardware.Hardware{}, err
	}

	err = json.Unmarshal(data, &result)
	if err != nil {
		return mockHardware.Hardware{}, err
	}

	return result, nil
}

func (u usecase) daemonConfiguration(hardware *rdModel.SystemInfo) hardwareModel.Session {
	var result hardwareModel.Session

	result = hardwareModel.Session{
		// network
		Ping:        int64(hardware.Network.Ping),
		LoadSpeed:   float64(hardware.Network.Download) * 1024,
		UploadSpeed: float64(hardware.Network.Upload) * 1024,

		// ram
		TotalRAM:     int64(hardware.Ram.TotalMem),
		AvailableRAM: int64(hardware.Ram.FreeMem),
		UsedRAM:      int64(hardware.Ram.UsedMem),
	}

	result.CPUs = append(result.CPUs, hardwareModel.CPU{
		Name:      hardware.Cpu.Name,
		Total:     int(hardware.Cpu.CoresCount),
		Available: int(hardware.Cpu.CoresCount),
	})

	result.StorageDevices = append(result.StorageDevices, hardwareModel.Storage{
		Total:     int64(hardware.Storage.TotalMem),
		Available: int64(hardware.Storage.FreeMem),
		Used:      int64(hardware.Storage.UsedMem),
	})

	if len(hardware.Storage.Types) > 0 {
		result.StorageDevices[0].Type = hardwareModel.DiskTypeName[hardwareModel.DiskType(hardware.Storage.Types[0])]
	}

	for _, card := range hardware.Gpu.Cards {
		result.GPUs = append(result.GPUs, hardwareModel.GPU{
			Name:      card.Name,
			Total:     int64(card.Mem.Total),
			Available: int64(card.Mem.Free),
			Used:      int64(card.Mem.Used),
		})
	}

	return result
}

func (u usecase) mergeConfigurations(hardware hardwareModel.Session, mock mockHardware.Hardware) hardwareModel.Session {
	if mock.RAM != nil {
		hardware.TotalRAM = mock.RAM.TotalRAM
		hardware.AvailableRAM = mock.RAM.AvailableRAM
		hardware.UsedRAM = mock.RAM.UsedRAM
	}

	if mock.Network != nil {
		hardware.LoadSpeed = mock.Network.LoadSpeed
		hardware.UploadSpeed = mock.Network.UploadSpeed
		hardware.Ping = mock.Network.Ping
	}

	if len(mock.CPUs) > 0 {
		hardware.CPUs = make([]hardwareModel.CPU, 0, len(mock.CPUs))

		for _, cpu := range mock.CPUs {
			hardware.CPUs = append(hardware.CPUs, hardwareModel.CPU{
				Name:      cpu.Name,
				Total:     cpu.Total,
				Available: cpu.Total,
			})
		}
	}

	if len(mock.GPUs) > 0 {
		hardware.GPUs = make([]hardwareModel.GPU, 0, len(mock.GPUs))

		for _, gpu := range mock.GPUs {
			hardware.GPUs = append(hardware.GPUs, hardwareModel.GPU{
				Name:      gpu.Name,
				Total:     gpu.Total,
				Available: gpu.Total,
				Used:      gpu.Used,
				Dlperf:    gpu.Dlperf,
			})
		}
	}

	if len(mock.Storages) > 0 {
		hardware.StorageDevices = make([]hardwareModel.Storage, 0, len(mock.Storages))

		for _, storage := range mock.Storages {
			hardware.StorageDevices = append(hardware.StorageDevices, hardwareModel.Storage{
				Name:      storage.Name,
				Type:      storage.Type,
				Total:     storage.Total,
				Available: storage.Total,
				Used:      storage.Used,
				Bandwidth: storage.Bandwidth,
			})
		}
	}

	return hardware
}
