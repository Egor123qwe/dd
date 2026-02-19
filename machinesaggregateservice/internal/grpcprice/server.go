package grpcprice

import (
	"context"

	pricingv1 "gitlab.roy9.ru/roy9/backend/statemachine/machinesaggregateservice/internal/pricingv1"
)

type Server struct {
	pricingv1.UnimplementedPriceServer
}

func New() *Server {
	return &Server{}
}

func (s *Server) Evaluate(ctx context.Context, req *pricingv1.HardwareRequest) (*pricingv1.HardwareResponse, error) {
	resp := &pricingv1.HardwareResponse{
		Id:            req.Id,
		LoadSpeed:     req.LoadSpeed,
		UploadSpeed:   req.UploadSpeed,
		Ping:          req.Ping,
		TotalRam:      req.TotalRam,
		AvailableRam:  req.AvailableRam,
		UsedRam:       req.UsedRam,
		TotalPrice:    0,
		Status:        "ok",
		PriceInternet: 0,
		PriceRam:      0,
	}
	for _, c := range req.Cpus {
		resp.Cpus = append(resp.Cpus, &pricingv1.CPUResponse{
			Id:        "",
			Available: c.Available,
			Name:      c.Name,
			Total:     c.Total,
			Price:     0,
		})
	}
	for _, g := range req.Gpus {
		resp.Gpus = append(resp.Gpus, &pricingv1.GPUResponse{
			Id:        "",
			Name:      g.Name,
			Available: g.Available,
			Used:      g.Used,
			Total:     g.Total,
			Dlperf:    g.Dlperf,
			Price:     0,
		})
	}
	for _, st := range req.Storages {
		resp.Storages = append(resp.Storages, &pricingv1.StorageResponse{
			Id:        "",
			Type:      st.Type,
			Available: st.Available,
			Used:      st.Used,
			Total:     st.Total,
			Name:      st.Name,
			Bandwidth: st.Bandwidth,
			Price:     0,
		})
	}
	return resp, nil
}
