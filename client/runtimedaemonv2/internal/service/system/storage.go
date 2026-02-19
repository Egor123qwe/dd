package system

import (
	"context"
	"fmt"

	"github.com/jaypipes/ghw"
	"github.com/shirou/gopsutil/disk"
	"gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/model/hardware"
	"gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/service/util"
)

const (
	unixMountPoint = "/"
)

type storageMem struct {
	Total uint64
	Used  uint64
	Free  uint64
}

func (s service) Storage(ctx context.Context) (hardware.Storage, error) {
	var result hardware.Storage

	types, err := s.getDiskTypes(ctx)
	if err != nil {
		return hardware.Storage{}, fmt.Errorf("failed to get disk types: %w", err)
	}

	parts, err := disk.PartitionsWithContext(ctx, true)
	if err != nil {
		return hardware.Storage{}, fmt.Errorf("failed to get storage infoFromCache: %w", err)
	}

	storageMem := storageMem{}

	for _, partition := range parts {
		// if it windows we use all mount points
		// if it unix we use unix only "/" mount point
		if partition.Mountpoint == unixMountPoint || util.IsWindows() {
			usage, err := disk.UsageWithContext(ctx, partition.Mountpoint)
			if err != nil {
				return hardware.Storage{}, fmt.Errorf("failed to get disk infoFromCache: %w", err)
			}

			storageMem.Total += usage.Total
			storageMem.Used += usage.Used
			storageMem.Free += usage.Free
		}
	}

	result = hardware.Storage{
		Types:    types,
		TotalMem: bytesToKB(storageMem.Total),
		UsedMem:  bytesToKB(storageMem.Used),
		FreeMem:  bytesToKB(storageMem.Free),
	}

	return result, nil
}

func (s service) getDiskTypes(ctx context.Context) ([]hardware.DiskType, error) {
	saved := s.cache.getStorageInfo()
	if saved != nil {
		return saved.Types, nil
	}

	blocks, err := ghw.Block()
	if err != nil {
		return nil, fmt.Errorf("failed to get blocks for disk types: %w", err)
	}

	var result []hardware.DiskType
	typesMap := make(map[hardware.DiskType]struct{})

	for _, d := range blocks.Disks {
		if d != nil {
			typesMap[hardware.DiskType(d.DriveType)] = struct{}{}
		}
	}

	for t := range typesMap {
		result = append(result, t)
	}

	s.cache.setStorageInfo(storageCache{Types: result})

	return result, nil
}
