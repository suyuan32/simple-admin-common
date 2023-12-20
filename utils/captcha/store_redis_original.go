package captcha

import (
	"context"
	"github.com/mojocn/base64Captcha"
	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/logx"
	"time"
)

// NewOriginalRedisStore returns a redis store for captcha.
func NewOriginalRedisStore(r *redis.Client) *OriginalRedisStore {
	return &OriginalRedisStore{
		Expiration: time.Minute * 5,
		PreKey:     Prefix,
		Redis:      r,
	}
}

// OriginalRedisStore stores captcha data.
type OriginalRedisStore struct {
	Expiration time.Duration
	PreKey     string
	Context    context.Context
	Redis      *redis.Client
}

// UseWithCtx add context for captcha.
func (r *OriginalRedisStore) UseWithCtx(ctx context.Context) base64Captcha.Store {
	r.Context = ctx
	return r
}

// Set sets the captcha KV to redis.
func (r *OriginalRedisStore) Set(id string, value string) error {
	err := r.Redis.Set(context.Background(), r.PreKey+id, value, r.Expiration).Err()
	if err != nil {
		logx.Errorw("error occurs when captcha key sets to redis", logx.Field("detail", err))
		return err
	}
	return nil
}

// Get gets the captcha KV from redis.
func (r *OriginalRedisStore) Get(key string, clear bool) string {
	val, err := r.Redis.Get(context.Background(), key).Result()
	if err != nil {
		logx.Errorw("error occurs when captcha key gets from redis", logx.Field("detail", err))
		return ""
	}
	if clear {
		_, err := r.Redis.Del(context.Background(), key).Result()
		if err != nil {
			logx.Errorw("error occurs when captcha key deletes from redis", logx.Field("detail", err))
			return ""
		}
	}
	return val
}

// Verify verifies the captcha whether it is correct.
func (r *OriginalRedisStore) Verify(id, answer string, clear bool) bool {
	key := r.PreKey + id
	v := r.Get(key, clear)
	return v == answer
}
