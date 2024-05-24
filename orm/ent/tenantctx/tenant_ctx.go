package tenantctx

import (
	"context"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest/enum"
	"google.golang.org/grpc/metadata"
	"strconv"
)

// GetTenantIDFromCtx returns tenant id from context.
// If error occurs, return 0.
func GetTenantIDFromCtx(ctx context.Context) int {
	var tenantId string
	var ok bool

	if tenantId, ok = ctx.Value(enum.TENANT_ID_CTX_KEY).(string); !ok {
		if md, ok := metadata.FromIncomingContext(ctx); !ok {
			logx.Error("failed to get tenant id from context", logx.Field("detail", ctx))
			return 0
		} else {
			if data := md.Get(enum.TENANT_ID_CTX_KEY); len(data) > 0 {
				tenantId = data[0]
			} else {
				return 0
			}
		}
	}

	id, err := strconv.Atoi(tenantId)
	if err != nil {
		logx.Error("failed to convert tenant id", logx.Field("detail", err))
		return 0
	}
	return id
}
