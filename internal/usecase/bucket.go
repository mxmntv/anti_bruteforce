package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/mxmntv/anti_bruteforce/internal/model"
)

type BucketRepository interface {
	GetSetBucket(ctx context.Context, bucket []model.Bucket) (bool, error)
	DeleteKeys(ctx context.Context, keys []string) error
	AddToBlacklist(ctx context.Context, ip string) error
	RemoveFromBlacklist(ctx context.Context, ip string) error
	AddToWhitelist(ctx context.Context, ip string) error
	RemoveFromWhitelist(ctx context.Context, ip string) error
	CheckBlackList(ctx context.Context, ip string) (bool, error)
	CheckWhiteList(ctx context.Context, ip string) (bool, error)
}

type BucketUsecase struct {
	repository BucketRepository
	capacity   model.BucketCapacity
}

func NewBucketUsecase(r BucketRepository, c model.BucketCapacity) *BucketUsecase {
	return &BucketUsecase{r, c}
}

func (u BucketUsecase) GetBucket(ctx context.Context, bucket []model.Bucket) (bool, error) {
	return u.repository.GetSetBucket(ctx, bucket)
}

func (u BucketUsecase) Delete(ctx context.Context, keys []string) error {
	return u.repository.DeleteKeys(ctx, keys)
}

func (u BucketUsecase) AddToBlacklist(ctx context.Context, ip string) error {
	return u.repository.AddToBlacklist(ctx, ip)
}

func (u BucketUsecase) RemoveFromBlacklist(ctx context.Context, ip string) error {
	return u.repository.RemoveFromBlacklist(ctx, ip)
}

func (u BucketUsecase) AddToWhitelist(ctx context.Context, ip string) error {
	return u.repository.AddToWhitelist(ctx, ip)
}

func (u BucketUsecase) RemoveFromWhitelist(ctx context.Context, ip string) error {
	return u.repository.RemoveFromWhitelist(ctx, ip)
}

func (u BucketUsecase) CheckList(ctx context.Context, ip string) (*model.Included, error) {
	var wg sync.WaitGroup
	l := &model.Included{}
	ers := make(chan error, 2)
	defer close(ers)
	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("context timeout has occurred") //todo
	default:
		wg.Add(2)

		go func() {
			res, err := u.repository.CheckBlackList(ctx, ip)
			if err != nil {
				ers <- err
			}
			l.Blacklist = res
			wg.Done()
		}()

		go func() {
			res, err := u.repository.CheckWhiteList(ctx, ip)
			if err != nil {
				ers <- err
			}
			l.Whitelist = res
			wg.Done()
		}()
		wg.Wait()
	}

	select {
	case e := <-ers:
		return nil, e
	default:
		break
	}
	return l, nil
}

func (u BucketUsecase) GetBucketList(ctx context.Context, body *io.ReadCloser) ([]model.Bucket, error) {
	var req model.Request
	err := json.NewDecoder(*body).Decode(&req)
	if err != nil {
		return nil, fmt.Errorf("Usecase/bucket/GetBucketList: %w", err)
	}
	defaultTtl := 1 * time.Minute
	buckets := []model.Bucket{
		{
			Key:      req.Login,
			Capacity: u.capacity.Login,
			TTL:      defaultTtl,
		},
		{
			Key:      req.Password,
			Capacity: u.capacity.Password,
			TTL:      defaultTtl,
		},
		{
			Key:      req.IP,
			Capacity: u.capacity.IP,
			TTL:      defaultTtl,
		},
	}
	return buckets, nil
}
