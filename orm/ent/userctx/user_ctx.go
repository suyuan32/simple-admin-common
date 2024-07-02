package userctx

import (
	"context"
	"github.com/zeromicro/go-zero/core/errorx"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest/enum"
	"google.golang.org/grpc/metadata"
)

// GetUserIDFromCtx returns user id from context.
// If error occurs, return default user ID.
func GetUserIDFromCtx(ctx context.Context) (string, error) {
	if userId, ok := ctx.Value("userId").(string); !ok {
		if md, ok := metadata.FromIncomingContext(ctx); !ok {
			logx.Error("failed to get user id from context", logx.Field("detail", ctx))
			return "", errorx.NewInvalidArgumentError("failed to get user id from context")
		} else {
			if data := md.Get(enum.USER_ID_RPC_CTX_KEY); len(data) > 0 {
				userId = data[0]
				return userId, nil
			} else {
				return "", errorx.NewInvalidArgumentError("failed to get user id from context")
			}
		}
	} else {
		return userId, nil
	}
}
