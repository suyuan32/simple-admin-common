package deptctx

import (
	"context"
	"reflect"
	"testing"
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
			name: "test department admin context",
			args: args{ctx: context.Background()},
			want: context.WithValue(context.Background(), DEPARTMENT_ADMIN, "allow"),
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

func TestGetDepartmentAdminCtx(t *testing.T) {
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "test department admin context",
			args: args{ctx: context.WithValue(context.Background(), DEPARTMENT_ADMIN, "allow")},
			want: true,
		},
		{
			name: "test department admin wrong context",
			args: args{ctx: context.WithValue(context.Background(), DEPARTMENT_ADMIN, "allowing")},
			want: false,
		},
		{
			name: "test department empty context",
			args: args{ctx: context.Background()},
			want: false,
		},
		{
			name: "test department admin context function",
			args: args{ctx: AdminCtx(context.Background())},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetDepartmentAdminCtx(tt.args.ctx); got != tt.want {
				t.Errorf("GetDepartmentAdminCtx() = %v, want %v", got, tt.want)
			}
		})
	}
}

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
