package repository

import (
	"context"
	"errors"
	"fmt"
	"net"
	"time"

	internalError "github.com/mxmntv/anti_bruteforce/internal/errors"
	"github.com/mxmntv/anti_bruteforce/internal/model"
	redis "github.com/redis/go-redis/v9"
)

type RedisStorage interface {
	Get(ctx context.Context, key string) *redis.StringCmd
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd
	Decr(ctx context.Context, key string) *redis.IntCmd
	Pipelined(ctx context.Context, fn func(redis.Pipeliner) error) ([]redis.Cmder, error)
	Del(ctx context.Context, keys ...string) *redis.IntCmd
	LPush(ctx context.Context, key string, values ...interface{}) *redis.IntCmd
	LRem(ctx context.Context, key string, count int64, value interface{}) *redis.IntCmd
	LRange(ctx context.Context, key string, start int64, stop int64) *redis.StringSliceCmd
}

type BucketRepository struct {
	storage RedisStorage
}

func NewBucketRepo(r RedisStorage) *BucketRepository {
	return &BucketRepository{
		storage: r,
	}
}

func (b BucketRepository) SetBucket(ctx context.Context, bucket model.Bucket) error {
	err := b.storage.Set(ctx, bucket.Key, bucket.Capacity-1, bucket.TTL).Err()
	if err != nil {
		return fmt.Errorf("[setBucket] error: %w", internalError.ErrorInternalDB)
	}
	return nil
}

func (b BucketRepository) DecrementValue(ctx context.Context, key string) error {
	err := b.storage.Decr(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("[decrementValue] error: %w", internalError.ErrorInternalDB)
	}
	return nil
}

func (b BucketRepository) GetBucket(ctx context.Context, keys []string) ([]model.BucketInDB, error) {
	bucketsInDatabase := make([]model.BucketInDB, 0, len(keys))
	cmds, _ := b.storage.Pipelined(ctx, func(pipe redis.Pipeliner) error {
		for _, b := range keys {
			pipe.Get(ctx, b)
		}
		return nil
	})

	for _, cmd := range cmds {
		val, err := cmd.(*redis.StringCmd).Int()
		bucketInDB := model.BucketInDB{
			Key:   cmd.Args()[1].(string),
			Value: val,
		}
		switch {
		case errors.Is(err, redis.Nil):
			bucketInDB.Error = fmt.Errorf("[getBucket] error: %w", internalError.ErrorDBNotFound)
		case err != nil:
			bucketInDB.Error = fmt.Errorf("[getBucket - cmds] error: %w", internalError.ErrorInternalDB)
		default:
			bucketInDB.Error = nil
		}
		bucketsInDatabase = append(bucketsInDatabase, bucketInDB)
	}
	return bucketsInDatabase, nil
}

func (b BucketRepository) DeleteKeys(ctx context.Context, keys []string) ([]string, error) {
	var delresult []string
	cmds, err := b.storage.Pipelined(ctx, func(pipe redis.Pipeliner) error {
		for _, key := range keys {
			pipe.Del(ctx, key)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("[deleteKeys] error: %w", internalError.ErrorInternalDB)
	}

	for _, c := range cmds {
		result := c.(*redis.IntCmd).Val()
		if result > 0 {
			delresult = append(delresult, c.Args()[1].(string))
		}
	}
	return delresult, nil
}

func (b BucketRepository) RemoveFromList(ctx context.Context, ip string, list string) (int, error) {
	if err := CheckListName(list); err != nil {
		return 0, fmt.Errorf("[removeFromList] error: %w", err)
	}
	res, err := b.storage.LRem(ctx, list, 0, ip).Result()
	if err != nil {
		return 0, fmt.Errorf("[removeFrom%s] error: %w", list, internalError.ErrorInternalDB)
	}
	return int(res), nil
}

func (b BucketRepository) AddToList(ctx context.Context, ip string, list string) error {
	if err := CheckListName(list); err != nil {
		return fmt.Errorf("[addToList] error: %w", err)
	}
	if err := CheckIPNet(ip); err != nil {
		return fmt.Errorf("[addTo%s] ip: %s error: %w", list, ip, internalError.ErrorInvalidIPNet)
	}
	if err := b.storage.LPush(ctx, list, ip).Err(); err != nil {
		return fmt.Errorf("[addTo%s] error: %w", list, internalError.ErrorInternalDB)
	}
	return nil
}

func (b BucketRepository) CheckList(ctx context.Context, ip string, list string) (bool, error) {
	if err := CheckListName(list); err != nil {
		return false, fmt.Errorf("[checkList] error: %w", err)
	}
	if err := CheckIP(ip); err != nil {
		return false, fmt.Errorf("[check%s] ip: %s error: %w", list, ip, internalError.ErrorInvalidIP)
	}
	ipList, err := b.storage.LRange(ctx, list, 0, -1).Result()
	if err != nil {
		return false, fmt.Errorf("[check%s - getList] error: %w", list, internalError.ErrorInternalDB)
	}
	targetIP := net.ParseIP(ip)
	for _, el := range ipList {
		_, ipnet, err := net.ParseCIDR(el)
		if err != nil {
			return false, fmt.Errorf("[check%s - parseIP] error: %w", list, internalError.ErrorInternalDB)
		}
		if ipnet.Contains(targetIP) {
			return true, nil
		}
	}
	return false, nil
}
