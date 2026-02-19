package template

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/model/docker/template"
)

func (s service) Get(ctx context.Context, templateId string) (template.Template, error) {
	savedTemplates, err := s.GetAll(ctx)
	if err != nil {
		return template.Template{}, err
	}

	for _, t := range savedTemplates {
		if t.ID == templateId {
			return t, nil
		}
	}

	return template.Template{}, NotFoundErr
}

func (s service) GetAll(ctx context.Context) ([]template.Template, error) {
	saved, err := s.storage.Downloads().Get()
	if err != nil {
		return nil, fmt.Errorf("failed to get saved template: %w", err)
	}

	haveChanges := false

	var templates []template.Template

	// filter template by exist in container
	for _, t := range saved {
		exist, err := s.api.IsExistImage(ctx, fmt.Sprintf("%s:%s", t.ImageName, t.ImageTag))
		if !exist {
			if err != nil {
				return nil, fmt.Errorf("failed to check if image exist: %w", err)
			}

			haveChanges = true

			continue
		}

		templates = append(templates, t)
	}

	if haveChanges {
		if err = s.storage.Downloads().Set(templates); err != nil {
			return nil, fmt.Errorf("failed to set saved template: %w", err)
		}
	}

	return templates, nil
}

func (s service) GetWithStat(ctx context.Context, templateId string) (template.Info, error) {
	savedTemplates, err := s.GetAllWithStat(ctx)
	if err != nil {
		return template.Info{}, err
	}

	for _, t := range savedTemplates {
		if t.Template.ID == templateId {
			return t, nil
		}
	}

	return template.Info{}, NotFoundErr
}

func (s service) GetAllWithStat(ctx context.Context) ([]template.Info, error) {
	templates, err := s.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get templates")
	}

	hostUsageVolumes, err := s.GetVolumes(ctx, s.volume.Path().HostVolumeDir)
	if err != nil {
		return nil, fmt.Errorf("failed to get host usage volumes: %w", err)
	}

	localUsageVolumes, err := s.GetVolumes(ctx, s.volume.Path().LocalVolumeDir)
	if err != nil {
		return nil, fmt.Errorf("failed to get local usage volumes: %w", err)
	}

	var result []template.Info

	for _, t := range templates {
		info := template.Info{
			Template: t,
		}

		for _, v := range hostUsageVolumes {
			if strings.HasPrefix(v.Name, t.ID) {
				info.RentMemUsage += v.Mem
			}
		}

		for _, v := range localUsageVolumes {
			if strings.HasPrefix(v.Name, t.ID) {
				info.LocalMemUsage += v.Mem
			}
		}

		info.ImageUsage, err = s.api.GetImageUsage(ctx, fmt.Sprintf("%s:%s", t.ImageName, t.ImageTag))
		if err != nil {
			return nil, fmt.Errorf("failed to get image usage: %w", err)
		}

		result = append(result, info)
	}

	return result, nil
}
func (s service) GetFiltredWithStat(ctx context.Context, templateTypes []string) ([]template.Info, error) {
	templates, err := s.GetAllWithStat(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get templates")
	}

	var result []template.Info

	for _, t := range templates {
		for _, tType := range templateTypes {
			if t.Template.Type == tType {
				result = append(result, t)
			}
		}
	}

	return result, nil
}

func (s service) GetVolumes(ctx context.Context, dir string) ([]template.Volume, error) {
	var result []template.Volume

	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		if file.IsDir() {
			size, err := getDirSize(ctx, filepath.Join(dir, file.Name()))
			if err != nil {
				return nil, err
			}

			v := template.Volume{
				Name: file.Name(),
				Mem:  float64(size) / 1024,
			}

			result = append(result, v)
		}
	}

	return result, nil
}

func getDirSize(ctx context.Context, dir string) (int64, error) {
	var size int64

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			size += info.Size()
		}

		return nil
	})

	return size, err
}
