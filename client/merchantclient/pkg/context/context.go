package context

import (
	"os"
	"os/signal"

	"golang.org/x/net/context"
)

func WithSignal(ctx context.Context, sig ...os.Signal) (context.Context, context.CancelFunc) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, sig...)

	ctx, cancel := context.WithCancel(ctx)

	go func() {
		select {
		case <-ctx.Done():
			return

		case <-quit:
			cancel()
		}
	}()

	return ctx, cancel
}
