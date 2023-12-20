// Copyright 2023 The Ryan SU Authors (https://github.com/suyuan32). All Rights Reserved.
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

package asynq

import (
	"fmt"
	"github.com/suyuan32/simple-admin-common/config"
	"time"

	"github.com/hibiken/asynq"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

// AsynqConf is the configuration struct for Asynq.
type AsynqConf struct {
	Addr         string `json:",default=127.0.0.1:6379"`
	Username     string `json:",optional"`
	Pass         string `json:",optional"`
	DB           int    `json:",optional,default=0"`
	Concurrency  int    `json:",optional,default=20"` // max concurrent process job task num
	SyncInterval int    `json:",optional,default=10"` // seconds, this field specifies how often sync should happen
	Enable       bool   `json:",default=true"`
}

// WithRedisConf sets redis configuration from RedisConf.
func (c *AsynqConf) WithRedisConf(r redis.RedisConf) *AsynqConf {
	c.Pass = r.Pass
	c.Addr = r.Host
	c.DB = 0
	return c
}

// WithOriginalRedisConf sets redis configuration from original RedisConf.
func (c *AsynqConf) WithOriginalRedisConf(r config.RedisConf) *AsynqConf {
	c.Pass = r.Pass
	c.Addr = r.Host
	c.Username = r.Username
	c.DB = r.Db
	return c
}

// NewRedisOpt returns a redis options from Asynq Configuration.
func (c *AsynqConf) NewRedisOpt() *asynq.RedisClientOpt {
	return &asynq.RedisClientOpt{
		Network:  "tcp",
		Addr:     c.Addr,
		Username: c.Username,
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
		return asynq.NewScheduler(c.NewRedisOpt(), &asynq.SchedulerOpts{Location: time.Local})
	} else {
		return nil
	}
}

// NewPeriodicTaskManager returns a periodic task manager from the configuration.
func (c *AsynqConf) NewPeriodicTaskManager(provider asynq.PeriodicTaskConfigProvider) *asynq.PeriodicTaskManager {
	if c.Enable {
		mgr, err := asynq.NewPeriodicTaskManager(
			asynq.PeriodicTaskManagerOpts{
				SchedulerOpts:              &asynq.SchedulerOpts{Location: time.Local},
				RedisConnOpt:               c.NewRedisOpt(),
				PeriodicTaskConfigProvider: provider,                                    // this provider object is the interface to your config source
				SyncInterval:               time.Duration(c.SyncInterval) * time.Second, // this field specifies how often sync should happen
			})
		logx.Must(err)
		return mgr
	} else {
		return nil
	}
}
