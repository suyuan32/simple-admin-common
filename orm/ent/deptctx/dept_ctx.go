package deptctx

import (
	"context"
	"encoding/json"
	"github.com/zeromicro/go-zero/core/errorx"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest/enum"
	"google.golang.org/grpc/metadata"
	"strconv"
)

const DEPARTMENT_ADMIN = "department-admin"

// GetDepartmentIDFromCtx returns department id from context.
func GetDepartmentIDFromCtx(ctx context.Context) (uint64, error) {
	var departmentId string

	if deptId, ok := ctx.Value("deptId").(json.Number); !ok {
		if md, ok := metadata.FromIncomingContext(ctx); !ok {
			logx.Error("failed to get department id from context", logx.Field("detail", ctx))
			return 0, errorx.NewInvalidArgumentError("failed to get department ID")
		} else {
			if data := md.Get(enum.DEPARTMENT_ID_RPC_CTX_KEY); len(data) > 0 {
				departmentId = data[0]
			} else {
				return 0, errorx.NewInvalidArgumentError("failed to get department ID")
			}
		}
	} else {
		departmentId = deptId.String()
	}

	id, err := strconv.Atoi(departmentId)
	if err != nil {
		logx.Error("failed to convert department id", logx.Field("detail", err))
		return 0, errorx.NewInvalidArgumentError("failed to get department ID")
	}
	return uint64(id), nil
}

// GetDepartmentAdminCtx returns true when context including admin authority info.
// If it returns true, the operation can be executed without department privacy layer.
func GetDepartmentAdminCtx(ctx context.Context) bool {
	var policy string
	var ok bool

	if policy, ok = ctx.Value(DEPARTMENT_ADMIN).(string); ok {
		if policy == "allow" {
			return true
		}
	}

	return false
}

// AdminCtx returns a context with admin authority info.
func AdminCtx(ctx context.Context) context.Context {
	return context.WithValue(ctx, DEPARTMENT_ADMIN, "allow")
}
