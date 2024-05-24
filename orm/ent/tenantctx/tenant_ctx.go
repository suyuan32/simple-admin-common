package tenantctx

import (
	"context"
	"errors"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest/enum"
	"google.golang.org/grpc/metadata"
	"strconv"
)

// GetTenantIDFromCtx returns tenant id from context
func GetTenantIDFromCtx(ctx context.Context) (int, error) {
	var tenantId string
	var ok bool

	if tenantId, ok = ctx.Value(enum.TENANT_ID_CTX_KEY).(string); !ok {
		if md, ok := metadata.FromIncomingContext(ctx); !ok {
			return 0, errors.New("failed to get tenant id from context")
		} else {
			tenantId = md.Get(enum.TENANT_ID_CTX_KEY)[0]
		}
	}

	id, err := strconv.Atoi(tenantId)
	if err != nil {
		logx.Error("failed to convert tenant id", logx.Field("detail", id))
		return 0, err
	}
	return id, err
}
