package v1

import (
	"context"
	"net/http"

	"github.com/mxmntv/anti_bruteforce/internal/model"
	"github.com/mxmntv/anti_bruteforce/pkg/logger"
)

const version = "v1"

type BucketUsecase interface {
	GetBucket(ctx context.Context, bucket []model.Bucket) (bool, error)
	Delete(ctx context.Context, keys []string) error
	AddToBlacklist(ctx context.Context, ip string) error
	RemoveFromBlacklist(ctx context.Context, ip string) error
	AddToWhitelist(ctx context.Context, ip string) error
	RemoveFromWhitelist(ctx context.Context, ip string) error
	CheckList(ctx context.Context, ip string) (*model.Included, error)
}

type BucketHandler struct {
	usecase BucketUsecase
	logger  logger.LogInterface
}

func NewEventHandler(u BucketUsecase, l logger.LogInterface) BucketHandler {
	return BucketHandler{
		usecase: u,
		logger:  l,
	}
}

func (h BucketHandler) Register(handler *http.ServeMux) {
	handler.Handle(version+"/", loggingMiddleware(h.logger, http.HandlerFunc(h.heartbeat)))
}

func (h BucketHandler) heartbeat(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("hello"))
}
