package config

import (
	"context"
	"testing"
	"time"
)

func TestRedis(t *testing.T) {
	// Test must redis
	conf := &RedisConf{
		Host:     "localhost:6379",
		Db:       0,
		Username: "",
		Pass:     "",
		Tls:      false,
	}

	rds := conf.MustNewRedis()

	rds.Set(context.Background(), "testKeyDb0", "testVal", 2*time.Minute)

	conf1 := &RedisConf{
		Host:     "localhost:6379",
		Db:       1,
		Username: "",
		Pass:     "",
		Tls:      false,
	}

	rds1 := conf1.MustNewRedis()

	rds1.Set(context.Background(), "testKeyDb1", "testVal", 2*time.Minute)
}
