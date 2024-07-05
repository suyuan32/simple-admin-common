package rolectx

import (
	"context"
	"slices"
	"testing"
)

func TestGetRoleIDFromCtx(t *testing.T) {
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		args    args
		want    []string
		wantErr bool
	}{
		{
			name:    "test role empty ctx",
			args:    args{ctx: context.Background()},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "test role ctx",
			args:    args{ctx: context.WithValue(context.Background(), "roleId", "001,002")},
			want:    []string{"001", "002"},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetRoleIDFromCtx(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetRoleIDFromCtx() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !slices.Equal(got, tt.want) {
				t.Errorf("GetRoleIDFromCtx() got = %v, want %v", got, tt.want)
			}
		})
	}
}
