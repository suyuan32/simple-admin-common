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

package dynamicconf

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"

	"github.com/suyuan32/simple-admin-common/config"
)

// SetDynamicConfigurationToRedis sets configuration data to redis
func SetDynamicConfigurationToRedis(rds redis.UniversalClient, category, key, value string) error {
	return rds.Set(context.Background(), fmt.Sprintf("%s%s:%s", config.RedisDynamicConfigurationPrefix, category, key), value, redis.KeepTTL).Err()
}

// GetDynamicConfigurationToRedis gets configuration data from redis
func GetDynamicConfigurationToRedis(rds redis.UniversalClient, category, key string) (string, error) {
	return rds.Get(context.Background(), fmt.Sprintf("%s%s:%s", config.RedisDynamicConfigurationPrefix, category, key)).Result()
}

// SetTenantDynamicConfigurationToRedis sets configuration data to redis by tenant ID
func SetTenantDynamicConfigurationToRedis(rds redis.UniversalClient, tenantId, category, key, value string) error {
	return rds.Set(context.Background(), fmt.Sprintf("%s%s:%s:%s", config.RedisDynamicConfigurationPrefix, tenantId, category, key), value, redis.KeepTTL).Err()
}

// GetTenantDynamicConfigurationToRedis gets configuration data from redis by tenant ID
func GetTenantDynamicConfigurationToRedis(rds redis.UniversalClient, tenantId, category, key string) (string, error) {
	return rds.Get(context.Background(), fmt.Sprintf("%s%s:%s:%s", config.RedisDynamicConfigurationPrefix, tenantId, category, key)).Result()
}
