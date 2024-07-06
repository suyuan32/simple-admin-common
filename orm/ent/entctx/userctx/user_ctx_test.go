package userctx

import (
	"context"
	"testing"

	"github.com/zeromicro/go-zero/rest/enum"
	"google.golang.org/grpc/metadata"
)

func TestGetUserIDFromCtx(t *testing.T) {
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
			name:    "test user empty ctx",
			args:    args{ctx: context.Background()},
			want:    "",
			wantErr: true,
		},
		{
			name:    "test user ctx",
			args:    args{ctx: context.WithValue(context.Background(), "userId", "asdfghjkl")},
			want:    "asdfghjkl",
			wantErr: false,
		},
		{
			name: "test meta context",
			args: args{ctx: metadata.NewIncomingContext(context.Background(), metadata.New(map[string]string{
				enum.UserIdRpcCtxKey: "asdfghjkl",
			}))},
			want: "asdfghjkl",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetUserIDFromCtx(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetUserIDFromCtx() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetUserIDFromCtx() got = %v, want %v", got, tt.want)
			}
		})
	}
}
