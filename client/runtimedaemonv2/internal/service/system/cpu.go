package system

import (
	"context"

	. "github.com/klauspost/cpuid/v2"
	"github.com/shirou/gopsutil/v3/cpu"
	"gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/model/hardware"
)

func (s service) CPU(ctx context.Context) hardware.CPU {
	cores, err := cpu.Counts(true)
	if err != nil {
		log.Errorf("failed to get cpu cores count: %s", err)
		cores = CPU.PhysicalCores
	}

	return hardware.CPU{
		Name:       CPU.BrandName,
		CoresCount: uint32(cores),
	}
}
