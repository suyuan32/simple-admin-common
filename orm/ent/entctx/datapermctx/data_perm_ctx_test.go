package datapermctx

import (
	"context"
	"reflect"
	"testing"

	"github.com/suyuan32/simple-admin-common/orm/ent/entenum"
	"google.golang.org/grpc/metadata"
)

func TestGetScopeFromCtx(t *testing.T) {
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		args    args
		want    uint8
		wantErr bool
	}{
		{
			name: "test scope",
			args: args{ctx: metadata.NewIncomingContext(context.Background(), metadata.New(map[string]string{
				string(ScopeKey): entenum.DataPermAllStr,
			}))},
			want:    entenum.DataPermAll,
			wantErr: false,
		},
		{
			name:    "test nil context",
			args:    args{ctx: metadata.NewIncomingContext(context.Background(), metadata.New(map[string]string{}))},
			want:    0,
			wantErr: true,
		},
		{
			name:    "test normal context",
			args:    args{ctx: context.WithValue(context.Background(), ScopeKey, entenum.DataPermAllStr)},
			want:    entenum.DataPermAll,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetScopeFromCtx(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetScopeFromCtx() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetScopeFromCtx() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetCustomDeptFromCtx(t *testing.T) {
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		args    args
		want    []uint64
		wantErr bool
	}{
		{
			name: "test custom department",
			args: args{ctx: metadata.NewIncomingContext(context.Background(), metadata.New(map[string]string{
				string(CustomDeptKey): "1,3,20,8",
			}))},
			want:    []uint64{1, 3, 20, 8},
			wantErr: false,
		},
		{
			name:    "test nil context",
			args:    args{ctx: metadata.NewIncomingContext(context.Background(), metadata.New(map[string]string{}))},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "test normal context",
			args:    args{ctx: context.WithValue(context.Background(), CustomDeptKey, "1,3,20,8")},
			want:    []uint64{1, 3, 20, 8},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetCustomDeptFromCtx(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetCustomDeptFromCtx() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetCustomDeptFromCtx() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetSubDeptFromCtx(t *testing.T) {
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		args    args
		want    []uint64
		wantErr bool
	}{
		{
			name: "test custom department",
			args: args{ctx: metadata.NewIncomingContext(context.Background(), metadata.New(map[string]string{
				string(SubDeptKey): "1,3,20,8",
			}))},
			want:    []uint64{1, 3, 20, 8},
			wantErr: false,
		},
		{
			name:    "test nil context",
			args:    args{ctx: metadata.NewIncomingContext(context.Background(), metadata.New(map[string]string{}))},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "test normal context",
			args:    args{ctx: context.WithValue(context.Background(), SubDeptKey, "1,3,20,8")},
			want:    []uint64{1, 3, 20, 8},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetSubDeptFromCtx(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetSubDeptFromCtx() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetSubDeptFromCtx() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetFilterFieldFromCtx(t *testing.T) {
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "test filter field",
			args: args{ctx: metadata.NewIncomingContext(context.Background(), metadata.New(map[string]string{
				string(FilterFieldKey): "userId",
			}))},
			want:    "userId",
			wantErr: false,
		},
		{
			name:    "test nil context",
			args:    args{ctx: metadata.NewIncomingContext(context.Background(), metadata.New(map[string]string{}))},
			want:    "",
			wantErr: true,
		},
		{
			name:    "test normal context",
			args:    args{ctx: context.WithValue(context.Background(), FilterFieldKey, "userId")},
			want:    "userId",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetFilterFieldFromCtx(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetFilterFieldFromCtx() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetFilterFieldFromCtx() = %v, want %v", got, tt.want)
			}
		})
	}
}
