package rent

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"gitlab.roy9.ru/roy9/backend/clientside/merchantclient/internal/model/event"
	"gitlab.roy9.ru/roy9/backend/clientside/merchantclient/internal/model/msg"
	hardwareModel "gitlab.roy9.ru/roy9/backend/clientside/merchantclient/internal/model/msg/content/hardware"
	"gitlab.roy9.ru/roy9/backend/clientside/merchantclient/internal/model/msg/content/shareP2P"
	"gitlab.roy9.ru/roy9/backend/clientside/merchantclient/internal/service/usecase/state"
	rdModel "gitlab.roy9.ru/roy9/backend/clientside/merchantclient/pkg/proto/runtimedaemon/generate"
)

const (
	schedule = "infinity"
)

func (u usecase) Init(ctx context.Context, cheatMode bool, nodeName string, memoryLimitBytes, storageLimitBytes, cpuLimit int64) error {
	u.state.Mutex().Lock()
	defer u.state.Mutex().Unlock()

	hardware, err := u.rd.GetHardware(ctx)
	if err != nil {
		return fmt.Errorf("failed to get hardware info from daemon: %w", err)
	}

	status := u.state.GetStatus()

	if status != state.Disabled {
		return fmt.Errorf("merchant have wrong status to start configuring: status: %s", status)
	}

	u.state.SetStatus(state.Configuring)
	u.state.SetMemoryLimitBytes(memoryLimitBytes)
	u.state.SetStorageLimitBytes(storageLimitBytes)
	u.state.SetCPULimit(cpuLimit)

	initResp, err := u.init(ctx, hardware, cheatMode, nodeName, memoryLimitBytes, storageLimitBytes, cpuLimit)
	if err != nil {
		u.state.SetStatus(state.Disabled)

		return fmt.Errorf("failed to init merchant: %w", err)
	}

	if err := u.ready(ctx, initResp.SessionID); err != nil {
		u.state.SetStatus(state.Disabled)

		return fmt.Errorf("failed to configurate merchant: %w", err)
	}

	u.state.SetStatus(state.Ready)
	u.state.SetSessionID(initResp.SessionID)
	u.state.SetTotalPrice(float64(initResp.Price))

	keepAliveCtx, cancel := context.WithCancel(context.Background())
	go func() { u.KeepAlive(keepAliveCtx) }()
	u.keepAlive.SetCancelFn(cancel)

	if u.nodeCheck != nil {
		u.nodeCheck.StartMerchantNodeCheckLoop(context.Background())
	}

	return nil
}

func (u usecase) init(ctx context.Context, hardware *rdModel.SystemInfo, cheatMode bool, nodeName string, memoryLimitBytes, storageLimitBytes, cpuLimit int64) (shareP2P.InitMerchantResp, error) {
	reqData := shareP2P.InitMerchantReq{
		Schedule: shareP2P.Schedule{Type: schedule},
		Prepull:  []hardwareModel.Prepull{},
		NodeName: nodeName,
	}

	var err error
	reqData.Hardware, err = u.newConfiguration(hardware, cheatMode, memoryLimitBytes, storageLimitBytes, cpuLimit)
	if err != nil {
		return shareP2P.InitMerchantResp{}, fmt.Errorf("failed to create configuration: %w", err)
	}

	var respData shareP2P.InitMerchantResp

	req := msg.Marshal(string(event.ShareP2PInit), msg.Meta{MessageID: uuid.New().String()}, reqData)

	resp, err := u.wsConn.MsgConn.Do(ctx, u.wsConn.Writer().Write, req)
	if err != nil {
		return shareP2P.InitMerchantResp{}, fmt.Errorf("failed to send merchant init request: %w", err)
	}

	msg, err := msg.Unmarshal(resp)
	if err != nil {
		return shareP2P.InitMerchantResp{}, fmt.Errorf("failed to parse message: %w", err)
	}

	if msg.Meta.Err != nil {
		msgText := msg.Meta.Err.Message
		if strings.Contains(strings.ToLower(msgText), "unidentified") {
			return shareP2P.InitMerchantResp{}, fmt.Errorf("не все компоненты вашего устройства корректно определены; обратитесь в поддержку для добавления оборудования в прайс")
		}
		return shareP2P.InitMerchantResp{}, fmt.Errorf("failed to init merchant: %s", msgText)
	}

	if err := msg.UnmarshalContent(&respData); err != nil {
		return shareP2P.InitMerchantResp{}, fmt.Errorf("failed to parse message content: %w", err)
	}

	return respData, nil
}

func (u usecase) ready(ctx context.Context, sessionID string) error {
	data := shareP2P.ReadyMerchantReq{
		SessionID: sessionID,
	}

	req := msg.Marshal(string(event.ShareP2PReady), msg.Meta{MessageID: uuid.New().String()}, data)

	resp, err := u.wsConn.MsgConn.Do(ctx, u.wsConn.Writer().Write, req)
	if err != nil {
		return fmt.Errorf("failed to send merchant init request: %w", err)
	}

	msg, err := msg.Unmarshal(resp)
	if err != nil {
		return fmt.Errorf("failed to parse message: %w", err)
	}

	if msg.Meta.Err != nil {
		return fmt.Errorf("failed to init merchant: %s", msg.Meta.Err.Message)
	}

	return nil
}
