package rolectx

import (
	"context"
	"slices"
	"strings"

	"github.com/zeromicro/go-zero/core/errorx"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest/enum"
	"google.golang.org/grpc/metadata"
)

// GetRoleIDFromCtx returns role id from context.
func GetRoleIDFromCtx(ctx context.Context) ([]string, error) {
	if roleId, ok := ctx.Value("roleId").(string); !ok {
		if md, ok := metadata.FromIncomingContext(ctx); !ok {
			logx.Error("failed to get role id from context", logx.Field("detail", ctx))
			return nil, errorx.NewInvalidArgumentError("failed to get role id from context")
		} else {
			if data := md.Get(enum.UserIdRpcCtxKey); len(data) > 0 {
				roleIds := strings.Split(data[0], ",")
				slices.Sort(roleIds)
				return roleIds, nil
			} else {
				return nil, errorx.NewInvalidArgumentError("failed to get role id from context")
			}
		}
	} else {
		roleIds := strings.Split(roleId, ",")
		slices.Sort(roleIds)
		return roleIds, nil
	}
}
