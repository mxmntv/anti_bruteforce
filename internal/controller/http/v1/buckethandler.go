package v1

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	internalError "github.com/mxmntv/anti_bruteforce/internal/errors"
	"github.com/mxmntv/anti_bruteforce/internal/model"
	"github.com/mxmntv/anti_bruteforce/pkg/logger"
)

const version = "/v1"

type BucketUsecase interface {
	GetSetBucket(ctx context.Context, bucket []model.Bucket, keys []string) (bool, error)
	Delete(ctx context.Context, keys []string) ([]string, error)
	AddToList(ctx context.Context, ip string, list string) error
	RemoveFromList(ctx context.Context, ip string, list string) (int, error)
	CheckList(ctx context.Context, ip string, list []string) (*model.Included, error)
	GetBucketList(ctx context.Context, req *model.Request) ([]model.Bucket, []string, error)
	GetListnameFromURL(url string) string
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
	md := h.newMiddleware(h.logger)
	handler.Handle("/", md.loggingMD(http.HandlerFunc(h.heartbeat)))
	handler.Handle(version+"/check", md.loggingMD(md.checkMethodMD("POST", http.HandlerFunc(h.checkBucket))))
	handler.Handle(version+"/remove/blacklist", md.loggingMD(md.checkMethodMD("POST", http.HandlerFunc(h.removeFromList))))
	handler.Handle(version+"/remove/whitelist", md.loggingMD(md.checkMethodMD("POST", http.HandlerFunc(h.removeFromList))))
	handler.Handle(version+"/add/blacklist", md.loggingMD(md.checkMethodMD("POST", http.HandlerFunc(h.addToList))))
	handler.Handle(version+"/add/whitelist", md.loggingMD(md.checkMethodMD("POST", http.HandlerFunc(h.addToList))))
	handler.Handle(version+"/remove/keys", md.loggingMD(md.checkMethodMD("POST", http.HandlerFunc(h.removeKeys))))
}

func (h BucketHandler) newMiddleware(l logger.LogInterface) middleware {
	return newMiddleware(l)
}

func (h BucketHandler) heartbeat(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (h BucketHandler) checkBucket(w http.ResponseWriter, r *http.Request) {
	var req model.Request
	defer r.Body.Close()
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	ctx := r.Context()
	buckets, keys, err := h.usecase.GetBucketList(ctx, &req)
	switch {
	case errors.Is(err, internalError.ErrorInvalidReqBody):
		h.logger.Error(err.Error())
		http.Error(w, errors.Unwrap(err).Error(), http.StatusBadRequest)
		return
	case errors.Is(err, internalError.ErrorContextTimeout):
		h.logger.Error(err.Error())
		http.Error(w, errors.Unwrap(err).Error(), http.StatusInternalServerError)
		return
	}

	checkListsRes, err := h.usecase.CheckList(ctx, req.IP, []string{"blacklist", "whitelist"})
	if errors.Is(err, internalError.ErrorInvalidIP) {
		h.logger.Error(err.Error())
		http.Error(w, errors.Unwrap(errors.Unwrap(err)).Error(), http.StatusBadRequest)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	var status checkStatus
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
		resp, err := h.usecase.GetSetBucket(ctx, buckets, keys)
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

func (h BucketHandler) removeFromList(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var ip struct {
		IP string `json:"ip"`
	}
	if err := json.NewDecoder(r.Body).Decode(&ip); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	ctx := r.Context()
	listName := h.usecase.GetListnameFromURL(r.URL.Path)
	res, err := h.usecase.RemoveFromList(ctx, ip.IP, listName)
	switch {
	case errors.Is(err, internalError.ErrorInvalidListName):
		h.logger.Error(err.Error())
		http.Error(w, errors.Unwrap(errors.Unwrap(err)).Error(), http.StatusBadRequest)
		return
	case errors.Is(err, internalError.ErrorInternalDB):
		h.logger.Error(err.Error())
		http.Error(w, errors.Unwrap(err).Error(), http.StatusInternalServerError)
		return
	}
	var item struct {
		DeletedItems int `json:"deleted"`
	}
	item.DeletedItems = res
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(item); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h BucketHandler) addToList(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var ip struct {
		IP string `json:"ip"`
	}
	if err := json.NewDecoder(r.Body).Decode(&ip); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	ctx := r.Context()
	listName := h.usecase.GetListnameFromURL(r.URL.Path)
	err := h.usecase.AddToList(ctx, ip.IP, listName)
	switch {
	case errors.Is(err, internalError.ErrorInvalidIPNet):
		http.Error(w, errors.Unwrap(err).Error(), http.StatusBadRequest)
		return
	case errors.Is(err, internalError.ErrorInvalidListName):
		h.logger.Error(err.Error())
		http.Error(w, errors.Unwrap(errors.Unwrap(err)).Error(), http.StatusBadRequest)
		return
	case errors.Is(err, internalError.ErrorInternalDB):
		h.logger.Error(err.Error())
		http.Error(w, errors.Unwrap(err).Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h BucketHandler) removeKeys(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var key struct {
		Key []string `json:"keys"`
	}
	if err := json.NewDecoder(r.Body).Decode(&key); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	ctx := r.Context()
	res, err := h.usecase.Delete(ctx, key.Key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	var delkey struct {
		DeletedKeys []string `json:"deleted"`
	}
	delkey.DeletedKeys = res
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(delkey); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
