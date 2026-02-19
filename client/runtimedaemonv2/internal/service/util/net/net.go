package net

import (
	"context"
	"fmt"
	"net"
)

func GetFreePorts(ctx context.Context, count int) ([]string, error) {
	result := make([]string, 0, count)

	// Loop through all possible ports
	for port := 0xFFFF; port >= 1; port-- {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		// Try to listen on the port
		listener, err := net.Listen(
			"tcp", fmt.Sprintf("localhost:%d", port),
		)

		// If we can listen on the port, it's free
		if err == nil {
			listener.Close()

			result = append(result, fmt.Sprintf("%d", port))
		}

		if len(result) == count {
			break
		}
	}

	if len(result) < count {
		return nil, fmt.Errorf("failed to find %d free ports", count)
	}

	return result, nil
}
