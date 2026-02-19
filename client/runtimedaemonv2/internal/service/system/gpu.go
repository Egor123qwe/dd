package system

import (
	"context"
	"fmt"

	"github.com/dustin/go-humanize"
	nvidiaInfoParser "github.com/fffaraz/nvidia-smi-json"
	"gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/model/hardware"
)

const (
	nvidiaExec = "nvidia-smi"
)

const (
	unknownData = "unknown"
)

var nvidiaErr = fmt.Errorf("this data not avalible for NVIDIA GPU. Please check nvidia drivers")

type nvidiaInfo struct {
	driverVersion string
	cards         []nvidiaCardInfo
}

type nvidiaCardInfo struct {
	name           string
	totalVMem      float64
	usedVMem       float64
	freeVMem       float64
	pciSubSystemID string
}

func (s service) GPU(ctx context.Context) (hardware.GPU, error) {
	var cards []hardware.Card

	nvidiaInfo, err := s.getNvidiaInfo(ctx)
	if err != nil {
		log.Warningf("failed to get nvidia infoFromCache: %v", err)
	}

	for _, card := range nvidiaInfo.cards {
		cards = append(cards, hardware.Card{
			IsNvidia: true,
			Name:     card.name,

			VMem: &hardware.VMem{
				Total: card.totalVMem,
				Used:  card.usedVMem,
				Free:  card.freeVMem,
			},
		})
	}

	nvidiaDriver := hardware.NvidiaDriver{
		Installed: err == nil,
		Version:   nvidiaInfo.driverVersion,
	}

	if err != nil {
		nvidiaDriver.Version = unknownData
	}

	result := hardware.GPU{
		NvidiaDriver: nvidiaDriver,
		Cards:        cards,
	}

	return result, nil
}

func (s service) getNvidiaInfo(ctx context.Context) (nvidiaInfo, error) {
	params := []string{
		nvidiaExec, "-q", "-x",
	}

	data, err := s.command.Run(ctx, params)
	if err != nil {
		return nvidiaInfo{}, nvidiaErr
	}

	info := nvidiaInfoParser.XmlToObject(data)
	if info == nil {
		return nvidiaInfo{}, fmt.Errorf("failed to parse nvidia infoFromCache")
	}

	result := nvidiaInfo{
		driverVersion: info.DriverVersion,
	}

	for _, gpu := range info.GPUS {
		total, err := humanize.ParseBytes(gpu.FbMemoryUsageTotal)
		if err != nil {
			total = 0
		}

		used, err := humanize.ParseBytes(gpu.FbMemoryUsageUsed)
		if err != nil {
			used = 0
		}

		free, err := humanize.ParseBytes(gpu.FbMemoryUsageFree)
		if err != nil {
			free = 0
		}

		result.cards = append(result.cards, nvidiaCardInfo{
			name:           gpu.ProductName,
			totalVMem:      bytesToKB(total),
			usedVMem:       bytesToKB(used),
			freeVMem:       bytesToKB(free),
			pciSubSystemID: gpu.PciSubSystemID,
		})
	}

	return result, nil
}
