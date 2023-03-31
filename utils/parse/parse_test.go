package parse

import (
	"reflect"
	"testing"

	"golang.org/x/text/language"
)

func TestParseTags(t *testing.T) {
	type args struct {
		lang string
	}
	tests := []struct {
		name string
		args args
		want []language.Tag
	}{
		{
			name: "testChinese",
			args: args{lang: "zh"},
			want: []language.Tag{language.Chinese},
		},
		{
			name: "testEn",
			args: args{lang: "en"},
			want: []language.Tag{language.English},
		},
		{
			name: "testWrong",
			args: args{lang: "one two"},
			want: []language.Tag{language.Chinese},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ParseTags(tt.args.lang); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseTags() = %v, want %v", got, tt.want)
			}
		})
	}
}
