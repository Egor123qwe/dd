package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"

	rentModel "gitlab.roy9.ru/roy9/backend/statemachine/sessionhandlerservice/internal/model/rent"
	"gitlab.roy9.ru/roy9/backend/statemachine/sessionhandlerservice/internal/service/rent"
	"gitlab.roy9.ru/roy9/backend/statemachine/sessionhandlerservice/internal/service/session"
	"gitlab.roy9.ru/roy9/backend/statemachine/sessionhandlerservice/internal/storage"
)

type Service interface {
	Session() session.Service
	Rent() Rent
	ListTemplates(ctx context.Context) ([]rentModel.Template, error)
	GetTemplate(ctx context.Context, id string) (rentModel.Template, error)
	CreateTemplate(ctx context.Context, title, description, shortDescription, imageName, imageTag string, useGPU bool, ports []rentModel.Port, envs []rentModel.Env, volumes []string) (rentModel.Template, error)
	UpdateTemplate(ctx context.Context, id string, title, description, shortDescription, imageName, imageTag string, useGPU bool, ports []rentModel.Port, envs []rentModel.Env, volumes []string) error
	GetRentClientID(ctx context.Context, requestID string) (string, error)
}

type Rent interface {
	GetRentSettings(ctx context.Context, requestID string) (rentModel.Settings, error)
}

type service struct {
	storage storage.Storage
	session session.Service
	rent    Rent
}

func New(storage storage.Storage) Service {
	rentSvc := rent.New(storage)

	return &service{
		storage: storage,
		session: session.New(storage, rentSvc),
		rent:    rentSvc,
	}
}

func (s *service) Session() session.Service {
	return s.session
}

func (s *service) Rent() Rent {
	return s.rent
}

func (s *service) ListTemplates(ctx context.Context) ([]rentModel.Template, error) {
	return s.storage.Template().ListAll(ctx)
}

func (s *service) GetTemplate(ctx context.Context, id string) (rentModel.Template, error) {
	return s.storage.Template().Get(ctx, id)
}

func generateID() (string, error) {
	var b [8]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(b[:]), nil
}

func (s *service) CreateTemplate(ctx context.Context, title, description, shortDescription, imageName, imageTag string, useGPU bool, ports []rentModel.Port, envs []rentModel.Env, volumes []string) (rentModel.Template, error) {
	id, err := generateID()
	if err != nil {
		return rentModel.Template{}, err
	}
	t := rentModel.Template{
		ID:               id,
		Title:            title,
		Type:             "proxy",
		Description:      description,
		ShortDescription: shortDescription,
		Version:          "1.0",
		ImageName:        imageName,
		ImageTag:         imageTag,
		Ports:            ports,
		Envs:             envs,
		Volumes:          volumes,
		UseGPU:           useGPU,
	}
	if err := s.storage.Template().Create(ctx, t); err != nil {
		return rentModel.Template{}, err
	}
	return t, nil
}

func (s *service) UpdateTemplate(ctx context.Context, id string, title, description, shortDescription, imageName, imageTag string, useGPU bool, ports []rentModel.Port, envs []rentModel.Env, volumes []string) error {
	return s.storage.Template().Update(ctx, id, title, "proxy", description, shortDescription, "1.0", imageName, imageTag, useGPU, ports, envs, volumes)
}

func (s *service) GetRentClientID(ctx context.Context, requestID string) (string, error) {
	r, err := s.storage.Rent().Get(ctx, requestID)
	if err != nil {
		return "", err
	}
	return r.ClientId, nil
}
