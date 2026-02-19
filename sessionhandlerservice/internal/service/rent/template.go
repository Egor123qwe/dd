package rent

import (
	"context"

	"github.com/dchest/uniuri"
	"gitlab.roy9.ru/roy9/backend/statemachine/sessionhandlerservice/internal/model/rent"
)

func (s service) template(ctx context.Context, templateID string) (rent.TemplateSettings, error) {
	template, err := s.storage.Template().Get(ctx, templateID)
	if err != nil {
		return rent.TemplateSettings{}, err
	}

	result := rent.TemplateSettings{
		Template:       template,
		Authentication: s.generateCredentials(ctx),
	}

	return result, nil
}

func (s service) generateCredentials(ctx context.Context) rent.Authentication {
	return rent.Authentication{
		Login:    uniuri.NewLen(6),
		Password: uniuri.NewLen(12),
	}
}
