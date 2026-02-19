package template

import (
	"context"
	"fmt"

	"gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/model/docker/container"
)

func (s service) AddUsage(ctx context.Context, templateID string, usage container.Usage) error {
	templates, err := s.GetAll(ctx)
	if err != nil {
		return fmt.Errorf("falied to get templates")
	}

	for i := range templates {
		if templates[i].ID == templateID {
			if templates[i].Usages == nil {
				templates[i].Usages = make(map[int32]struct{})
			}

			templates[i].Usages[int32(usage)] = struct{}{}
		}
	}

	if err := s.storage.Downloads().Set(templates); err != nil {
		return fmt.Errorf("failed to update templates: %w", err)
	}

	return nil
}
