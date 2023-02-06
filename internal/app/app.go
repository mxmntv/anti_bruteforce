package app

import (
	"context"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/mxmntv/anti_bruteforce/config/config"
	"github.com/mxmntv/anti_bruteforce/internal/usecase/repository"
	"github.com/mxmntv/anti_bruteforce/pkg/httpserver"
	"github.com/mxmntv/anti_bruteforce/pkg/logger"
)

func Run(cfg *config.Config) error {
	logger := logger.New(cfg.Log.Level)

	repo := repository.BucketRepository

	eventUseCase := usecase.NewEventUsecase(repo)

	mux := http.NewServeMux()
	handle := handler.NewEventHandler(eventUseCase, logger)
	handle.Register(mux)

	server := httpserver.NewServer(cfg.HTTP.Host, cfg.HTTP.Port, mux)

	ctx, cancel := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer cancel()

	go func() {
		<-ctx.Done()
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
		defer cancel()

		if err := server.Stop(ctx); err != nil {
			logger.Error("failed to stop http server: " + err.Error())
		}
	}()

	logger.Info("calendar is running...")

	if err := server.Start(ctx); err != nil {
		logger.Error("failed to start http server: " + err.Error())
		cancel()
		return err
	}
	return nil
}
