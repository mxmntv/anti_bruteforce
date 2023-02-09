package v1

import (
	"context"
	"encoding/json"
	"io"
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
	GetBucketList(ctx context.Context, body *io.ReadCloser) ([]model.Bucket, error)
}

type BucketHandler struct {
	usecase BucketUsecase
	logger  logger.LogInterface
}

type response struct {
	Ok bool `json:"ok"`
}

func NewBucketHandler(u BucketUsecase, l logger.LogInterface) BucketHandler {
	return BucketHandler{
		usecase: u,
		logger:  l,
	}
}

func (h BucketHandler) Register(handler *http.ServeMux) {
	handler.Handle(version+"/check", loggingMiddleware(h.logger, http.HandlerFunc(h.checkBucket)))
	handler.Handle(version+"/remove/blacklist", loggingMiddleware(h.logger, http.HandlerFunc(h.removeFromBlacklist)))
	handler.Handle(version+"/remove/whitelist", loggingMiddleware(h.logger, http.HandlerFunc(h.removeFromWhitelist)))
	handler.Handle(version+"/add/blacklist", loggingMiddleware(h.logger, http.HandlerFunc(h.addToBlacklist)))
	handler.Handle(version+"/add/whitelist", loggingMiddleware(h.logger, http.HandlerFunc(h.addToWhitelist)))
	handler.Handle(version+"/remove/keys", loggingMiddleware(h.logger, http.HandlerFunc(h.removeKeys)))
}

func (h BucketHandler) checkBucket(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var response response
	ctx := r.Context()
	defer r.Body.Close()
	buckets, err := h.usecase.GetBucketList(ctx, &r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	checkListsRes, err := h.usecase.CheckList(ctx, buckets[2].Key) // doub magic number (?)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if checkListsRes.Blacklist {
		response.Ok = false
		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else if checkListsRes.Whitelist {
		response.Ok = true
		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		resp, err := h.usecase.GetBucket(ctx, buckets)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		response.Ok = resp
		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func (h BucketHandler) removeFromBlacklist(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	ctx := r.Context()
	defer r.Body.Close()
	var ip struct {
		Ip string `json:"ip"`
	}
	if err := json.NewEncoder(w).Encode(ip); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err := h.usecase.RemoveFromBlacklist(ctx, ip.Ip)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h BucketHandler) removeFromWhitelist(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	ctx := r.Context()
	defer r.Body.Close()
	var ip struct {
		Ip string `json:"ip"`
	}
	if err := json.NewEncoder(w).Encode(ip); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err := h.usecase.RemoveFromWhitelist(ctx, ip.Ip)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h BucketHandler) addToBlacklist(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	ctx := r.Context()
	defer r.Body.Close()
	var ip struct {
		Ip string `json:"ip"`
	}
	if err := json.NewEncoder(w).Encode(ip); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err := h.usecase.AddToBlacklist(ctx, ip.Ip)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h BucketHandler) addToWhitelist(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	ctx := r.Context()
	defer r.Body.Close()
	var ip struct {
		Ip string `json:"ip"`
	}
	if err := json.NewEncoder(w).Encode(ip); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err := h.usecase.AddToWhitelist(ctx, ip.Ip)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h BucketHandler) removeKeys(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	ctx := r.Context()
	defer r.Body.Close()
	var key struct {
		Key []string `json:"keys"`
	}
	if err := json.NewEncoder(w).Encode(key); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err := h.usecase.Delete(ctx, key.Key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
