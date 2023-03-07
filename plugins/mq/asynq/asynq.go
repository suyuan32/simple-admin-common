package asynq

import (
	"fmt"

	"github.com/hibiken/asynq"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

// AsynqConf is the configuration struct for Asynq.
type AsynqConf struct {
	Addr        string `json:",default=127.0.0.1:6379"`
	Pass        string `json:",optional"`
	DB          int    `json:",optional,default=0"`
	Concurrency int    `json:",optional,default=20"` // max concurrent process job task num
	Enable      bool   `json:",default=true"`
	Location    string `json:",default="`
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

// NewServer returns a worker from the configuration.
func (c *AsynqConf) NewServer() *asynq.Server {
	if c.Enable {
		return asynq.NewServer(
			c.NewRedisOpt(),
			asynq.Config{
				IsFailure: func(err error) bool {
					fmt.Printf("failed to exec asynq task, err : %+v \n", err)
					return true
				},
				Concurrency: c.Concurrency,
			},
		)
	} else {
		return nil
	}
}

// NewScheduler returns a scheduler from the configuration.
func (c *AsynqConf) NewScheduler() *asynq.Scheduler {
	if c.Enable {
		return asynq.NewScheduler(c.NewRedisOpt(), nil)
	} else {
		return nil
	}
}
