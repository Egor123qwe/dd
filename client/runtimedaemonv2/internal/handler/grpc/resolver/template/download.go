package template

import (
	"context"
	"fmt"
	"time"

	model "gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/handler/grpc/generate"
	"gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/model/docker/template"
	"gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/service/docker/api"
	"gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/service/util"
)

const (
	downloadInfoDelay = 500 * time.Millisecond
)

func (h Handler) DownloadStream(req *model.DownloadTemplateReq, sender model.Template_DownloadStreamServer) error {
	log.Infof("[HANDLER] download template started: %s", req.Template.Id)
	defer log.Infof("[HANDLER] download template finished: %s", req.Template.Id)

	srvReq := template.Template{
		ID:      req.Template.Id,
		Type:    req.Template.Type,
		Version: req.Template.Version,

		ImageName: req.Download.ImageName,
		ImageTag:  req.Download.ImageTag,

		Configuration: template.Configuration{
			UseGPU: req.Template.Configuration.UseGPU,

			Volumes: req.Template.Configuration.Volumes,
			Envs:    req.Template.Configuration.Envs,
		},

		Data: req.Template.Data,
	}

	for _, port := range req.Template.Configuration.Ports {
		port := template.Port{
			Port:     port.Port,
			Title:    port.Title,
			Protocol: template.Protocol(port.Protocol),

			AuthAvailable: port.AuthAvailable,
		}

		srvReq.Configuration.Ports = append(srvReq.Configuration.Ports, port)
	}

	resCh, err := h.srv.Download(sender.Context(), srvReq)
	if err != nil {
		return fmt.Errorf("failed to download template: %w", err)
	}

	if resCh == nil {
		return nil
	}

	var info api.PullInfo
	lastSend := time.Now()

	for info = range resCh {
		if time.Since(lastSend) > downloadInfoDelay {
			if err := sender.Send(h.convertDownloadEvent(info)); err != nil {
				return fmt.Errorf("failed to send download template response: %w", err)
			}

			lastSend = time.Now()
		}

		if info.Err != nil {
			break
		}
	}

	sender.Send(h.convertDownloadEvent(info))

	if info.Err != nil {
		return fmt.Errorf("failed to download template: %w", info.Err)
	}

	return nil
}

func (h Handler) convertDownloadEvent(event api.PullInfo) *model.DownloadTemplateStreamResp {
	result := &model.DownloadTemplateStreamResp{
		General: &model.DownloadTemplateStreamResp_Progress{
			Current: float32(event.FullProgress.Current / 1024),
			Total:   float32(event.FullProgress.Total / 1024),
			Percent: float32(util.SmartDivision(event.FullProgress.Current, event.FullProgress.Total) * 100),
		},

		Layers: make([]*model.DownloadTemplateStreamResp_Progress, len(event.LayersProgress)),
	}

	var ind int

	for _, layer := range event.LayersProgress {
		result.Layers[ind] = &model.DownloadTemplateStreamResp_Progress{
			Current: float32(layer.Current / 1024),
			Total:   float32(layer.Total / 1024),
			Percent: float32(util.SmartDivision(layer.Current, layer.Total) * 100),
		}

		ind++
	}

	if event.Err != nil {
		result.Err = util.Ptr(event.Err.Error())
	}

	return result
}

// Deprecated
func (h Handler) Download(ctx context.Context, req *model.DownloadTemplateReq) (*model.DownloadTemplateResp, error) {
	log.Infof("[HANDLER] download template started: %s", req.Template.Id)
	defer log.Infof("[HANDLER] download template finished: %s", req.Template.Id)

	srvReq := template.Template{
		ID:      req.Template.Id,
		Type:    req.Template.Type,
		Version: req.Template.Version,

		ImageName: req.Download.ImageName,
		ImageTag:  req.Download.ImageTag,

		Configuration: template.Configuration{
			UseGPU: req.Template.Configuration.UseGPU,

			Volumes: req.Template.Configuration.Volumes,
			Envs:    req.Template.Configuration.Envs,
		},

		Data: req.Template.Data,
	}

	for _, port := range req.Template.Configuration.Ports {
		port := template.Port{
			Port:     port.Port,
			Title:    port.Title,
			Protocol: template.Protocol(port.Protocol),

			AuthAvailable: port.AuthAvailable,
		}

		srvReq.Configuration.Ports = append(srvReq.Configuration.Ports, port)
	}

	if err := h.srv.Download_OLD(ctx, srvReq); err != nil {
		return nil, fmt.Errorf("failed to download template: %w", err)
	}

	return &model.DownloadTemplateResp{}, nil
}
