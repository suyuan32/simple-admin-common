package asynq

import (
	"github.com/hibiken/asynq"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

// AsynqConf is the configuration struct for Asynq.
type AsynqConf struct {
	Addr   string `json:",default=127.0.0.1:6379"`
	Pass   string `json:",optional"`
	DB     int    `json:",optional,default=0"`
	Enable bool   `json:",default=true"`
}

// WithRedisConf sets redis configuration from RedisConf.
func (c *AsynqConf) WithRedisConf(r redis.RedisConf) *AsynqConf {
	c.Pass = r.Pass
	c.Addr = r.Host
	c.DB = 0
	return c
}

// NewRedisOpt returns a redis options from Asynq Configuration.
func (c *AsynqConf) NewRedisOpt() *asynq.RedisClientOpt {
	return &asynq.RedisClientOpt{
		Network:  "tcp",
		Addr:     c.Addr,
		Username: "",
		Password: c.Pass,
		DB:       c.DB,
	}
}

// NewClient returns a client from the configuration.
func (c *AsynqConf) NewClient() *asynq.Client {
	if c.Enable {
		return asynq.NewClient(c.NewRedisOpt())
	} else {
		return nil
	}
}

// NewWorker returns a worker from the configuration.
func (c *AsynqConf) NewWorker(cfg asynq.Config) *asynq.Server {
	if c.Enable {
		return asynq.NewServer(c.NewRedisOpt(), cfg)
	} else {
		return nil
	}
}
