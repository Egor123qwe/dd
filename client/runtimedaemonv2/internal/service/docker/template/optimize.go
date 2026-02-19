package template

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/model/docker/template"
)

func (s service) OptimizeMemUsage(ctx context.Context) error {
	// also remove unused templates (templates that have no image)
	templates, err := s.GetAll(ctx)
	if err != nil {
		return err
	}

	if err := s.optimizeImages(ctx, templates); err != nil {
		return fmt.Errorf("failed to optimize host usage volumes: %w", err)
	}

	if err := s.optimizeVolumeStorage(ctx, templates, s.volume.Path().HostVolumeDir); err != nil {
		return fmt.Errorf("failed to optimize host usage volumes: %w", err)
	}

	if err := s.optimizeVolumeStorage(ctx, templates, s.volume.Path().LocalVolumeDir); err != nil {
		return fmt.Errorf("failed to optimize local usage volumes: %w", err)
	}

	return nil
}

func (s service) optimizeImages(ctx context.Context, templates []template.Template) error {
	images, err := s.api.GetImageList(ctx)
	if err != nil {
		return fmt.Errorf("failed to get image list: %w", err)
	}

	for _, img := range images {
		exist := false

		for _, t := range templates {
			if fmt.Sprintf("%s:%s", t.ImageName, t.ImageTag) == img.Name {
				exist = true

				break
			}
		}

		if !exist {
			if err := s.api.RemoveImage(ctx, img.ID); err != nil {
				return fmt.Errorf("failed to remove image: %w", err)
			}
		}
	}

	return nil
}

func (s service) optimizeVolumeStorage(ctx context.Context, templates []template.Template, dir string) error {
	volumes, err := s.GetVolumes(ctx, dir)
	if err != nil {
		return fmt.Errorf("failed to get volumes: %w", err)
	}

	for _, v := range volumes {
		exist := false

		for _, t := range templates {
			if strings.HasPrefix(v.Name, t.ID) {
				exist = true

				break
			}
		}

		if !exist {
			if err := os.RemoveAll(filepath.Join(dir, v.Name)); err != nil {
				return err
			}
		}
	}

	return nil
}
