package app

import (
	"context"
	"errors"
	"sync"

	"github.com/op/go-logging"
	proto "gitlab.roy9.ru/roy9/backend/clientside/merchantclient/pkg/proto/runtimedaemon/generate"
	"gitlab.roy9.ru/roy9/backend/clientside/merchantclient/internal/service/runtimedaemon"
	"gitlab.roy9.ru/roy9/backend/clientside/merchantclient/internal/service/usecase/rent"
)

var backendLog = logging.MustGetLogger("app.backend")

var (
	ErrTokenRequired    = errors.New("token is required")
	ErrAlreadyConnected = errors.New("already connected, disconnect first")
)

// WebBackend holds optional connected app for the web UI (connect via token, then start/stop rent).
type WebBackend struct {
	rd runtimedaemon.API

	mu          sync.RWMutex
	app         *App
	serveCancel context.CancelFunc
}

func NewWebBackend(rd runtimedaemon.API) *WebBackend {
	return &WebBackend{rd: rd}
}

func (b *WebBackend) Connect(ctx context.Context, token string) error {
	if token == "" {
		return ErrTokenRequired
	}

	b.mu.Lock()
	if b.app != nil {
		b.mu.Unlock()
		return ErrAlreadyConnected
	}
	b.mu.Unlock()

	a, err := NewWithRD(token, "", b.rd)
	if err != nil {
		return err
	}

	serveCtx, cancel := context.WithCancel(ctx)
	go func() {
		if err := a.ServeWS(serveCtx); err != nil && serveCtx.Err() == nil {
			backendLog.Warningf("ws serve stopped: %v", err)
		}
	}()

	b.mu.Lock()
	b.app = &a
	b.serveCancel = cancel
	b.mu.Unlock()

	return nil
}

func (b *WebBackend) Disconnect() {
	b.mu.Lock()
	app := b.app
	cancel := b.serveCancel
	b.app = nil
	b.serveCancel = nil
	b.mu.Unlock()

	if cancel != nil {
		cancel()
	}
	if app != nil {
		_ = app.Close()
		backendLog.Info("disconnected")
	}
}

func (b *WebBackend) Rent() (rent.Usecase, bool) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	if b.app == nil {
		return nil, false
	}
	return b.app.Rent(), true
}

func (b *WebBackend) GetDockerInfo(ctx context.Context) (available bool, version string, err error) {
	info, err := b.rd.GetInfo(ctx, &proto.InfoReq{})
	if err != nil {
		return false, "", err
	}
	d := info.GetDocker()
	if d == nil {
		return false, "", nil
	}
	return d.GetAvailable(), d.GetVersion(), nil
}

func (b *WebBackend) GetHardware(ctx context.Context) (map[string]interface{}, error) {
	hw, err := b.rd.GetHardware(ctx)
	if err != nil {
		return nil, err
	}
	storageType := "SSD"
	if len(hw.Storage.Types) > 0 {
		if name, ok := proto.Storage_DiskType_name[int32(hw.Storage.Types[0])]; ok {
			storageType = name
		}
	}
	// Преобразуем SystemInfo в JSON-совместимую структуру
	result := map[string]interface{}{
		"ram": map[string]interface{}{
			"total":     hw.Ram.TotalMem,     // KB
			"available": hw.Ram.FreeMem,     // KB
			"used":      hw.Ram.UsedMem,      // KB
		},
		"network": map[string]interface{}{
			"load_speed":   float64(hw.Network.Download), // Мбит/с или Kbit/s от демона — отдаём как есть
			"upload_speed": float64(hw.Network.Upload),
			"ping":         hw.Network.Ping,
		},
		"cpus": []map[string]interface{}{
			{
				"name":      hw.Cpu.Name,
				"total":     int(hw.Cpu.CoresCount),
				"available": int(hw.Cpu.CoresCount),
			},
		},
		"storages": []map[string]interface{}{
			{
				"type":      storageType,
				"name":      "Основной диск",
				"total":     hw.Storage.TotalMem,     // KB
				"available": hw.Storage.FreeMem,       // KB
			},
		},
		"gpus": []map[string]interface{}{},
	}
	for _, card := range hw.Gpu.Cards {
		gpuInfo := map[string]interface{}{
			"name": card.Name,
		}
		if card.Mem != nil {
			gpuInfo["total_vram"] = card.Mem.Total     // KB
			gpuInfo["available_vram"] = card.Mem.Free  // KB
			gpuInfo["used_vram"] = card.Mem.Used        // KB
		}
		result["gpus"] = append(result["gpus"].([]map[string]interface{}), gpuInfo)
	}
	return result, nil
}

// RentOnlyBackend adapts a single rent.Usecase as a Backend (always "connected", no Connect/Disconnect).
// Used when the app is started with a token (e.g. legacy Start flow).
type RentOnlyBackend struct {
	RentUC rent.Usecase
	RD     runtimedaemon.API
}

func (b RentOnlyBackend) Rent() (rent.Usecase, bool) {
	return b.RentUC, b.RentUC != nil
}

func (b RentOnlyBackend) Connect(ctx context.Context, token string) error {
	return ErrAlreadyConnected
}

func (b RentOnlyBackend) Disconnect() {}

func (b RentOnlyBackend) GetDockerInfo(ctx context.Context) (available bool, version string, err error) {
	if b.RD == nil {
		return false, "", nil
	}
	info, err := b.RD.GetInfo(ctx, &proto.InfoReq{})
	if err != nil {
		return false, "", err
	}
	d := info.GetDocker()
	if d == nil {
		return false, "", nil
	}
	return d.GetAvailable(), d.GetVersion(), nil
}

func (b RentOnlyBackend) GetHardware(ctx context.Context) (map[string]interface{}, error) {
	if b.RD == nil {
		return nil, errors.New("hardware info not available")
	}
	hw, err := b.RD.GetHardware(ctx)
	if err != nil {
		return nil, err
	}
	storageType := "SSD"
	if len(hw.Storage.Types) > 0 {
		if name, ok := proto.Storage_DiskType_name[int32(hw.Storage.Types[0])]; ok {
			storageType = name
		}
	}
	result := map[string]interface{}{
		"ram": map[string]interface{}{
			"total":     hw.Ram.TotalMem,     // KB
			"available": hw.Ram.FreeMem,      // KB
			"used":      hw.Ram.UsedMem,      // KB
		},
		"network": map[string]interface{}{
			"load_speed":   float64(hw.Network.Download),
			"upload_speed": float64(hw.Network.Upload),
			"ping":         hw.Network.Ping,
		},
		"cpus": []map[string]interface{}{
			{
				"name":      hw.Cpu.Name,
				"total":     int(hw.Cpu.CoresCount),
				"available": int(hw.Cpu.CoresCount),
			},
		},
		"storages": []map[string]interface{}{
			{
				"type":      storageType,
				"name":      "Основной диск",
				"total":     hw.Storage.TotalMem,     // KB
				"available": hw.Storage.FreeMem,      // KB
			},
		},
		"gpus": []map[string]interface{}{},
	}
	for _, card := range hw.Gpu.Cards {
		gpuInfo := map[string]interface{}{
			"name": card.Name,
		}
		if card.Mem != nil {
			gpuInfo["total_vram"] = card.Mem.Total     // KB
			gpuInfo["available_vram"] = card.Mem.Free  // KB
			gpuInfo["used_vram"] = card.Mem.Used       // KB
		}
		result["gpus"] = append(result["gpus"].([]map[string]interface{}), gpuInfo)
	}
	return result, nil
}
