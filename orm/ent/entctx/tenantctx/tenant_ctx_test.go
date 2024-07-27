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

package tenantctx

import (
	"context"
	"testing"

	"github.com/suyuan32/simple-admin-common/orm/ent/entenum"
	"github.com/zeromicro/go-zero/rest/enum"
	"google.golang.org/grpc/metadata"
)

func TestGetTenantAdminCtx(t *testing.T) {
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "test tenant admin context",
			args: args{ctx: context.WithValue(context.Background(), TenantAdmin, "allow")},
			want: true,
		},
		{
			name: "test tenant admin wrong context",
			args: args{ctx: context.WithValue(context.Background(), TenantAdmin, "allowing")},
			want: false,
		},
		{
			name: "test tenant empty context",
			args: args{ctx: context.Background()},
			want: false,
		},
		{
			name: "test tenant admin context function",
			args: args{ctx: AdminCtx(context.Background())},
			want: true,
		},
		{
			name: "test meta context",
			args: args{ctx: metadata.NewIncomingContext(context.Background(), metadata.New(map[string]string{
				string(TenantAdmin): "allow",
			}))},
			want: true,
		},
		{
			name: "test meta deny context",
			args: args{ctx: metadata.NewIncomingContext(context.Background(), metadata.New(map[string]string{
				string(TenantAdmin): "deny",
			}))},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetTenantAdminCtx(tt.args.ctx); got != tt.want {
				t.Errorf("GetTenantAdminCtx() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetTenantIDFromCtx(t *testing.T) {
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name string
		args args
		want uint64
	}{
		{
			name: "test tenant id",
			args: args{ctx: context.WithValue(context.Background(), enum.TenantIdCtxKey, "10")},
			want: 10,
		},
		{
			name: "test tenant default id",
			args: args{ctx: context.Background()},
			want: entenum.TenantDefaultId,
		},
		{
			name: "test meta context",
			args: args{ctx: metadata.NewIncomingContext(context.Background(), metadata.New(map[string]string{
				enum.TenantIdCtxKey: "10",
			}))},
			want: 10,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetTenantIDFromCtx(tt.args.ctx); got != tt.want {
				t.Errorf("GetTenantIDFromCtx() = %v, want %v", got, tt.want)
			}
		})
	}
}
