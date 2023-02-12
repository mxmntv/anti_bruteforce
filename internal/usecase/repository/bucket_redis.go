package repository

import (
	"context"
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/mxmntv/anti_bruteforce/internal/model"
	"github.com/mxmntv/anti_bruteforce/pkg/logger"
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
	logger  logger.LogInterface
}

func NewBucketRepo(r RedisStorage, l logger.LogInterface) *BucketRepository {
	return &BucketRepository{
		storage: r,
		logger:  l,
	}
}

func (b BucketRepository) GetSetBucket(ctx context.Context, bucket map[string]model.Bucket) (bool, error) {
	var bslice []model.Bucket
	cmds, _ := b.storage.Pipelined(ctx, func(pipe redis.Pipeliner) error {
		for _, b := range bucket {
			pipe.Get(ctx, b.Key)
			bslice = append(bslice, b)
		}
		return nil
	})
	var c int

	for i, cmd := range cmds {
		val, err := cmd.(*redis.StringCmd).Int()
		arg := cmd.Args()[1].(string)
		switch {
		case errors.Is(err, redis.Nil):
			err := b.storage.Set(ctx, arg, bslice[i].Capacity-1, bslice[i].TTL).Err()
			if err != nil {
				return false, err // todo wrap error
			}
		case err != nil:
			return false, err // todo wrap error
		case val == 0:
			c++
		case val > 0:
			err := b.storage.Decr(ctx, arg).Err()
			if err != nil {
				return false, err // todo wrap error
			}
		}
	}
	if c > 0 {
		return false, nil
	}
	return true, nil
}

func (b BucketRepository) DeleteKeys(ctx context.Context, keys []string) error {
	_, err := b.storage.Pipelined(ctx, func(pipe redis.Pipeliner) error {
		for _, key := range keys {
			pipe.Del(ctx, key)
		}
		return nil
	})
	if err != nil {
		return err // todo wrap err
	}
	return nil
}

func (b BucketRepository) AddToBlacklist(ctx context.Context, ip string) error {
	if err := b.storage.LPush(ctx, "blacklist", ip).Err(); err != nil {
		return err
	}
	return nil
}

func (b BucketRepository) RemoveFromBlacklist(ctx context.Context, ip string) error {
	res, err := b.storage.LRem(ctx, "blacklist", 0, ip).Result()
	if err != nil {
		return err
	}
	fmt.Println(res) // todo log success status
	return nil
}

func (b BucketRepository) AddToWhitelist(ctx context.Context, ip string) error {
	if err := b.storage.LPush(ctx, "whitelist", ip).Err(); err != nil {
		return err
	}
	return nil
}

func (b BucketRepository) RemoveFromWhitelist(ctx context.Context, ip string) error {
	res, err := b.storage.LRem(ctx, "whitelist", 0, ip).Result()
	if err != nil {
		return err
	}
	fmt.Println(res) // todo log success status
	return nil
}

func (b BucketRepository) CheckBlackList(ctx context.Context, ip string) (bool, error) {
	list, err := b.storage.LRange(ctx, "blacklist", 0, -1).Result()
	if err != nil {
		fmt.Println(err) // todo return err
	}
	targetip := net.ParseIP(ip)
	for _, el := range list {
		_, ipnet, err := net.ParseCIDR(el)
		if err != nil {
			return false, err
		}
		if ipnet.Contains(targetip) {
			return true, nil
		}
	}
	return false, nil
}

func (b BucketRepository) CheckWhiteList(ctx context.Context, ip string) (bool, error) {
	list, err := b.storage.LRange(ctx, "whitelist", 0, -1).Result()
	if err != nil {
		fmt.Println(err)
	}
	targetip := net.ParseIP(ip)
	for _, el := range list {
		_, ipnet, err := net.ParseCIDR(el)
		if err != nil {
			return false, err
		}
		if ipnet.Contains(targetip) {
			return true, nil
		}
	}
	return false, nil
}
