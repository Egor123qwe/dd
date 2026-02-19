package template

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/service/docker/container"
)

var (
	templateInUsageErr = errors.New("template in usage")

	currentTemplateRemoveErr       = errors.New("you can't remove current template")
	currentTemplateVolumeRemoveErr = errors.New("you can't remove current template volume")
)

type RemoveVolumeOptions struct {
	HostUsageVolumes  bool
	LocalUsageVolumes bool
}

func (s service) Remove(ctx context.Context, templateID string) error {
	if err := s.templateValidation(templateID); err != nil {
		if errors.Is(err, templateInUsageErr) {
			return currentTemplateRemoveErr
		}

		return fmt.Errorf("falied to validate template: %w", err)
	}

	templates, err := s.GetAll(ctx)
	if err != nil {
		return fmt.Errorf("falied to get templates")
	}

	for _, t := range templates {
		if t.ID == templateID {
			if err := s.api.RemoveContainer(ctx, container.Name(t.ID)); err != nil {
				return fmt.Errorf("failed to remove container: %w", err)
			}

			if err := s.api.RemoveImage(ctx, fmt.Sprintf("%s:%s", t.ImageName, t.ImageTag)); err != nil {
				return fmt.Errorf("failed to remove image: %w", err)
			}
		}
	}

	if err := s.OptimizeMemUsage(ctx); err != nil {
		return fmt.Errorf("failed to optimize mem usage: %w", err)
	}

	return nil
}

func (s service) RemoveVolumes(ctx context.Context, templateID string, opts RemoveVolumeOptions) error {
	if err := s.templateValidation(templateID); err != nil {
		if errors.Is(err, templateInUsageErr) {
			return currentTemplateVolumeRemoveErr
		}

		return fmt.Errorf("falied to validate template: %w", err)
	}

	if opts.HostUsageVolumes {
		if err := s.removeVolumes(ctx, templateID, s.volume.Path().HostVolumeDir); err != nil {
			return fmt.Errorf("falid to remove host usage volumes: %w", err)
		}
	}

	if opts.LocalUsageVolumes {
		if err := s.removeVolumes(ctx, templateID, s.volume.Path().LocalVolumeDir); err != nil {
			return fmt.Errorf("falid to remove local usage volumes: %w", err)
		}
	}

	return nil
}

func (s service) removeVolumes(ctx context.Context, templateID string, dir string) error {
	volumes, err := s.GetVolumes(ctx, dir)
	if err != nil {
		return fmt.Errorf("failed to get volumes: %w", err)
	}

	for _, v := range volumes {
		if strings.HasPrefix(v.Name, templateID) {
			if err := os.RemoveAll(filepath.Join(dir, v.Name)); err != nil {
				return err
			}
		}
	}

	return nil
}

func (s service) templateValidation(templateID string) error {
	settings, err := s.storage.Settings().Get()
	if err != nil {
		return fmt.Errorf("failed to get settings: %w", err)
	}

	if settings.Options != nil && settings.Options.Container.TemplateID == templateID {
		return templateInUsageErr
	}

	return nil
}
