package dynamicconf

import (
	"testing"

	"github.com/redis/go-redis/v9"

	"github.com/suyuan32/simple-admin-common/config"
)

func TestSetDynamicConfigurationToRedis(t *testing.T) {
	t.Skip("skip TestSetDynamicConfigurationToRedis")

	conf := config.RedisConf{
		Host:     "localhost:6379",
		Db:       0,
		Username: "",
		Pass:     "",
		Tls:      false,
		Master:   "",
	}

	redisClient := conf.MustNewUniversalRedis()

	type args struct {
		rds      redis.UniversalClient
		category string
		key      string
		value    string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "test1",
			args: args{
				rds:      redisClient,
				category: "system",
				key:      "logo",
				value:    "https://ryansu.tech/logo.jpg",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := SetDynamicConfigurationToRedis(tt.args.rds, tt.args.category, tt.args.key, tt.args.value); (err != nil) != tt.wantErr {
				t.Errorf("SetDynamicConfigurationToRedis() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGetDynamicConfigurationToRedis(t *testing.T) {
	t.Skip("skip TestGetDynamicConfigurationToRedis")

	conf := config.RedisConf{
		Host:     "localhost:6379",
		Db:       0,
		Username: "",
		Pass:     "",
		Tls:      false,
		Master:   "",
	}

	redisClient := conf.MustNewUniversalRedis()

	type args struct {
		rds      redis.UniversalClient
		category string
		key      string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "test2",
			args: args{
				rds:      redisClient,
				category: "system",
				key:      "logo",
			},
			want:    "https://ryansu.tech/logo.jpg",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetDynamicConfigurationToRedis(tt.args.rds, tt.args.category, tt.args.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetDynamicConfigurationToRedis() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetDynamicConfigurationToRedis() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSetTenantDynamicConfigurationToRedis(t *testing.T) {
	t.Skip("skip TestSetTenantDynamicConfigurationToRedis")

	conf := config.RedisConf{
		Host:     "localhost:6379",
		Db:       0,
		Username: "",
		Pass:     "",
		Tls:      false,
		Master:   "",
	}

	redisClient := conf.MustNewUniversalRedis()

	type args struct {
		rds      redis.UniversalClient
		tenantId string
		category string
		key      string
		value    string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "test2",
			args: args{
				rds:      redisClient,
				tenantId: "1",
				category: "system",
				key:      "logo",
				value:    "https://ryansu.tech/logo.jpg",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := SetTenantDynamicConfigurationToRedis(tt.args.rds, tt.args.tenantId, tt.args.category, tt.args.key, tt.args.value); (err != nil) != tt.wantErr {
				t.Errorf("SetTenantDynamicConfigurationToRedis() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGetTenantDynamicConfigurationToRedis(t *testing.T) {

	t.Skip("skip TestGetTenantDynamicConfigurationToRedis")

	conf := config.RedisConf{
		Host:     "localhost:6379",
		Db:       0,
		Username: "",
		Pass:     "",
		Tls:      false,
		Master:   "",
	}

	redisClient := conf.MustNewUniversalRedis()

	type args struct {
		rds      redis.UniversalClient
		tenantId string
		category string
		key      string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "test4",
			args: args{
				rds:      redisClient,
				tenantId: "1",
				category: "system",
				key:      "logo",
			},
			want:    "https://ryansu.tech/logo.jpg",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetTenantDynamicConfigurationToRedis(tt.args.rds, tt.args.tenantId, tt.args.category, tt.args.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetTenantDynamicConfigurationToRedis() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetTenantDynamicConfigurationToRedis() got = %v, want %v", got, tt.want)
			}
		})
	}
}
