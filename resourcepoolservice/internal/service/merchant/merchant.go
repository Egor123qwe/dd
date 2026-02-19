package merchant

import (
	"context"
	"math"
	"strings"
	"time"

	"gitlab.roy9.ru/roy9/backend/statemachine/resourcepoolservice/internal/config"
	"gitlab.roy9.ru/roy9/backend/statemachine/resourcepoolservice/internal/domain/category"
	"gitlab.roy9.ru/roy9/backend/statemachine/resourcepoolservice/internal/domain/hardware"
	"gitlab.roy9.ru/roy9/backend/statemachine/resourcepoolservice/internal/domain/sharep2p"
	pricingv1 "gitlab.roy9.ru/roy9/backend/statemachine/resourcepoolservice/internal/proto/gen/pricing.v1"

	"github.com/google/uuid"
	"google.golang.org/grpc"
)

const (
	requestTimeout  = 5 * time.Second
	sessionIDPrefix = "session_id"
)

type Service struct {
	merchantRepo MerchantRepository
	cache        Cache
	cfg          config.RedisConfig
	priceService PriceClient
}

type MerchantRepository interface {
	Create(ctx context.Context, hw *pricingv1.HardwareResponse, prepull []hardware.Prepull, userID, connectionID, nodeName string) error
	Ready(ctx context.Context, sessionID string) error
	Stop(ctx context.Context, sessionID, deletionReason string) error
	GetPricingConfig(ctx context.Context) (category.PricingConfig, error)
	GetCPUByName(ctx context.Context, name string) (*category.CPUDict, error)
	InsertCPU(ctx context.Context, name string) error
	GetGPUBestMatch(ctx context.Context, name string) (*category.GPUDict, bool)
	InsertGPU(ctx context.Context, name string, totalVram int64, avgDlperf float64) error
}

type PriceClient interface {
	Evaluate(ctx context.Context, in *pricingv1.HardwareRequest, opts ...grpc.CallOption) (*pricingv1.HardwareResponse, error)
}

type Cache interface {
	Get(ctx context.Context, key string) (any, error)
	Set(ctx context.Context, key string, value any, exp time.Duration) error
	SetXX(ctx context.Context, key string, value any, exp time.Duration) (bool, error)
	Del(ctx context.Context, key string) error
}

func New(merchantrepo MerchantRepository, cache Cache, cfg config.RedisConfig, priceService PriceClient) *Service {
	service := &Service{
		merchantRepo: merchantrepo,
		cache:        cache,
		cfg:          cfg,
		priceService: priceService,
	}

	return service
}

func (s *Service) ShareP2PInit(ctx context.Context, req sharep2p.InitMerchantRequest) (string, float32, error) {
	sessionID := uuid.New().String()
	req.Content.Hardware.ID = sessionID

	rpcReq := &pricingv1.HardwareRequest{
		Id:           req.Content.Hardware.ID,
		LoadSpeed:    int32(req.Content.Hardware.LoadSpeed),
		UploadSpeed:  int32(req.Content.Hardware.UploadSpeed),
		Ping:         int32(req.Content.Hardware.Ping),
		TotalRam:     req.Content.Hardware.TotalRAM,
		AvailableRam: req.Content.Hardware.AvailableRAM,
		UsedRam:      req.Content.Hardware.UsedRAM,
	}

	for _, cpu := range req.Content.Hardware.CPUs {
		rpcReq.Cpus = append(rpcReq.Cpus, &pricingv1.CPURequest{
			Available: int32(cpu.Available),
			Name:      cpu.Name,
			Total:     int32(cpu.Total),
		})
	}

	for _, gpu := range req.Content.Hardware.GPUs {
		rpcReq.Gpus = append(rpcReq.Gpus, &pricingv1.GPURequest{
			Name:      gpu.Name,
			Available: gpu.Available,
			Used:      gpu.Used,
			Total:     gpu.Total,
			Dlperf:    float32(gpu.Dlperf),
		})
	}

	for _, storage := range req.Content.Hardware.StorageDevices {
		rpcReq.Storages = append(rpcReq.Storages, &pricingv1.StorageRequest{
			Type:      storage.Type,
			Available: storage.Available,
			Used:      storage.Used,
			Total:     storage.Total,
			Name:      storage.Name,
			Bandwidth: float32(storage.Bandwidth),
		})
	}

	c, cancel := context.WithTimeout(ctx, requestTimeout)
	defer cancel()

	hw, err := s.priceService.Evaluate(c, rpcReq)
	if err != nil {
		return "", 0, err
	}

	if err := s.evaluatePrice(ctx, hw); err != nil {
		return "", 0, err
	}

	err = s.merchantRepo.Create(ctx, hw, req.Content.Prepull, req.Meta.Conn.UserID, req.Meta.Conn.ConnectionID, req.Content.NodeName)
	if err != nil {
		return "", 0, err
	}

	return sessionID, hw.TotalPrice, nil
}

// evaluatePrice fills hw.TotalPrice (per minute, BYN) and per-component prices. Returns ErrUnidentifiedHardware if any CPU/GPU is unknown (then it is added with price 0 and session must not start).
func (s *Service) evaluatePrice(ctx context.Context, hw *pricingv1.HardwareResponse) error {
	cfg, err := s.merchantRepo.GetPricingConfig(ctx)
	if err != nil {
		return err
	}

	for _, c := range hw.Cpus {
		cpu, err := s.merchantRepo.GetCPUByName(ctx, c.Name)
		if err != nil {
			return err
		}
		if cpu == nil {
			_ = s.merchantRepo.InsertCPU(ctx, c.Name)
			return sharep2p.ErrUnidentifiedHardware
		}
		if cpu.PricePerMinute == 0 {
			return sharep2p.ErrUnidentifiedHardware
		}
		c.Price = float32(cpu.PricePerMinute)
	}

	for _, g := range hw.Gpus {
		gpu, found := s.merchantRepo.GetGPUBestMatch(ctx, g.Name)
		if !found {
			_ = s.merchantRepo.InsertGPU(ctx, g.Name, g.Total, float64(g.Dlperf))
			return sharep2p.ErrUnidentifiedHardware
		}
		if gpu.Price == 0 {
			return sharep2p.ErrUnidentifiedHardware
		}
		g.Price = float32(gpu.Price)
	}

	ramGB := float64(hw.TotalRam) / (1024 * 1024)
	var storagePart float64
	for _, st := range hw.Storages {
		gb := float64(st.Total) / (1024 * 1024)
		if strings.ToUpper(strings.TrimSpace(st.Type)) == "SSD" {
			storagePart += gb * cfg.StorageSSDPerGBPerMinute
		} else {
			storagePart += gb * cfg.StorageHDDPerGBPerMinute
		}
	}
	internetMbit := (float64(hw.LoadSpeed) + float64(hw.UploadSpeed)) / 2 / 1024
	internetPart := internetMbit * cfg.InternetPerMbitPerMinute

	total := cfg.BasePerMinute +
		ramGB*cfg.RAMPerGBPerMinute +
		storagePart +
		internetPart
	for _, c := range hw.Cpus {
		total += float64(c.Price)
	}
	for _, g := range hw.Gpus {
		total += float64(g.Price)
	}

	// Округляем до сотых (2 знака после запятой)
	total = math.Round(total*100) / 100
	hw.TotalPrice = float32(total)
	hw.PriceRam = float32(math.Round(cfg.RAMPerGBPerMinute*100) / 100)
	hw.PriceInternet = float32(math.Round(cfg.InternetPerMbitPerMinute*100) / 100)
	return nil
}

func (s *Service) ShareP2Pready(ctx context.Context, sessionID string) error {
	return s.merchantRepo.Ready(ctx, sessionID)
}

func (s *Service) ShareP2PStop(ctx context.Context, sessionID string) error {
	return s.merchantRepo.Stop(ctx, sessionID, "stopped by merchant")
}
