// Copyright 2023 The Ryan SU Authors. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package captcha

import (
	"context"
	"github.com/redis/go-redis/v9"
	"image/color"
	"time"

	"github.com/mojocn/base64Captcha"
	"github.com/zeromicro/go-zero/core/logx"
)

const Prefix = "CAPTCHA:"

// NewRedisStore returns a redis store for captcha.
func NewRedisStore(r *redis.Client) *RedisStore {
	return &RedisStore{
		Expiration: time.Minute * 5,
		PreKey:     Prefix,
		Redis:      r,
	}
}

// RedisStore stores captcha data.
type RedisStore struct {
	Expiration time.Duration
	PreKey     string
	Context    context.Context
	Redis      *redis.Client
}

// UseWithCtx add context for captcha.
func (r *RedisStore) UseWithCtx(ctx context.Context) base64Captcha.Store {
	r.Context = ctx
	return r
}

// Set sets the captcha KV to redis.
func (r *RedisStore) Set(id string, value string) error {
	err := r.Redis.Set(context.Background(), r.PreKey+id, value, r.Expiration).Err()
	if err != nil {
		logx.Errorw("error occurs when captcha key sets to redis", logx.Field("detail", err))
		return err
	}
	return nil
}

// Get gets the captcha KV from redis.
func (r *RedisStore) Get(key string, clear bool) string {
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
func (r *RedisStore) Verify(id, answer string, clear bool) bool {
	key := r.PreKey + id
	v := r.Get(key, clear)
	return v == answer
}

// MustNewRedisCaptcha returns the captcha using redis, it will exit when error occur
func MustNewRedisCaptcha(c Conf, r *redis.Client) *base64Captcha.Captcha {
	driver := NewDriver(c)

	store := NewRedisStore(r)

	return base64Captcha.NewCaptcha(driver, store)
}

func NewDriver(c Conf) base64Captcha.Driver {
	var driver base64Captcha.Driver

	bgColor := &color.RGBA{
		R: 254,
		G: 254,
		B: 254,
		A: 254,
	}

	fonts := []string{
		"ApothecaryFont.ttf",
		"DENNEthree-dee.ttf",
		"Flim-Flam.ttf",
		"RitaSmith.ttf",
		"actionj.ttf",
		"chromohv.ttf",
	}

	switch c.Driver {
	case "digit":
		driver = base64Captcha.NewDriverDigit(c.ImgHeight, c.ImgWidth,
			c.KeyLong, 0.7, 80)
	case "string":
		driver = base64Captcha.NewDriverString(c.ImgHeight, c.ImgWidth, 12, 3, c.KeyLong,
			"qwertyupasdfghjkzxcvbnm23456789",
			bgColor, nil, fonts)
	case "math":
		driver = base64Captcha.NewDriverMath(c.ImgHeight, c.ImgWidth, 12, 3, bgColor,
			nil, fonts)
	case "chinese":
		driver = base64Captcha.NewDriverChinese(c.ImgHeight, c.ImgWidth, 10, 4, c.KeyLong,
			"天地玄黄宇宙洪荒日月盈辰宿列张寒来暑往秋收冬藏闰余成岁律吕调阳夫大水浸灌草木破坡石右云虫师军舰流浪数据速度", bgColor, nil,
			[]string{"wqy-microhei.ttc"})
	default:
		driver = base64Captcha.NewDriverDigit(c.ImgHeight, c.ImgWidth,
			c.KeyLong, 0.7, 80)
	}

	return driver
}
