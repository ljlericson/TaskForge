package api

import (
	"context"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/ljlericson/TaskForge/internal/logging"
)

func Server(ctx context.Context, addr string, logger *logging.Logger, handler *Handler) error {
	logger.Infoln("starting server")

	r := chi.NewRouter()

	// r.Use(logger.RequestLogger())

	r.Use(cors.Handler(cors.Options{

		AllowedOrigins: []string{"*"},

		AllowedMethods: []string{
			"GET",
			"POST",
			"PUT",
			"DELETE",
			"OPTIONS",
		},

		AllowedHeaders: []string{
			"Accept",
			"Authorization",
			"Content-Type",
			"X-CSRF-Token",
		},

		AllowCredentials: false,
		MaxAge:           300,
	}))

	ConfigureRoutes(handler, r)

	srv := &http.Server{
		Addr:    addr,
		Handler: r,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil &&
			err != http.ErrServerClosed {

			logger.Abortln(err.Error())
		}
	}()

	<-ctx.Done()
	logger.Infoln("shutting down server")
	shutdownCtx, cancel := context.WithTimeout(
		context.Background(),
		5*time.Second,
	)
	defer cancel()

	return srv.Shutdown(shutdownCtx)
}
