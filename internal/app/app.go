package app

import (
	"context"
	"fmt"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/mxmntv/anti_bruteforce/config"
	handler "github.com/mxmntv/anti_bruteforce/internal/controller/http/v1"
	internalError "github.com/mxmntv/anti_bruteforce/internal/errors"
	"github.com/mxmntv/anti_bruteforce/internal/usecase"
	"github.com/mxmntv/anti_bruteforce/internal/usecase/repository"
	"github.com/mxmntv/anti_bruteforce/pkg/httpserver"
	"github.com/mxmntv/anti_bruteforce/pkg/logger"
	rs "github.com/mxmntv/anti_bruteforce/pkg/redis"
)

func Run(cfg *config.Config) error {
	logger := logger.New(cfg.Log.Level) // tdebt
	rs, err := rs.NewRedisClent(fmt.Sprintf("%s:%d", cfg.Redis.RsHost, cfg.Redis.RsPort))
	if err != nil {
		return fmt.Errorf("[run] error: %w", internalError.ErrorRedisConnect)
	}
	bucketRepository := repository.NewBucketRepo(rs.Client)
	bucketUsecase := usecase.NewBucketUsecase(bucketRepository, cfg.BucketCapacity)

	mux := http.NewServeMux()
	handle := handler.NewBucketHandler(bucketUsecase, logger)
	handle.Register(mux)

	server := httpserver.NewServer(cfg.HTTP.Port, mux)

	ctx, cancel := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer cancel()

	go func() {
		<-ctx.Done()
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
		defer cancel()

		if err := server.Stop(ctx); err != nil {
			logger.Error("[run] failed to stop http server error: %w", err)
		}
		if err := rs.Close(); err != nil {
			logger.Error("[run] redis close error: %w", err)
		}
	}()

	logger.Info("[app] antibruteforce is running...")

	if err := server.Start(ctx); err != nil {
		cancel()
		logger.Error("[run] failed to start http server error:", err.Error())
	}
	// cmd.Execute()
	return nil
}
