package jwt

import "testing"

func TestNewJwtToken(t *testing.T) {
	type args struct {
		secretKey string
		iat       int64
		seconds   int64
		opt       []Option
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "test",
			args: args{
				secretKey: "jS6VKDtsJf3z1n2VKDtsJf3z1n2",
				iat:       1000,
				seconds:   10,
				opt: []Option{
					WithOption("userId", "abc"),
					WithOption("roleId", 1),
				},
			},
			want:    "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjEwMTAsImlhdCI6MTAwMCwicm9sZUlkIjoxLCJ1c2VySWQiOiJhYmMifQ.wWLNg-FLr0d6encPe0p8Dw17orN89oLK_KAV0VDVBLk",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewJwtToken(tt.args.secretKey, tt.args.iat, tt.args.seconds, tt.args.opt...)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewJwtToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("NewJwtToken() got = %v, want %v", got, tt.want)
			}
		})
	}
}
