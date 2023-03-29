package mongodb

import "testing"

func TestConf_GetDSN(t *testing.T) {
	type fields struct {
		Host                  string
		Username              string
		Password              string
		Port                  int
		DBName                string
		Option                string
		AuthMechanism         string
		AuthSource            string
		TlsCAFile             string
		TlsCertificateKeyFile string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "testNone",
			fields: fields{
				Host:          "127.0.0.1",
				Port:          27017,
				AuthMechanism: "None",
			},
			want: "mongodb://127.0.0.1:27017",
		},
		{
			name: "testX509",
			fields: fields{
				Host:                  "127.0.0.1",
				Port:                  27017,
				AuthMechanism:         "MONGODB-X509",
				TlsCAFile:             "/home/ca-certificate.crt",
				TlsCertificateKeyFile: "/home/client.pem",
			},
			want: "mongodb://127.0.0.1:27017",
		},
		{
			name: "testSHA256",
			fields: fields{
				Host:          "127.0.0.1",
				Port:          27017,
				AuthMechanism: "SCRAM-SHA-256",
				DBName:        "school",
				Username:      "root",
				Password:      "simple_admin",
			},
			want: "mongodb://127.0.0.1:27017",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := Conf{
				Host:                  tt.fields.Host,
				Username:              tt.fields.Username,
				Password:              tt.fields.Password,
				Port:                  tt.fields.Port,
				DBName:                tt.fields.DBName,
				Option:                tt.fields.Option,
				AuthMechanism:         tt.fields.AuthMechanism,
				AuthSource:            tt.fields.AuthSource,
				TlsCAFile:             tt.fields.TlsCAFile,
				TlsCertificateKeyFile: tt.fields.TlsCertificateKeyFile,
			}
			if got := c.GetDSN(); got != tt.want {
				t.Errorf("GetDSN() = %v, want %v", got, tt.want)
			}
		})
	}
}
