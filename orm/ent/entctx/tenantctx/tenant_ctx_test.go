package tenantctx

import (
	"context"
	"reflect"
	"testing"

	"github.com/suyuan32/simple-admin-common/orm/ent/entenum"
	"github.com/zeromicro/go-zero/rest/enum"
	"google.golang.org/grpc/metadata"
)

func TestAdminCtx(t *testing.T) {
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name string
		args args
		want context.Context
	}{
		{
			name: "test tenant admin context",
			args: args{ctx: context.Background()},
			want: context.WithValue(context.Background(), TenantAdmin, "allow"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := AdminCtx(tt.args.ctx); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AdminCtx() = %v, want %v", got, tt.want)
			}
		})
	}
}

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
