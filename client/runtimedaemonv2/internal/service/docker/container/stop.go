package container

import "context"

func (s service) Stop(ctx context.Context, name string) error {
	return s.api.StopContainer(ctx, name)
}
