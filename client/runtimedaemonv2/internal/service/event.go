package service

import (
	"context"
	"fmt"
	"time"

	"gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/model/runtime"
	"gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/model/runtime/event"
)

const (
	eventCheckInterval = 5 * time.Second
)

func (s service) ExecuteEvent(ctx context.Context, req event.Event) error {
	s.mutex.event.Lock()
	defer s.mutex.event.Unlock()

	switch req {
	case event.StopSharing:
		return s.execStopSharingEvent(ctx)

	case event.ResumeSharing:
		return s.execResumeSharingEvent(ctx)
	}

	return nil
}

func (s service) eventLoop(ctx context.Context) {
	for {
		s.mutex.event.Lock()

		if err := s.checkEvents(context.Background()); err != nil {
			log.Errorf("failed to check container health: %s", err)
		}

		s.mutex.event.Unlock()

		time.Sleep(eventCheckInterval)
	}
}

func (s service) checkEvents(ctx context.Context) error {
	if err := s.checkStopSharingEvent(ctx); err != nil {
		return fmt.Errorf("failed to check stop sharing event: %w", err)
	}

	return nil
}

func (s service) discardEvents(ctx context.Context) error {
	s.mutex.event.Lock()
	defer s.mutex.event.Unlock()

	if err := s.storage.Event().Sharing().Set(nil); err != nil {
		return fmt.Errorf("failed to delete sharing event info from storage: %w", err)
	}

	return nil
}

func (s service) execResumeSharingEvent(ctx context.Context) error {
	info, err := s.storage.Event().Sharing().Get()
	if err != nil {
		return fmt.Errorf("failed to get sharing event info from storage: %w", err)
	}

	if info == nil {
		return fmt.Errorf("failed to resume sharing: nothing to resume")
	}

	if time.Now().After(info.Exp) {
		return fmt.Errorf("failed to resume sharing: stop sharing event already change to disable mode")
	}

	if err := s.changeMode(ctx, info.LastConfiguration); err != nil {
		return fmt.Errorf("failed to resume sharing: %w", err)
	}

	if err := s.storage.Event().Sharing().Set(nil); err != nil {
		return fmt.Errorf("failed to delete sharing event info from storage: %w", err)
	}

	return nil
}

func (s service) execStopSharingEvent(ctx context.Context) error {
	info, err := s.storage.Event().Sharing().Get()
	if err != nil {
		return fmt.Errorf("failed to get sharing event info from storage: %w", err)
	}

	if info != nil {
		return fmt.Errorf("failed to stop sharing: this event is already in progress")
	}

	req, err := s.getCurrentConfiguration()
	if err != nil {
		return fmt.Errorf("failed to get current getCurrentConfiguration: %w", err)
	}

	if req.Settings.Mode != runtime.ModeHostP2PVM && req.Settings.Mode != runtime.ModeHostProxyVM {
		return fmt.Errorf("failed to stop sharing: current mode not supported this event")
	}

	if err := s.storage.Event().Sharing().Set(&event.StopSharingData{
		LastConfiguration: req,
		Exp:               time.Now().Add(s.config.stopSharingTimeout),
	}); err != nil {
		return fmt.Errorf("failed to save sharing event info: %w", err)
	}

	req.Settings.Mode = runtime.ModeLocal // use this mode to stop sharing
	req.Components.Network = nil          // and discard data about network

	return s.changeMode(ctx, req)
}

func (s service) checkStopSharingEvent(ctx context.Context) error {
	event, err := s.storage.Event().Sharing().Get()
	if err != nil {
		return fmt.Errorf("failed to get sharing event info from storage: %w", err)
	}

	if event != nil && time.Now().After(event.Exp) {
		err := s.changeMode(context.Background(), runtime.Configuration{
			Settings: runtime.Settings{Mode: runtime.ModeDisable},
		})

		if err != nil {
			return fmt.Errorf("failed to disable runtime: %w", err)
		}

		if err := s.storage.Event().Sharing().Set(nil); err != nil {
			return fmt.Errorf("failed to delete sharing event info from storage: %w", err)
		}
	}

	return nil
}
