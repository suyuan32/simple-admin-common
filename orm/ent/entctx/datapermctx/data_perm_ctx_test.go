package datapermctx

import (
	"context"
	"reflect"
	"testing"

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
				ScopeKey: "1",
			}))},
			want:    uint8(1),
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
				CustomDeptKey: "1,3,20,8",
			}))},
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
				FilterFieldKey: "userId",
			}))},
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
