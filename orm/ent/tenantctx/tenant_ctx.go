package tenantctx

import (
	"context"
	"github.com/suyuan32/simple-admin-common/orm/ent/entenum"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest/enum"
	"google.golang.org/grpc/metadata"
	"strconv"
)

// GetTenantIDFromCtx returns tenant id from context.
// If error occurs, return 0.
func GetTenantIDFromCtx(ctx context.Context) uint64 {
	var tenantId string
	var ok bool

	if tenantId, ok = ctx.Value(enum.TENANT_ID_CTX_KEY).(string); !ok {
		if md, ok := metadata.FromIncomingContext(ctx); !ok {
			logx.Error("failed to get tenant id from context", logx.Field("detail", ctx))
			return entenum.TENANT_DEFAULT_ID
		} else {
			if data := md.Get(enum.TENANT_ID_CTX_KEY); len(data) > 0 {
				tenantId = data[0]
			} else {
				return entenum.TENANT_DEFAULT_ID
			}
		}
	}

	id, err := strconv.Atoi(tenantId)
	if err != nil {
		logx.Error("failed to convert tenant id", logx.Field("detail", err))
		return entenum.TENANT_DEFAULT_ID
	}
	return uint64(id)
}
