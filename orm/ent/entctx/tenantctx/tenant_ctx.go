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
	"strconv"

	"github.com/suyuan32/simple-admin-common/orm/ent/entenum"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest/enum"
	"google.golang.org/grpc/metadata"
)

type TenantKey string

const TenantAdmin TenantKey = "tenant-admin"

// GetTenantIDFromCtx returns tenant id from context.
// If error occurs, return default tenant ID.
func GetTenantIDFromCtx(ctx context.Context) uint64 {
	var tenantId string
	var ok bool

	if tenantId, ok = ctx.Value(enum.TenantIdCtxKey).(string); !ok {
		if md, ok := metadata.FromIncomingContext(ctx); !ok {
			logx.Error("failed to get tenant id from context", logx.Field("detail", ctx))
			return entenum.TenantDefaultId
		} else {
			if data := md.Get(enum.TenantIdCtxKey); len(data) > 0 {
				tenantId = data[0]
			} else {
				return entenum.TenantDefaultId
			}
		}
	}

	id, err := strconv.Atoi(tenantId)
	if err != nil {
		logx.Error("failed to convert tenant id", logx.Field("detail", err))
		return entenum.TenantDefaultId
	}
	return uint64(id)
}

// GetTenantAdminCtx returns true when context including admin authority info.
// If it returns true, the operation can be executed without tenant privacy layer.
func GetTenantAdminCtx(ctx context.Context) bool {
	var policy string
	var ok bool

	if policy, ok = ctx.Value(TenantAdmin).(string); !ok {
		if md, ok := metadata.FromIncomingContext(ctx); !ok {
			return false
		} else {
			if data := md.Get(string(TenantAdmin)); len(data) > 0 {
				policy = data[0]
			} else {
				return false
			}
		}
	}

	if policy == "allow" {
		return true
	}

	return false
}

// AdminCtx returns a context with admin authority info.
func AdminCtx(ctx context.Context) context.Context {
	ctx = metadata.AppendToOutgoingContext(ctx, string(TenantAdmin), "allow")
	return context.WithValue(ctx, TenantAdmin, "allow")
}
