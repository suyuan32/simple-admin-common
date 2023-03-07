package uuidx

import (
	"reflect"
	"testing"

	"github.com/gofrs/uuid/v5"
)

func TestParseUUIDSlice(t *testing.T) {
	type args struct {
		ids []string
	}
	tests := []struct {
		name string
		args args
		want []uuid.UUID
	}{
		{
			name: "test1",
			args: args{ids: []string{"123"}},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ParseUUIDSlice(tt.args.ids); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseUUIDSlice() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseUUIDString(t *testing.T) {
	type args struct {
		id string
	}
	tests := []struct {
		name string
		args args
		want uuid.UUID
	}{
		{
			name: "test1",
			args: args{id: "123456"},
			want: uuid.UUID{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ParseUUIDString(tt.args.id); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseUUIDString() = %v, want %v", got, tt.want)
			}
		})
	}
}
