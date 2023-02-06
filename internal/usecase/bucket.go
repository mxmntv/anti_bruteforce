package usecase

import (
	"context"
	"sync"

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
}

func NewBucketUsecase(r BucketRepository) *BucketUsecase {
	return &BucketUsecase{r}
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
	select {
	case e := <-ers:
		return nil, e
	default:
		break
	}
	return l, nil
}
