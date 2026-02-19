package http

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"gitlab.roy9.ru/roy9/backend/statemachine/connectioncoordinatorservice/internal/server/launcher"
)

type server struct {
	srv    *http.Server
	config Config
}

func New(router http.Handler, config Config) launcher.Server {
	return &server{
		srv: &http.Server{
			Addr:        fmt.Sprintf("%s:%d", config.Host, config.Port),
			Handler:     router,
			ReadTimeout: config.ReadTime,
		},

		config: config,
	}
}

func (s *server) Serve(ctx context.Context) error {
	errCh := make(chan error)

	go func() {
		errCh <- s.srv.ListenAndServe()
	}()

	log.Printf("server: serving on http://%s\n", s.srv.Addr)

	select {
	case err := <-errCh:
		return fmt.Errorf("server: %w", err)

	case <-ctx.Done():
		ctx, cancel := context.WithTimeout(context.Background(), s.config.ShutdownTime)
		defer cancel()

		if err := s.srv.Shutdown(ctx); err != nil {
			log.Println("server: Shutdown error: " + err.Error())
		}

		log.Println("server: server stopped successfully.")
	}

	return nil
}
