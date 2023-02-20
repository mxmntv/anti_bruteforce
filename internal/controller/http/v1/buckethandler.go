package v1

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/mxmntv/anti_bruteforce/internal/model"
	"github.com/mxmntv/anti_bruteforce/internal/usecase/repository"
	"github.com/mxmntv/anti_bruteforce/pkg/logger"
	"gopkg.in/validator.v2"
)

const version = "/v1"

type BucketUsecase interface {
	GetBucket(ctx context.Context, bucket map[string]model.Bucket) (bool, error)
	Delete(ctx context.Context, keys []string) ([]string, error)
	AddToBlacklist(ctx context.Context, ip string) error
	RemoveFromBlacklist(ctx context.Context, ip string) (int, error)
	AddToWhitelist(ctx context.Context, ip string) error
	RemoveFromWhitelist(ctx context.Context, ip string) (int, error)
	CheckList(ctx context.Context, ip string) (*model.Included, error)
	GetBucketList(ctx context.Context, req *model.Request) (map[string]model.Bucket, error)
}

type BucketHandler struct {
	usecase BucketUsecase
	logger  logger.LogInterface
}

type checkStatus struct {
	Ok bool `json:"ok"`
}

func NewBucketHandler(u BucketUsecase, l logger.LogInterface) BucketHandler {
	return BucketHandler{
		usecase: u,
		logger:  l,
	}
}

func (h BucketHandler) Register(handler *http.ServeMux) {
	handler.Handle("/", loggingMiddleware(h.logger, http.HandlerFunc(h.heartbeat)))
	handler.Handle(version+"/check", loggingMiddleware(h.logger, http.HandlerFunc(h.checkBucket)))
	handler.Handle(version+"/remove/blacklist", loggingMiddleware(h.logger, http.HandlerFunc(h.removeFromBlacklist)))
	handler.Handle(version+"/remove/whitelist", loggingMiddleware(h.logger, http.HandlerFunc(h.removeFromWhitelist)))
	handler.Handle(version+"/add/blacklist", loggingMiddleware(h.logger, http.HandlerFunc(h.addToBlacklist)))
	handler.Handle(version+"/add/whitelist", loggingMiddleware(h.logger, http.HandlerFunc(h.addToWhitelist)))
	handler.Handle(version+"/remove/keys", loggingMiddleware(h.logger, http.HandlerFunc(h.removeKeys)))
}

func (h BucketHandler) heartbeat(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (h BucketHandler) checkBucket(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req model.Request
	var status checkStatus
	ctx := r.Context()
	defer r.Body.Close()
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := validator.Validate(req); err != nil {
		e := errors.New("validate request body err:" + err.Error())
		http.Error(w, e.Error(), http.StatusBadRequest)
		return
	}

	buckets, err := h.usecase.GetBucketList(ctx, &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	checkListsRes, err := h.usecase.CheckList(ctx, buckets["ip"].Key)
	var badIP *repository.BadIPError
	if errors.As(err, &badIP) {
		http.Error(w, badIP.Error(), http.StatusBadRequest)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	switch {
	case checkListsRes.Blacklist:
		status.Ok = false
		if err := json.NewEncoder(w).Encode(status); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	case checkListsRes.Whitelist:
		status.Ok = true
		if err := json.NewEncoder(w).Encode(status); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	default:
		resp, err := h.usecase.GetBucket(ctx, buckets)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		status.Ok = resp
		if err := json.NewEncoder(w).Encode(status); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func (h BucketHandler) removeFromBlacklist(w http.ResponseWriter, r *http.Request) { //nolint:dupl
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	ctx := r.Context()
	defer r.Body.Close()
	var ip struct {
		IP string `json:"ip"`
	}
	if err := json.NewDecoder(r.Body).Decode(&ip); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	var item struct {
		DeletedItems int `json:"deleted"`
	}
	res, err := h.usecase.RemoveFromBlacklist(ctx, ip.IP)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	item.DeletedItems = res
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(item); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h BucketHandler) removeFromWhitelist(w http.ResponseWriter, r *http.Request) { //nolint:dupl
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	ctx := r.Context()
	defer r.Body.Close()
	var ip struct {
		IP string `json:"ip"`
	}
	if err := json.NewDecoder(r.Body).Decode(&ip); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	var item struct {
		DeletedItems int `json:"deleted"`
	}
	res, err := h.usecase.RemoveFromWhitelist(ctx, ip.IP)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	item.DeletedItems = res
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(item); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h BucketHandler) addToBlacklist(w http.ResponseWriter, r *http.Request) { //nolint:dupl
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	ctx := r.Context()
	defer r.Body.Close()
	var ip struct {
		IP string `json:"ip"`
	}
	if err := json.NewDecoder(r.Body).Decode(&ip); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	err := h.usecase.AddToBlacklist(ctx, ip.IP)
	var badIPNet *repository.BadIPNetError
	if errors.As(err, &badIPNet) {
		http.Error(w, badIPNet.Error(), http.StatusBadRequest)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h BucketHandler) addToWhitelist(w http.ResponseWriter, r *http.Request) { //nolint:dupl
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	ctx := r.Context()
	defer r.Body.Close()
	var ip struct {
		IP string `json:"ip"`
	}
	if err := json.NewDecoder(r.Body).Decode(&ip); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	err := h.usecase.AddToWhitelist(ctx, ip.IP)
	var badIPNet *repository.BadIPNetError
	if errors.As(err, &badIPNet) {
		http.Error(w, badIPNet.Error(), http.StatusBadRequest)
		return
	}
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
	if err := json.NewDecoder(r.Body).Decode(&key); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	var delkey struct {
		DeletedKeys []string `json:"deleted"`
	}
	res, err := h.usecase.Delete(ctx, key.Key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	delkey.DeletedKeys = res
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(delkey); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
