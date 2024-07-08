package deptctx

import (
	"context"
	"testing"
)

func TestGetDepartmentIDFromCtx(t *testing.T) {
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		args    args
		want    uint64
		wantErr bool
	}{
		{
			name:    "test department default id",
			args:    args{ctx: context.Background()},
			want:    0,
			wantErr: true,
		},
		{
			name:    "test department id",
			args:    args{ctx: context.WithValue(context.Background(), "deptId", "")},
			want:    0,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetDepartmentIDFromCtx(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetDepartmentIDFromCtx() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetDepartmentIDFromCtx() got = %v, want %v", got, tt.want)
			}
		})
	}
}
