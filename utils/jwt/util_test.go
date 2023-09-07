package jwt

import "testing"

func TestStripBearerPrefixFromToken(t *testing.T) {
	type args struct {
		token string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "test1",
			args: args{token: "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2OTQyMjExMzksImlhdCI6MTY5Mzk2MTkzOSwicm9sZUlkIjoiMDAxIiwidXNlcklkIjoiMDE4YTY3ZmYtODgxOS03NmI4LWE5MzAtMWIxMzRjZjFjMWFmIn0.wURcTLPsO1EO-ok_-Uv2o0t3qq6iEDWbc7v7WmL--HQ"},
			want: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2OTQyMjExMzksImlhdCI6MTY5Mzk2MTkzOSwicm9sZUlkIjoiMDAxIiwidXNlcklkIjoiMDE4YTY3ZmYtODgxOS03NmI4LWE5MzAtMWIxMzRjZjFjMWFmIn0.wURcTLPsO1EO-ok_-Uv2o0t3qq6iEDWbc7v7WmL--HQ",
		},
		{
			name: "test1",
			args: args{token: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2OTQyMjExMzksImlhdCI6MTY5Mzk2MTkzOSwicm9sZUlkIjoiMDAxIiwidXNlcklkIjoiMDE4YTY3ZmYtODgxOS03NmI4LWE5MzAtMWIxMzRjZjFjMWFmIn0.wURcTLPsO1EO-ok_-Uv2o0t3qq6iEDWbc7v7WmL--HQ"},
			want: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2OTQyMjExMzksImlhdCI6MTY5Mzk2MTkzOSwicm9sZUlkIjoiMDAxIiwidXNlcklkIjoiMDE4YTY3ZmYtODgxOS03NmI4LWE5MzAtMWIxMzRjZjFjMWFmIn0.wURcTLPsO1EO-ok_-Uv2o0t3qq6iEDWbc7v7WmL--HQ",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := StripBearerPrefixFromToken(tt.args.token); got != tt.want {
				t.Errorf("StripBearerPrefixFromToken() = %v, want %v", got, tt.want)
			}
		})
	}
}
