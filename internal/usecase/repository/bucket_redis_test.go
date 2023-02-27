package repository

import (
	"context"
	"fmt"
	"reflect"
	"testing"
	"time"

	miniredis "github.com/alicebob/miniredis/v2"
	internalError "github.com/mxmntv/anti_bruteforce/internal/errors"
	"github.com/mxmntv/anti_bruteforce/internal/model"
	redis "github.com/redis/go-redis/v9"
)

var (
	redisServer *miniredis.Miniredis
	redisClient *redis.Client
)

func TestSetBucket(t *testing.T) {
	setup()
	defer teardown()
	type fields struct {
		storage RedisStorage
	}
	type args struct {
		ctx    context.Context
		bucket model.Bucket
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:   "Set name bucket",
			fields: fields{redisClient},
			args: args{
				ctx: context.Background(),
				bucket: model.Bucket{
					Key:      "Petya",
					Capacity: 10,
					TTL:      time.Minute,
				},
			},
			wantErr: false,
		},
		{
			name:   "Set IP bucket",
			fields: fields{redisClient},
			args: args{
				ctx: context.Background(),
				bucket: model.Bucket{
					Key:      "10.40.210.253",
					Capacity: 10,
					TTL:      time.Minute,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := BucketRepository{
				storage: tt.fields.storage,
			}
			if err := b.SetBucket(tt.args.ctx, tt.args.bucket); (err != nil) != tt.wantErr {
				t.Errorf("BucketRepository.SetBucket() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDecrementValue(t *testing.T) {
	setup()
	defer teardown()
	redisClient.Set(context.Background(), "qwerty", 10, time.Minute)
	type fields struct {
		storage RedisStorage
	}
	type args struct {
		ctx context.Context
		key string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    int
		wantErr bool
	}{
		{
			name:   "Decrement value",
			fields: fields{redisClient},
			args: args{
				ctx: context.Background(),
				key: "qwerty",
			},
			want:    9,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := BucketRepository{
				storage: tt.fields.storage,
			}
			if err := b.DecrementValue(tt.args.ctx, tt.args.key); (err != nil) != tt.wantErr {
				t.Errorf("BucketRepository.DecrementValue() error = %v, wantErr %v", err, tt.wantErr)
			}
			got, _ := redisClient.Get(context.Background(), "qwerty").Int()
			if got != tt.want {
				t.Errorf("BucketRepository.DecrementValue() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetBucket(t *testing.T) {
	setup()
	defer teardown()
	redisClient.Set(context.Background(), "Fedya", 10, time.Minute)
	type fields struct {
		storage RedisStorage
	}
	type args struct {
		ctx  context.Context
		keys []string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []model.BucketInDB
		wantErr bool
	}{
		{
			name:   "Get bucket",
			fields: fields{redisClient},
			args: args{
				ctx:  context.Background(),
				keys: []string{"Fedya"},
			},
			want: []model.BucketInDB{
				{
					Key:   "Fedya",
					Value: 10,
					Error: nil,
				},
			},
			wantErr: false,
		},
		{
			name:   "Get bucket width error",
			fields: fields{redisClient},
			args: args{
				ctx:  context.Background(),
				keys: []string{"Petya"},
			},
			want: []model.BucketInDB{
				{
					Key:   "Petya",
					Value: 0,
					Error: fmt.Errorf("[getBucket] error: %w", internalError.ErrorDBNotFound),
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := BucketRepository{
				storage: tt.fields.storage,
			}
			got, err := b.GetBucket(tt.args.ctx, tt.args.keys)
			if (err != nil) != tt.wantErr {
				t.Errorf("BucketRepository.GetBucket() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BucketRepository.GetBucket() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDeleteKeys(t *testing.T) {
	setup()
	defer teardown()
	redisClient.Set(context.Background(), "qwerty", 0, 0)
	redisClient.Set(context.Background(), "qazwsx", 0, 0)
	type fields struct {
		storage RedisStorage
	}
	type args struct {
		ctx  context.Context
		keys []string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []string
		wantErr bool
	}{
		{
			name:    "Delete qwerty & qazwsx",
			fields:  fields{redisClient},
			args:    args{context.Background(), []string{"qwerty", "qazwsx"}},
			want:    []string{"qwerty", "qazwsx"},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := BucketRepository{
				storage: tt.fields.storage,
			}
			got, err := b.DeleteKeys(tt.args.ctx, tt.args.keys)
			if (err != nil) != tt.wantErr {
				t.Errorf("BucketRepository.DeleteKeys() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BucketRepository.DeleteKeys() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRemoveFromList(t *testing.T) {
	setup()
	defer teardown()
	redisClient.LPush(context.Background(), "blacklist", "10.40.210.253/8")
	redisClient.LPush(context.Background(), "whitelist", "11.40.210.253/6")
	type fields struct {
		storage RedisStorage
	}
	type args struct {
		ctx  context.Context
		ip   string
		list string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    int
		wantErr bool
	}{
		{
			name:    "Remove from blacklist",
			fields:  fields{redisClient},
			args:    args{context.Background(), "10.40.210.253/8", "blacklist"},
			want:    1,
			wantErr: false,
		},
		{
			name:    "Remove from whitelist",
			fields:  fields{redisClient},
			args:    args{context.Background(), "11.40.210.253/6", "whitelist"},
			want:    1,
			wantErr: false,
		},
		{
			name:    "Remove from somelist",
			fields:  fields{redisClient},
			args:    args{context.Background(), "11.40.210.253/6", "somelist"},
			want:    0,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := BucketRepository{
				storage: tt.fields.storage,
			}
			got, err := b.RemoveFromList(tt.args.ctx, tt.args.ip, tt.args.list)
			if (err != nil) != tt.wantErr {
				t.Errorf("BucketRepository.RemoveFromList() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("BucketRepository.RemoveFromList() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAddToList(t *testing.T) {
	setup()
	defer teardown()
	type fields struct {
		storage RedisStorage
	}
	type args struct {
		ctx  context.Context
		ip   string
		list string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:    "Add to blacklist",
			fields:  fields{redisClient},
			args:    args{context.Background(), "10.40.210.253/8", "blacklist"},
			wantErr: false,
		},
		{
			name:    "Add to whitelist",
			fields:  fields{redisClient},
			args:    args{context.Background(), "11.40.210.253/6", "whitelist"},
			wantErr: false,
		},
		{
			name:    "Add to somelist (error)",
			fields:  fields{redisClient},
			args:    args{context.Background(), "11.40.210.253/6", "somelist"},
			wantErr: true,
		},
		{
			name:    "Add broken ip whitelist (error)",
			fields:  fields{redisClient},
			args:    args{context.Background(), "1100.40.210.253/6", "whitelist"},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := BucketRepository{
				storage: tt.fields.storage,
			}
			if err := b.AddToList(tt.args.ctx, tt.args.ip, tt.args.list); (err != nil) != tt.wantErr {
				t.Errorf("BucketRepository.AddToList() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCheckList(t *testing.T) {
	setup()
	defer teardown()
	redisClient.LPush(context.Background(), "blacklist", "172.17.0.0/16")
	redisClient.LPush(context.Background(), "whitelist", "192.168.0.0/16")
	type fields struct {
		storage RedisStorage
	}
	type args struct {
		ctx  context.Context
		ip   string
		list string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    bool
		wantErr bool
	}{
		{
			name:    "Found in blacklist",
			fields:  fields{redisClient},
			args:    args{context.Background(), "172.17.0.2", "blacklist"},
			want:    true,
			wantErr: false,
		},
		{
			name:    "Not found in blacklist",
			fields:  fields{redisClient},
			args:    args{context.Background(), "192.0.0.1", "blacklist"},
			want:    false,
			wantErr: false,
		},
		{
			name:    "Found in whitelist",
			fields:  fields{redisClient},
			args:    args{context.Background(), "192.168.0.1", "whitelist"},
			want:    true,
			wantErr: false,
		},
		{
			name:    "Not found in whitelist",
			fields:  fields{redisClient},
			args:    args{context.Background(), "172.17.0.2", "whitelist"},
			want:    false,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := BucketRepository{
				storage: tt.fields.storage,
			}
			got, err := b.CheckList(tt.args.ctx, tt.args.ip, tt.args.list)
			if (err != nil) != tt.wantErr {
				t.Errorf("BucketRepository.CheckList() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("BucketRepository.CheckList() = %v, want %v", got, tt.want)
			}
		})
	}
}

func mockRedis() *miniredis.Miniredis {
	s, err := miniredis.Run()
	if err != nil {
		panic(err)
	}
	return s
}

func setup() {
	redisServer = mockRedis()
	redisClient = redis.NewClient(&redis.Options{
		Addr: redisServer.Addr(),
	})
}

func teardown() {
	redisServer.Close()
}
