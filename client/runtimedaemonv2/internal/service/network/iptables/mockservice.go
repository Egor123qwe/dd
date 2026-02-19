package iptables

import (
	"context"
	"runtime"
)

func (s mockService) Set(ctx context.Context) error {
	log.Warningf("iptables is not working on %s. Used mock serveice to set", runtime.GOOS)

	return nil
}

func (s mockService) Discard(ctx context.Context) error {
	log.Warningf("iptables is not working on %s. Used mock serveice to discard", runtime.GOOS)

	return nil
}

func (s mockService) IsCorrect(ctx context.Context, mustBeSet bool) (bool, error) {
	log.Warningf("iptables is not working on %s. Used mock serveice to check correctness", runtime.GOOS)

	return true, nil
}
