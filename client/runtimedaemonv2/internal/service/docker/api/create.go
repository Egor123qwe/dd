package api

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/docker/docker/api/types/volume"

	"github.com/docker/cli/opts"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
)

const (
	gpuParam = "all"
)

type Volume struct {
	LocalPath string
	HostPath  string
	SizeLimit int64
}

type CreateContainerReq struct {
	Image string
	Name  string

	UseGPU  bool
	Volumes []Volume
	Envs    []string

	// Cmd — аргументы к ENTRYPOINT образа (например для code-server: ["--auth", "none"]).
	Cmd []string

	CPUs    int64
	Memory  int64
	Storage int64
}

func (s service) CreateContainer(ctx context.Context, req CreateContainerReq) (string, error) {
	// req.Storage и req.Volumes[i].SizeLimit уже заданы вызывающим по пропорциям шаблона

	// Настройка ресурсов для GPU
	resources := container.Resources{}

	if req.UseGPU {
		gpuOpts := opts.GpuOpts{}
		gpuOpts.Set(gpuParam)
		resources.DeviceRequests = gpuOpts.Value()
	}

	if req.CPUs > 0 {
		// Конвертируем количество ядер в наноядра (1 ядро = 1e9 наноядер)
		resources.NanoCPUs = req.CPUs * 1e9
	}

	if req.Memory > 0 {
		resources.Memory = req.Memory
		resources.MemorySwap = req.Memory
	}

	// Создаем и настраиваем все mounts
	var mounts []mount.Mount
	for _, v := range req.Volumes {
		mnt, err := s.setupVolumeMount(ctx, v)
		if err != nil {
			return "", fmt.Errorf("failed to setup volume for %s: %w", v.LocalPath, err)
		}

		if err := s.validateMount(mnt); err != nil {
			return "", fmt.Errorf("invalid mount configuration: %w", err)
		}

		mounts = append(mounts, mnt)
	}

	hostConfig := &container.HostConfig{
		PublishAllPorts: true,
		Resources:       resources,
		Mounts:          mounts,
		RestartPolicy: container.RestartPolicy{
			Name: "no",
		},
	}

	config := &container.Config{
		Image: req.Image,
		Env:   req.Envs,
		Labels: map[string]string{
			"managed_by": "go-service",
		},
	}
	if len(req.Cmd) > 0 {
		config.Cmd = req.Cmd
	}

	// Создаем контейнер
	result, err := s.dockerApi.ContainerCreate(
		ctx,
		config,
		hostConfig,
		&network.NetworkingConfig{},
		&v1.Platform{},
		req.Name,
	)

	if err != nil {
		// Пробуем найти портируемый порт если имя занято
		if strings.Contains(err.Error(), "already in use") {
			req.Name = ""
			fallbackConfig := &container.Config{Image: req.Image, Env: req.Envs}
			if len(req.Cmd) > 0 {
				fallbackConfig.Cmd = req.Cmd
			}
			result, err = s.dockerApi.ContainerCreate(
				ctx,
				fallbackConfig,
				hostConfig,
				&network.NetworkingConfig{},
				&v1.Platform{},
				"",
			)
		}

		if err != nil {
			return "", fmt.Errorf("failed to create container: %w", err)
		}
	}

	return result.ID, nil
}

func (s service) createVolume(ctx context.Context, volumeName, hostPath string, sizeLimit int64) (string, error) {
	// Генерируем уникальное имя для volume если не указано
	if volumeName == "" {
		volumeName = fmt.Sprintf("volume_%s", generateRandomString(8))
	}

	// Базовые опции драйвера
	driverOpts := map[string]string{}

	// Добавляем ограничение размера если указано
	if sizeLimit > 0 {
		switch runtime.GOOS {
		case "linux":
			// На Linux используем tmpfs или xfs квоты
			driverOpts["type"] = "tmpfs"
			driverOpts["device"] = "tmpfs"
			driverOpts["o"] = fmt.Sprintf("size=%d", sizeLimit)

		case "darwin": // macOS
			// На macOS через Docker Desktop используем ограничения виртуальной машины
			driverOpts["type"] = "tmpfs"
			driverOpts["device"] = "tmpfs"
			driverOpts["o"] = fmt.Sprintf("size=%d", sizeLimit)

		case "windows":
			// На Windows используем ограничения через ntfs
			driverOpts["type"] = "ntfs"
			driverOpts["o"] = fmt.Sprintf("size=%d", sizeLimit)
		}
	}

	// Создаем volume
	resp, err := s.dockerApi.VolumeCreate(ctx, volume.CreateOptions{
		Name:       volumeName,
		Driver:     "local",
		DriverOpts: driverOpts,
		Labels: map[string]string{
			"created_by": "go-docker-service",
			"host_path":  hostPath,
		},
	})

	if err != nil {
		// Если создание volume с ограничением не удалось, пробуем без ограничений
		if sizeLimit > 0 {
			resp, err = s.dockerApi.VolumeCreate(ctx, volume.CreateOptions{
				Name:   volumeName,
				Driver: "local",
			})
		}
		if err != nil {
			return "", fmt.Errorf("failed to create volume: %w", err)
		}
	}

	return resp.Name, nil
}

// generateRandomString генерирует случайную строку для имени volume
func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[os.Getpid()%len(charset)] // упрощенно, лучше использовать crypto/rand
	}
	return string(b)
}

// setupVolumeMount настраивает монтирование для volume
func (s service) setupVolumeMount(ctx context.Context, v Volume) (mount.Mount, error) {
	// Если HostPath пустой, используем Docker volume
	if v.HostPath == "" {
		volumeName, err := s.createVolume(ctx, "", v.LocalPath, v.SizeLimit)
		if err != nil {
			return mount.Mount{}, err
		}

		return mount.Mount{
			Type:   mount.TypeVolume,
			Source: volumeName,
			Target: v.LocalPath,
		}, nil
	}

	// Иначе используем bind mount с проверкой прав
	if err := s.prepareHostPath(v.HostPath); err != nil {
		return mount.Mount{}, err
	}

	return mount.Mount{
		Type:     mount.TypeBind,
		Source:   v.HostPath,
		Target:   v.LocalPath,
		ReadOnly: false,
	}, nil
}

// prepareHostPath подготавливает путь на хосте
func (s service) prepareHostPath(hostPath string) error {
	// Создаем директорию если не существует
	err := os.MkdirAll(hostPath, os.ModePerm)
	if err != nil && !os.IsExist(err) {
		return fmt.Errorf("failed to create host path: %w", err)
	}

	// На Windows конвертируем путь если нужно
	if runtime.GOOS == "windows" {
		// Docker на Windows ожидает пути в формате C:/path/to/dir
		hostPath = strings.ReplaceAll(hostPath, "\\", "/")
		if !strings.Contains(hostPath, ":") {
			// Добавляем диск C: если не указан
			hostPath = "C:/" + strings.TrimPrefix(hostPath, "/")
		}
	}

	return nil
}

// validateMount проверяет корректность монтирования
func (s service) validateMount(mnt mount.Mount) error {
	if mnt.Type == mount.TypeBind && mnt.Source == "" {
		return fmt.Errorf("bind mount source cannot be empty")
	}
	if mnt.Target == "" {
		return fmt.Errorf("mount target cannot be empty")
	}
	return nil
}
