package repository

import (
	"context"
	"reflect"
	"testing"
	"time"

	miniredis "github.com/alicebob/miniredis/v2"
	"github.com/mxmntv/anti_bruteforce/internal/model"
	redis "github.com/redis/go-redis/v9"
)

var (
	redisServer *miniredis.Miniredis
	redisClient *redis.Client
)

func TestGetSetBucket(t *testing.T) {
	setup()
	defer teardown()
	type args struct {
		ctx    context.Context
		bucket map[string]model.Bucket
	}
	tests := []struct {
		name    string
		b       BucketRepository
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "Set key package",
			b:    BucketRepository{storage: redisClient},
			args: args{
				ctx: context.Background(),
				bucket: map[string]model.Bucket{
					"login": {
						Key:      "petya",
						Capacity: 2,
						TTL:      1 * time.Minute,
					},
					"pwd": {
						Key:      "qwerty",
						Capacity: 100,
						TTL:      1 * time.Minute,
					},
					"ip": {
						Key:      "10.40.210.253",
						Capacity: 1000,
						TTL:      1 * time.Minute,
					},
				},
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "Get keys",
			b:    BucketRepository{storage: redisClient},
			args: args{
				ctx: context.Background(),
				bucket: map[string]model.Bucket{
					"login": {
						Key:      "petya",
						Capacity: 2,
						TTL:      1 * time.Minute,
					},
					"pwd": {
						Key:      "qwerty",
						Capacity: 100,
						TTL:      1 * time.Minute,
					},
					"ip": {
						Key:      "10.40.210.253",
						Capacity: 1000,
						TTL:      1 * time.Minute,
					},
				},
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "Empty login bucket",
			b:    BucketRepository{storage: redisClient},
			args: args{
				ctx: context.Background(),
				bucket: map[string]model.Bucket{
					"login": {
						Key:      "petya",
						Capacity: 2,
						TTL:      1 * time.Minute,
					},
					"pwd": {
						Key:      "qwerty",
						Capacity: 100,
						TTL:      1 * time.Minute,
					},
					"ip": {
						Key:      "10.40.210.253",
						Capacity: 1000,
						TTL:      1 * time.Minute,
					},
				},
			},
			want:    false,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.b.GetSetBucket(tt.args.ctx, tt.args.bucket)
			if (err != nil) != tt.wantErr {
				t.Errorf("BucketRepository.GetSetBucket() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("BucketRepository.GetSetBucket() = %v, want %v", got, tt.want)
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

func TestAddToBlacklist(t *testing.T) {
	setup()
	defer teardown()
	type args struct {
		ctx context.Context
		ip  string
	}
	tests := []struct {
		name    string
		b       BucketRepository
		args    args
		wantErr bool
	}{
		{
			name:    "Add ip widthout error",
			b:       BucketRepository{storage: redisClient},
			args:    args{context.Background(), "192.0.2.2/24"},
			wantErr: false,
		},
		{
			name:    "Add ip width error",
			b:       BucketRepository{storage: redisClient},
			args:    args{context.Background(), "1000.40.210.253/11"},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.b.AddToBlacklist(tt.args.ctx, tt.args.ip); (err != nil) != tt.wantErr {
				t.Errorf("BucketRepository.AddToBlacklist() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRemoveFromBlacklist(t *testing.T) {
	setup()
	defer teardown()
	redisClient.LPush(context.Background(), "blacklist", "10.40.210.253/8")
	type fields struct {
		storage RedisStorage
	}
	type args struct {
		ctx context.Context
		ip  string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    int
		wantErr bool
	}{
		{
			name:    "Item not found",
			fields:  fields{redisClient},
			args:    args{context.Background(), "192.0.2.2/24"},
			want:    0,
			wantErr: false,
		},
		{
			name:    "Item found",
			fields:  fields{redisClient},
			args:    args{context.Background(), "10.40.210.253/8"},
			want:    1,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := BucketRepository{
				storage: tt.fields.storage,
			}
			got, err := b.RemoveFromBlacklist(tt.args.ctx, tt.args.ip)
			if (err != nil) != tt.wantErr {
				t.Errorf("BucketRepository.RemoveFromBlacklist() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("BucketRepository.RemoveFromBlacklist() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAddToWhitelist(t *testing.T) {
	setup()
	defer teardown()
	repo := BucketRepository{storage: redisClient}
	type args struct {
		ctx context.Context
		ip  string
	}
	tests := []struct {
		name    string
		b       BucketRepository
		args    args
		wantErr bool
	}{
		{
			name:    "Add ip widthout error",
			b:       repo,
			args:    args{context.Background(), "192.0.2.2/24"},
			wantErr: false,
		},
		{
			name:    "Add ip width error",
			b:       repo,
			args:    args{context.Background(), "1000.40.210.253/11"},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.b.AddToWhitelist(tt.args.ctx, tt.args.ip); (err != nil) != tt.wantErr {
				t.Errorf("BucketRepository.AddToWhitelist() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRemoveFromWhitelist(t *testing.T) {
	setup()
	defer teardown()
	repo := BucketRepository{storage: redisClient}
	redisClient.LPush(context.Background(), "whitelist", "10.40.210.253/8")
	type args struct {
		ctx context.Context
		ip  string
	}
	tests := []struct {
		name    string
		fields  BucketRepository
		args    args
		want    int
		wantErr bool
	}{
		{
			name:    "Item not found",
			fields:  repo,
			args:    args{context.Background(), "192.0.2.2/24"},
			want:    0,
			wantErr: false,
		},
		{
			name:    "Item found",
			fields:  repo,
			args:    args{context.Background(), "10.40.210.253/8"},
			want:    1,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := BucketRepository{
				storage: tt.fields.storage,
			}
			got, err := b.RemoveFromWhitelist(tt.args.ctx, tt.args.ip)
			if (err != nil) != tt.wantErr {
				t.Errorf("BucketRepository.RemoveFromWhitelist() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("BucketRepository.RemoveFromWhitelist() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCheckBlackList(t *testing.T) {
	setup()
	defer teardown()
	redisClient.LPush(context.Background(), "blacklist", "172.17.0.0/16")
	type args struct {
		ctx context.Context
		ip  string
	}
	tests := []struct {
		name    string
		meth    func(context.Context, string) (bool, error)
		b       BucketRepository
		args    args
		want    bool
		wantErr bool
	}{
		{
			name:    "Item not found in blacklist",
			b:       BucketRepository{redisClient},
			args:    args{context.Background(), "192.0.0.1"},
			want:    false,
			wantErr: false,
		},
		{
			name:    "Item found in blacklist",
			b:       BucketRepository{redisClient},
			args:    args{context.Background(), "172.17.0.2"},
			want:    true,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.b.CheckBlackList(tt.args.ctx, tt.args.ip)
			if (err != nil) != tt.wantErr {
				t.Errorf("BucketRepository.CheckBlackList() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("BucketRepository.CheckBlackList() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCheckWhiteList(t *testing.T) {
	setup()
	defer teardown()
	repo := BucketRepository{redisClient}
	redisClient.LPush(context.Background(), "whitelist", "172.17.0.0/16")
	type args struct {
		ctx context.Context
		ip  string
	}
	tests := []struct {
		name    string
		b       BucketRepository
		args    args
		want    bool
		wantErr bool
	}{
		{
			name:    "Item not found in whitelist",
			b:       repo,
			args:    args{context.Background(), "192.0.0.1"},
			want:    false,
			wantErr: false,
		},
		{
			name:    "Item found in whitelist",
			b:       repo,
			args:    args{context.Background(), "172.17.0.2"},
			want:    true,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.b.CheckWhiteList(tt.args.ctx, tt.args.ip)
			if (err != nil) != tt.wantErr {
				t.Errorf("BucketRepository.CheckWhiteList() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("BucketRepository.CheckWhiteList() = %v, want %v", got, tt.want)
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
