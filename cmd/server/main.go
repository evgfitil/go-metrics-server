package main

import (
	"github.com/evgfitil/go-metrics-server.git/internal/handlers"
	"github.com/evgfitil/go-metrics-server.git/internal/logger"
	"github.com/evgfitil/go-metrics-server.git/internal/storage"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func MetricsRouter(s *storage.MemStorage) chi.Router {
	r := chi.NewRouter()
	r.Use(middleware.Compress(5))
	r.Get("/", handlers.GetAllMetrics(s))
	r.Route("/value", func(r chi.Router) {
		r.Post("/", handlers.GetMetricsJSON(s))
		r.Get("/{type}/{name}", handlers.GetMetricsPlain(s))
	})
	r.Route("/update", func(r chi.Router) {
		r.Post("/", handlers.UpdateMetricsJSON(s))
		r.Post("/{type}/{name}/{value}", handlers.UpdateMetricsPlain(s))
	})
	return r
}

func main() {
	logger.InitLogger()
	defer logger.Sugar.Sync()

	config := NewConfig()
	err := config.ParseFlags()
	if err != nil {
		logger.Sugar.Fatalf("error getting arguments: %v", err)
	}
	var saveSignal chan struct{}
	if config.storeInterval == 0 {
		saveSignal = make(chan struct{})
	}
	s := storage.NewMemStorage(saveSignal)

	var fileStorage *storage.FileStorage
	if config.fileStoragePath != "" {
		fileStorage, err = storage.NewFileStorage(
			config.fileStoragePath, s, time.Duration(config.storeInterval)*time.Second, saveSignal)
		if err != nil {
			logger.Sugar.Fatalf("error initializing file storage: %v", err)
		}
		defer fileStorage.Close()
		if config.restore {
			if err := fileStorage.LoadMetrics(); err != nil {
				logger.Sugar.Errorf("error loading metrics: %v", err)
			}
		}
	}
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		logger.Sugar.Infoln("starting server")
		err = http.ListenAndServe(config.bindAddress, logger.WithLogging(MetricsRouter(s)))
		if err != nil {
			logger.Sugar.Fatalf("error starting server: %v", err)
		}
	}()

	<-quit
	logger.Sugar.Info("shutting down server")

	if fileStorage != nil {
		if err := fileStorage.Close(); err != nil {
			logger.Sugar.Errorf("error closing file storage: %v", err)
		}
	}
}
