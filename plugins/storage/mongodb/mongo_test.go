// Copyright 2023 The Ryan SU Authors (https://github.com/suyuan32). All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
