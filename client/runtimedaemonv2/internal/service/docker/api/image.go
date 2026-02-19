package api

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/docker/docker/api/types/image"
	docker "github.com/docker/docker/client"
)

var (
	downloadDelay = 2 * time.Second
)

type ImageInfo struct {
	ID   string
	Name string
}

type LayerProgress struct {
	// Current is the current status and value of the progress made towards Total.
	Current int64 `json:"current,omitempty"`
	// Total is the end value describing when we made 100% progress for an operation.
	Total int64 `json:"total,omitempty"`
}

type PullError struct {
	Code    int    `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
}

type PullEvent struct {
	LayerProgress LayerProgress `json:"progressDetail,omitempty"`
	Error         *PullError    `json:"errorDetail,omitempty"`
	ID            string        `json:"id,omitempty"`
}

type ProgressInfo struct {
	Total   int64
	Current int64
}

type PullInfo struct {
	FullProgress   ProgressInfo
	LayersProgress map[string]ProgressInfo

	Err error
}

func (s service) PullImage(ctx context.Context, name string) (<-chan PullInfo, error) {
	resCh := make(chan PullInfo)

	var registryAuth string

	if strings.HasPrefix(name, s.config.registry) {
		jsonBytes, _ := json.Marshal(map[string]string{
			"username": s.config.username,
			"password": s.config.password,
		})

		registryAuth = base64.StdEncoding.EncodeToString(jsonBytes)
	}

	pullResp, err := s.dockerApi.ImagePull(
		ctx,
		name,
		image.PullOptions{
			RegistryAuth: registryAuth,
		},
	)

	if err != nil {
		return nil, fmt.Errorf("failed to start pull image: %w", err)
	}

	go func() {
		defer close(resCh)
		defer pullResp.Close()

		downloadStartTime := time.Now()
		decoder := json.NewDecoder(pullResp)

		var event *PullEvent

		progress := PullInfo{
			LayersProgress: make(map[string]ProgressInfo),
		}

		for {
			if err := decoder.Decode(&event); err != nil {
				if err == io.EOF {
					return
				}

				progress.Err = fmt.Errorf("failed to decode pull event: %w", err)
				resCh <- progress

				return
			}

			if event.Error != nil {
				progress.Err = fmt.Errorf("failed to pull image: %s", event.Error.Message)
				resCh <- progress

				return
			}

			progress.LayersProgress[event.ID] = ProgressInfo{
				Current: event.LayerProgress.Current,
				Total:   event.LayerProgress.Total,
			}

			// recalculate progress
			progress.FullProgress.Total = 0
			progress.FullProgress.Current = 0

			for _, layer := range progress.LayersProgress {
				progress.FullProgress.Total += layer.Total
				progress.FullProgress.Current += layer.Current
			}

			if downloadStartTime.Add(downloadDelay).Before(time.Now()) {
				resCh <- progress
			}
		}
	}()

	return resCh, nil
}

func (s service) PullImageOld(ctx context.Context, name string) error {
	var registryAuth string

	if strings.HasPrefix(name, s.config.registry) {
		jsonBytes, _ := json.Marshal(map[string]string{
			"username": s.config.username,
			"password": s.config.password,
		})

		registryAuth = base64.StdEncoding.EncodeToString(jsonBytes)
	}

	pullResp, err := s.dockerApi.ImagePull(
		ctx,
		name,
		image.PullOptions{
			RegistryAuth: registryAuth,
		},
	)

	if err != nil {
		return fmt.Errorf("failed to pull image: %w", err)
	}

	defer pullResp.Close()

	errCh := make(chan error)

	go func() {

		_, err = io.Copy(io.Discard, pullResp)
		errCh <- err
	}()

	select {
	case err := <-errCh:
		if err != nil {
			return fmt.Errorf("failed to pull image: %w", err)
		}

	case <-ctx.Done():
		return fmt.Errorf("pulling was interrupted")
	}

	return nil
}

func (s service) GetImageList(ctx context.Context) ([]ImageInfo, error) {
	var result []ImageInfo

	images, err := s.dockerApi.ImageList(context.Background(), image.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get image list: %w", err)
	}

	for _, i := range images {
		var name string

		for _, tag := range i.RepoTags {
			name += tag
		}

		image := ImageInfo{
			ID:   i.ID,
			Name: name,
		}

		result = append(result, image)
	}

	return result, nil
}

func (s service) GetImageUsage(ctx context.Context, name string) (float64, error) {
	info, _, err := s.dockerApi.ImageInspectWithRaw(ctx, name)
	if err != nil {
		if docker.IsErrNotFound(err) {
			return 0, nil
		}

		return 0, err
	}

	return float64(info.Size) / 1024, nil

}

func (s service) IsExistImage(ctx context.Context, name string) (bool, error) {
	_, _, err := s.dockerApi.ImageInspectWithRaw(ctx, name)
	if err != nil {
		if docker.IsErrNotFound(err) {
			return false, nil
		}

		return false, err
	}

	return true, nil
}

func (s service) RemoveImage(ctx context.Context, name string) error {
	target, _, err := s.dockerApi.ImageInspectWithRaw(ctx, name)
	if err != nil {
		if docker.IsErrNotFound(err) {
			return nil
		}

		return err
	}

	if _, err := s.dockerApi.ImageRemove(ctx, target.ID, image.RemoveOptions{Force: true}); err != nil {
		return fmt.Errorf("falid to remove image %w", err)
	}

	return nil
}
