package template

import (
	"context"
	"fmt"

	model "gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/handler/grpc/generate"
)

func (h Handler) Get(ctx context.Context, req *model.GetTemplateReq) (*model.GetTemplateResp, error) {
	result, err := h.srv.GetFiltredWithStat(ctx, req.Types)
	if err != nil {
		return nil, fmt.Errorf("failed to get allowed template: %w", err)
	}

	var templates []*model.GetTemplateResp_TemplateInfo

	for _, template := range result {
		t := &model.GetTemplateResp_TemplateInfo{
			Template: &model.TemplateData{
				Id:      template.Template.ID,
				Type:    template.Template.Type,
				Version: template.Template.Version,

				Configuration: &model.TemplateData_Configuration{
					UseGPU: template.Template.Configuration.UseGPU,

					Volumes: template.Template.Configuration.Volumes,
					Envs:    template.Template.Configuration.Envs,
				},

				Data: template.Template.Data,
			},

			ImageUsage: float32(template.ImageUsage),

			LocalUsage: &model.GetTemplateResp_TemplateInfo_LocalUsage{
				MemUsage: float32(template.LocalMemUsage),
			},

			RentUsage: &model.GetTemplateResp_TemplateInfo_RentUsage{
				MemUsage: float32(template.RentMemUsage),
			},
		}

		for _, port := range template.Template.Configuration.Ports {
			p := &model.TemplatePort{
				Port:     port.Port,
				Title:    port.Title,
				Protocol: model.PortProtocol(port.Protocol),

				AuthAvailable: port.AuthAvailable,
			}

			t.Template.Configuration.Ports = append(t.Template.Configuration.Ports, p)
		}

		templates = append(templates, t)
	}

	return &model.GetTemplateResp{Templates: templates}, nil
}
