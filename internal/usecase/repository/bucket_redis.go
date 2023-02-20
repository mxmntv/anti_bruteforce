package repository

import (
	"context"
	"errors"
	"fmt"
	"net"
	"time"

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
				return false, fmt.Errorf("repo - set bucket - set key failed: %w", err)
			}
		case err != nil:
			return false, fmt.Errorf("repo - set bucket - key request failed: %w", err)
		case val == 0:
			c++
		case val > 0:
			err := b.storage.Decr(ctx, arg).Err()
			if err != nil {
				return false, fmt.Errorf("repo - set bucket - key decrement failed: %w", err)
			}
		}
	}
	if c > 0 {
		return false, nil
	}
	return true, nil
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
		return delresult, fmt.Errorf("repo - delete keys - key delete failed: %w", err)
	}

	for _, c := range cmds {
		val := c.(*redis.IntCmd).Val()
		if val > 0 {
			delresult = append(delresult, c.Args()[1].(string))
		}
	}
	return delresult, nil
}

func (b BucketRepository) AddToBlacklist(ctx context.Context, ip string) error {
	if err := CheckIPNet(ip); err != nil {
		return err
	}
	if err := b.storage.LPush(ctx, "blacklist", ip).Err(); err != nil {
		return fmt.Errorf("repo - add to blacklist - key push failed: %w", err)
	}
	return nil
}

func (b BucketRepository) RemoveFromBlacklist(ctx context.Context, ip string) (int, error) {
	res, err := b.storage.LRem(ctx, "blacklist", 0, ip).Result()
	if err != nil {
		return 0, fmt.Errorf("repo - remove from blacklist - key remove failed: %w", err)
	}
	return int(res), nil
}

func (b BucketRepository) AddToWhitelist(ctx context.Context, ip string) error {
	if err := CheckIPNet(ip); err != nil {
		return err
	}
	if err := b.storage.LPush(ctx, "whitelist", ip).Err(); err != nil {
		return fmt.Errorf("repo - add to whitelist - key push failed: %w", err)
	}
	return nil
}

func (b BucketRepository) RemoveFromWhitelist(ctx context.Context, ip string) (int, error) {
	res, err := b.storage.LRem(ctx, "whitelist", 0, ip).Result()
	if err != nil {
		return 0, fmt.Errorf("repo - remove from whitelist - key remove failed: %w", err)
	}
	return int(res), nil
}

func (b BucketRepository) CheckBlackList(ctx context.Context, ip string) (bool, error) {
	if err := CheckIP(ip); err != nil {
		return true, err
	}
	list, err := b.storage.LRange(ctx, "blacklist", 0, -1).Result()
	if err != nil {
		return true, fmt.Errorf("repo - check blacklist - get blacklist failed: %w", err)
	}
	targetip := net.ParseIP(ip)
	for _, el := range list {
		_, ipnet, err := net.ParseCIDR(el)
		if err != nil {
			return true, fmt.Errorf("repo - check blacklist - parse ip failed: %w", err)
		}
		if ipnet.Contains(targetip) {
			return true, nil
		}
	}
	return false, nil
}

func (b BucketRepository) CheckWhiteList(ctx context.Context, ip string) (bool, error) {
	if err := CheckIP(ip); err != nil {
		return false, err
	}
	list, err := b.storage.LRange(ctx, "whitelist", 0, -1).Result()
	if err != nil {
		return false, fmt.Errorf("repo - check whitelist - get whitelist failed: %w", err)
	}
	targetip := net.ParseIP(ip)
	for _, el := range list {
		_, ipnet, err := net.ParseCIDR(el)
		if err != nil {
			return false, fmt.Errorf("repo - check whitelist - parse ip failed: %w", err)
		}
		if ipnet.Contains(targetip) {
			return true, nil
		}
	}
	return false, nil
}
