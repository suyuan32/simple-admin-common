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

	if policy, ok = ctx.Value(TenantAdmin).(string); ok {
		if policy == "allow" {
			return true
		}
	}

	return false
}

// AdminCtx returns a context with admin authority info.
func AdminCtx(ctx context.Context) context.Context {
	return context.WithValue(ctx, TenantAdmin, "allow")
}
