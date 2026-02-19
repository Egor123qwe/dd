package system

import (
	"context"

	"github.com/pbnjay/memory"
	"gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/model/hardware"
)

func (s service) RAM(ctx context.Context) hardware.RAM {
	total := memory.TotalMemory()
	free := memory.FreeMemory()
	used := total - free

	return hardware.RAM{
		TotalMem: bytesToKB(total),
		UsedMem:  bytesToKB(used),
		FreeMem:  bytesToKB(free),
	}
}
