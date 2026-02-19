package session

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
	"gitlab.roy9.ru/roy9/backend/clientside/merchantclient/internal/model/event"
	"gitlab.roy9.ru/roy9/backend/clientside/merchantclient/internal/model/msg"
	"gitlab.roy9.ru/roy9/backend/clientside/merchantclient/internal/model/msg/content/session"
	settingsModel "gitlab.roy9.ru/roy9/backend/clientside/merchantclient/internal/model/msg/content/session/settings"
	"gitlab.roy9.ru/roy9/backend/clientside/merchantclient/internal/service/usecase/state"
	proto "gitlab.roy9.ru/roy9/backend/clientside/merchantclient/pkg/proto/runtimedaemon/generate"
)

func (u usecase) Start(ctx context.Context, req session.MerchantRentStartReq) error {
	ctx, cancel := context.WithCancel(ctx)
	u.startController.SetCancelFn(cancel)

	u.state.Mutex().Lock()
	defer u.state.Mutex().Unlock()

	u.state.SetStatus(state.Configuring)
	log.Infof("client [%s] connected", req.ClientID)
	log.Infof("session [%s] in configuring...", req.RequestID)

	reqData := session.RentRequestStatusUpdatedReq{
		RequestID: req.RequestID,
	}

	fmt.Println("req id: ", reqData.RequestID)

	if err := u.configure(ctx, req.ClientID, req.Settings); err != nil {
		u.state.SetStatus(state.Ready)
		reqData.Status = session.ErrorMerchantRentStatus
		req := msg.Marshal(string(event.RentRequestStatusUpdatedEvent), msg.Meta{MessageID: uuid.New().String()}, reqData)

		if err := u.wsConn.Writer().Write(ctx, req); err != nil {
			log.Errorf("failed to send merchant stop request: %s", err)
		}

		return fmt.Errorf("failed to configuring session: %w", err)
	}

	reqData.Status = session.RunningMerchantRentStatus
	srvReq := msg.Marshal(string(event.RentRequestStatusUpdatedEvent), msg.Meta{MessageID: uuid.New().String()}, reqData)

	resp, err := u.wsConn.MsgConn.Do(ctx, u.wsConn.Writer().Write, srvReq)
	if err != nil {
		u.state.SetStatus(state.Ready)
		return fmt.Errorf("failed to send merchant-ready event: %w", err)
	}

	respMsg, err := msg.Unmarshal(resp)
	if err != nil {
		u.state.SetStatus(state.Ready)
		return fmt.Errorf("failed to parse message: %w", err)
	}

	if respMsg.Meta.Err != nil {
		u.state.SetStatus(state.Ready)
		return fmt.Errorf("failed to start session: %s", respMsg.Meta.Err.Message)
	}

	now := time.Now()
	u.state.SetStatus(state.InRent)
	u.state.SetSessionID(req.SessionID)
	u.state.SetRequestID(req.RequestID)
	u.state.SetRentStartedAt(&now)
	u.currentSessionID = req.SessionID

	healthCheckCtx, cancel := context.WithCancel(context.Background())

	go func() { u.HealthCheck(healthCheckCtx) }()
	u.sessionHealthCheck.SetCancelFn(cancel)

	statusHTTPCtx, statusHTTPCancel := context.WithCancel(context.Background())
	go func() { u.StatusCheckLoop(statusHTTPCtx) }()
	u.sessionStatusHTTPCheck.SetCancelFn(statusHTTPCancel)

	// MerchantNodeCheckLoop уже запущен при переходе в Ready (rent.Init), не дублируем

	log.Infof("session [%s] started successfully", req.RequestID)

	return nil
}

func (u usecase) configure(ctx context.Context, clientUserID string, settings settingsModel.Settings) error {
	downloadReq := proto.DownloadTemplateReq{
		Template: &proto.TemplateData{
			Id:      settings.Template.ID,
			Type:    settings.Template.Type,
			Version: settings.Template.Version,

			Configuration: &proto.TemplateData_Configuration{
				UseGPU:  settings.Template.UseGPU,
				Volumes: settings.Template.Volumes,
				Envs:    u.enrichEnvs(settings.Template),
			},
		},

		Download: &proto.DownloadTemplateReq_DownloadInfo{
			ImageName: settings.Template.ImageName,
			ImageTag:  settings.Template.ImageTag,
		},
	}

	for _, p := range settings.Template.Ports {
		port := &proto.TemplatePort{
			Port:          strconv.Itoa(p.Port),
			Title:         p.Title,
			AuthAvailable: p.Auth,
		}

		switch settingsModel.PortType(p.Type) {
		case settingsModel.HTTP:
			port.Protocol = proto.PortProtocol_HTTP

		case settingsModel.TCP:
			port.Protocol = proto.PortProtocol_TCP
		}

		downloadReq.Template.Configuration.Ports = append(downloadReq.Template.Configuration.Ports, port)
	}

	log.Infof("template \"%s\" dowload started...", settings.Template.Title)

	if _, err := u.rd.Download(ctx, &downloadReq); err != nil {
		return fmt.Errorf("failed to download template in runtime daemon: %w", err)
	}

	log.Infof("template \"%s\" dowloaded successfully", settings.Template.Title)

	changeModeReq := proto.ChangeModeReq{
		Docker: &proto.DockerConfiguration{
			ClientUserId: clientUserID,

			Container: &proto.ContainerConfiguration{
				TemplateID: settings.Template.ID,
				Options:    &proto.ContainerConfiguration_Options{},
			},

			Auth: &proto.ContainerAuthConfiguration{
				Enabled: true,
				Credentials: &proto.ContainerAuthConfiguration_Credentials{
					Login:    settings.Template.Authentication.Login,
					Password: settings.Template.Authentication.Password,
				},
			},
		},

		Network: &proto.NetworkConfiguration{},
	}

	switch settings.Mode {
	case settingsModel.P2PMode:
		if settings.Network.Tailscale == nil {
			return fmt.Errorf("tailscale settings is empty. Invalid configuration request")
		}

		changeModeReq.Mode = proto.Mode_Host_P2P_VM

		changeModeReq.Network.Tailscale = &proto.TailscaleConfiguration{
			AuthKey: settings.Network.Tailscale.AuthKey,
		}

	case settingsModel.ProxyMode:
		if settings.Network.Piko == nil {
			return fmt.Errorf("piko settings is empty. Invalid configuration request")
		}

		changeModeReq.Mode = proto.Mode_Host_Proxy_VM

		changeModeReq.Network.Piko = &proto.PikoConfiguration{
			AuthKey: settings.Network.Piko.AuthKey,
		}

		for _, e := range settings.Network.Piko.Endpoints {
			endpoint := &proto.PikoConfiguration_Endpoint{
				TemplatePort: strconv.Itoa(e.TemplatePort),
				Name:         e.Endpoint,
			}

			changeModeReq.Network.Piko.Endpoints = append(changeModeReq.Network.Piko.Endpoints, endpoint)
		}
	}

	log.Infof("template \"%s\" configuring...", settings.Template.Title)

	if _, err := u.rd.ChangeMode(ctx, &changeModeReq); err != nil {
		return fmt.Errorf("failed to change mode in runtime daemon: %w", err)
	}

	log.Infof("template \"%s\" configured in mode \"%s\" successfully", settings.Template.Title, settings.Mode)

	return nil
}

func (u usecase) enrichEnvs(settings settingsModel.Template) []string {
	var result []string

	for _, e := range settings.Envs {
		env := fmt.Sprintf("%s=", e.Key)

		switch settingsModel.EnvType(e.Type) {
		case settingsModel.Username:
			env += settings.Authentication.Login

		case settingsModel.Password:
			env += settings.Authentication.Password

		case settingsModel.Other:
			env += e.Key

		default:
			env += "not_defined_value"
		}

		result = append(result, env)
	}

	return result
}
