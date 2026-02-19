package template

import (
	"context"
	"errors"
	"fmt"

	"gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/model/docker/template"
	"gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/service/docker/api"
	"gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/service/docker/container"
)

func (s service) Download(ctx context.Context, t template.Template) (<-chan api.PullInfo, error) {
	saved, err := s.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get saved template: %w", err)
	}

	result := make([]template.Template, len(saved))

	copy(result, saved)

	exist := false
	validVersion := true

	for i, template := range saved {
		if template.ID == t.ID {
			exist = true
			// if version not changed - do not update
			validVersion = template.Version == t.Version

			if err := s.api.RemoveContainer(ctx, container.Name(t.ID)); err != nil && !errors.Is(err, api.ErrContainerNotFound) {
				return nil, fmt.Errorf("failed to remove container: %w", err)
			}

			result[i] = t

			break
		}
	}

	if !exist {
		result = append(result, t)
	}

	if !validVersion || !exist {
		imageName := fmt.Sprintf("%s:%s", t.ImageName, t.ImageTag)

		if !validVersion {
			if err := s.api.RemoveImage(ctx, imageName); err != nil {
				return nil, fmt.Errorf("failed to remove image: %w", err)
			}
		}

		reader, err := s.api.PullImage(ctx, imageName)
		if err != nil {
			return nil, fmt.Errorf("failed to download template: %w", err)
		}

		resCh := make(chan api.PullInfo)

		go func() {
			defer close(resCh)

			for info := range reader {
				resCh <- info

				if info.Err != nil {
					log.Errorf("failed to download template: %v", info.Err)
					return
				}
			}

			if err = s.storage.Downloads().Set(result); err != nil {
				log.Errorf("failed to set saved template: %v", err)
				resCh <- api.PullInfo{Err: fmt.Errorf("failed to set saved template: %w", err)}
			}
		}()

		return resCh, nil
	}

	if err = s.storage.Downloads().Set(result); err != nil {
		return nil, fmt.Errorf("failed to set saved template: %w", err)
	}

	return nil, nil
}

func (s service) Download_OLD(ctx context.Context, t template.Template) error {
	saved, err := s.GetAll(ctx)
	if err != nil {
		return fmt.Errorf("failed to get saved template: %w", err)
	}

	result := make([]template.Template, len(saved))

	copy(result, saved)

	exist := false
	validVersion := true

	for i, template := range saved {
		if template.ID == t.ID {
			exist = true
			// if version not changed - do not update
			validVersion = template.Version == t.Version

			if err := s.api.RemoveContainer(ctx, container.Name(t.ID)); err != nil && !errors.Is(err, api.ErrContainerNotFound) {
				return fmt.Errorf("failed to remove container: %w", err)
			}

			result[i] = t

			break
		}
	}

	if !exist {
		result = append(result, t)
	}

	if !validVersion || !exist {
		imageName := fmt.Sprintf("%s:%s", t.ImageName, t.ImageTag)

		if !validVersion {
			if err := s.api.RemoveImage(ctx, imageName); err != nil {
				return fmt.Errorf("failed to remove image: %w", err)
			}
		}

		if err := s.api.PullImageOld(ctx, imageName); err != nil {
			return fmt.Errorf("failed to download template: %w", err)
		}
	}

	if err = s.storage.Downloads().Set(result); err != nil {
		return fmt.Errorf("failed to set saved template: %w", err)
	}

	return nil
}
