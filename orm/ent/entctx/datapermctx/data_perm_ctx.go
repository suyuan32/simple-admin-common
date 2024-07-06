package datapermctx

import (
	"context"
	"strconv"
	"strings"

	"github.com/zeromicro/go-zero/core/errorx"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/metadata"
)

const (
	// ScopeKey is the key to store data scope
	ScopeKey = "data-perm-scope"

	// CustomDeptKey is the key to store custom department ids
	CustomDeptKey = "data-perm-custom-dept"

	// FilterFieldKey is the key to store filter field
	FilterFieldKey = "data-perm-filter-field"
)

// WithScopeContext returns context with data scope
func WithScopeContext(ctx context.Context, scope string) context.Context {
	ctx = metadata.AppendToOutgoingContext(ctx, ScopeKey, scope)
	return ctx
}

// GetScopeFromCtx returns data scope from context.
func GetScopeFromCtx(ctx context.Context) (uint8, error) {
	if md, ok := metadata.FromIncomingContext(ctx); !ok {
		logx.Error("failed to get data scope from context", logx.Field("detail", ctx))
		return 0, errorx.NewInvalidArgumentError("failed to get data scope")
	} else {
		if data := md.Get(ScopeKey); len(data) > 0 {
			scope := data[0]

			id, err := strconv.Atoi(scope)
			if err != nil {
				logx.Error("failed to convert data scope", logx.Field("detail", err))
				return 0, errorx.NewInvalidArgumentError("failed to get data scope")
			}
			return uint8(id), nil
		} else {
			return 0, errorx.NewInvalidArgumentError("failed to get data scope")
		}
	}
}

// WithCustomDeptContext returns context with custom department ids
func WithCustomDeptContext(ctx context.Context, deptIds string) context.Context {
	ctx = metadata.AppendToOutgoingContext(ctx, CustomDeptKey, deptIds)
	return ctx
}

// GetCustomDeptFromCtx returns custom department ids from context
func GetCustomDeptFromCtx(ctx context.Context) ([]uint64, error) {
	var customDeptIds []uint64

	if md, ok := metadata.FromIncomingContext(ctx); !ok {
		logx.Error("failed to get custom departmrnt ids from context", logx.Field("detail", ctx))
		return nil, errorx.NewInvalidArgumentError("failed to get custom departmrnt ids")
	} else {
		if data := md.Get(CustomDeptKey); len(data) > 0 {
			customDept := data[0]

			for _, v := range strings.Split(customDept, ",") {
				id, err := strconv.Atoi(v)
				if err != nil {
					logx.Error("failed to convert custom departmrnt ids", logx.Field("detail", err), logx.Field("data", v))
					return nil, errorx.NewInvalidArgumentError("failed to get custom departmrnt ids")
				}
				customDeptIds = append(customDeptIds, uint64(id))
			}

			return customDeptIds, nil
		} else {
			return nil, errorx.NewInvalidArgumentError("failed to get custom departmrnt ids")
		}
	}
}

// WithFilterFieldContext returns context with filter field
func WithFilterFieldContext(ctx context.Context, filterField string) context.Context {
	ctx = metadata.AppendToOutgoingContext(ctx, FilterFieldKey, filterField)
	return ctx
}

// GetFilterFieldFromCtx returns filter field from context
func GetFilterFieldFromCtx(ctx context.Context) (string, error) {
	if md, ok := metadata.FromIncomingContext(ctx); !ok {
		logx.Error("failed to get filter field from context", logx.Field("detail", ctx))
		return "", errorx.NewInvalidArgumentError("failed to get filter field")
	} else {
		if data := md.Get(FilterFieldKey); len(data) > 0 {
			return data[0], nil
		} else {
			return "", errorx.NewInvalidArgumentError("failed to get filter field")
		}
	}
}
