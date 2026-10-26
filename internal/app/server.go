package app

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	fsadapter "job-search-cockpit/internal/adapter/fs"
	"job-search-cockpit/internal/adapter/sqlite"
)

func Run(ctx context.Context, cfg Config) error {
	store, err := sqlite.OpenAndMigrate(cfg.DBPath())
	if err != nil {
		return fmt.Errorf("initialize sqlite store: %w", err)
	}
	defer func() {
		_ = store.Close()
	}()

	handler := NewHTTPHandler(store, Dependencies{
		ArtifactBlobStore: fsadapter.NewLocalArtifactBlobStore(cfg.ArtifactsDir()),
	})
	srv := &http.Server{
		Addr:              cfg.Addr,
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		log.Printf("Job Search Cockpit backend listening on %s (db: %s)", cfg.Addr, cfg.DBPath())
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
		close(errCh)
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return srv.Shutdown(shutdownCtx)
	case err := <-errCh:
		return err
	}
}
