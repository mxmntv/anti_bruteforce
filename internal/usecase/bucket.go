package usecase

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	internalError "github.com/mxmntv/anti_bruteforce/internal/errors"
	"github.com/mxmntv/anti_bruteforce/internal/model"
)

const defaultTTL = time.Minute

type BucketRepository interface {
	GetBucket(ctx context.Context, keys []string) ([]model.BucketInDB, error)
	DeleteKeys(ctx context.Context, keys []string) ([]string, error)
	SetBucket(ctx context.Context, bucket model.Bucket) error
	DecrementValue(ctx context.Context, key string) error
	RemoveFromList(ctx context.Context, ip string, list string) (int, error)
	AddToList(ctx context.Context, ip string, list string) error
	CheckList(ctx context.Context, ip string, list string) (bool, error)
}

type BucketUsecase struct {
	repository BucketRepository
	capacity   model.BucketCapacity
}

func NewBucketUsecase(r BucketRepository, c model.BucketCapacity) *BucketUsecase {
	return &BucketUsecase{r, c}
}

func (u BucketUsecase) GetSetBucket(ctx context.Context, bucket []model.Bucket, keys []string) (bool, error) {
	dbBucketRes, err := u.repository.GetBucket(ctx, keys)
	if err != nil {
		return false, fmt.Errorf("[uc - getBucket] error: %w", err)
	}

	var c int

	for i, res := range dbBucketRes {
		bucketError := res.Error
		value := res.Value
		switch {
		case errors.Is(bucketError, internalError.ErrorDBNotFound):
			setCandidate := model.Bucket{Key: res.Key, Capacity: bucket[i].Capacity, TTL: defaultTTL}
			if err := u.repository.SetBucket(ctx, setCandidate); err != nil {
				return false, fmt.Errorf("[uc - getBucket] error: %w", err)
			}
		case errors.Is(bucketError, internalError.ErrorInternalDB):
			return false, fmt.Errorf("[uc - getBucket] error: %w", bucketError)
		case value == 0:
			c++
		case value > 0:
			if err := u.repository.DecrementValue(ctx, res.Key); err != nil {
				return false, fmt.Errorf("[uc - getBucket] error: %w", err)
			}
		}
	}

	if c > 0 {
		return false, nil
	}
	return true, nil
}

func (u BucketUsecase) Delete(ctx context.Context, keys []string) ([]string, error) {
	return u.repository.DeleteKeys(ctx, keys)
}

func (u BucketUsecase) AddToList(ctx context.Context, ip string, list string) error {
	return u.repository.AddToList(ctx, ip, list)
}

func (u BucketUsecase) RemoveFromList(ctx context.Context, ip string, list string) (int, error) {
	return u.repository.RemoveFromList(ctx, ip, list)
}

func (u BucketUsecase) CheckList(ctx context.Context, ip string, list []string) (*model.Included, error) {
	var wg sync.WaitGroup
	included := &model.Included{}
	ers := make(chan error, 2)
	defer close(ers)
	select {
	case <-ctx.Done():
	default:
		for _, el := range list {
			wg.Add(1)
			go func(el string) {
				res, err := u.repository.CheckList(ctx, ip, el)
				if err != nil {
					ers <- err
				}
				if el == "blacklist" {
					included.Blacklist = res
				} else if el == "whitelist" {
					included.Whitelist = res
				}
				wg.Done()
			}(el)
		}
		wg.Wait()
	}

	select {
	case e := <-ers:
		return nil, fmt.Errorf("[uc - checkList] error: %w", e)
	default:
		break
	}
	return included, nil
}

func (u BucketUsecase) GetBucketList(ctx context.Context, req *model.Request) ([]model.Bucket, []string, error) {
	select {
	case <-ctx.Done():
		return nil, nil, fmt.Errorf("[getBucketList] error: %w", internalError.ErrorContextTimeout)
	default:
		if req.Login == "" || req.Password == "" || req.IP == "" {
			return nil, nil, fmt.Errorf("[getBucketList] error: %w", internalError.ErrorInvalidReqBody)
		}
		buckets := []model.Bucket{
			{
				Key:      req.Login,
				Capacity: u.capacity.Login,
				TTL:      defaultTTL,
			},
			{
				Key:      req.Password,
				Capacity: u.capacity.Password,
				TTL:      defaultTTL,
			},
			{
				Key:      req.IP,
				Capacity: u.capacity.IP,
				TTL:      defaultTTL,
			},
		}
		keys := []string{req.Login, req.Password, req.IP}
		return buckets, keys, nil
	}
}

func (u BucketUsecase) GetListnameFromURL(url string) string {
	eps := strings.Split(url, "/")
	return eps[len(eps)-1]
}
