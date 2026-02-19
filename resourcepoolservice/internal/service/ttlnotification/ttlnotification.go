package ttlnotification

import (
	"context"
)

type Service struct {
	merchantRepository MerchantRepository
}

type MerchantRepository interface {
	Delete(ctx context.Context, sessionID, deletionReason string) (string, string, error)
}

func New(merchantRepository MerchantRepository) *Service {
	service := &Service{
		merchantRepository: merchantRepository,
	}

	return service
}

func (s *Service) Delete(ctx context.Context, sessionID string) (string, string, error) {
	return s.merchantRepository.Delete(ctx, sessionID, "ttl expired")
}
