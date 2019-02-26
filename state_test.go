package main

import (
	"os"
	"testing"
)

func Test_state_validate(t *testing.T) {
	type fields struct {
		Metadata     map[string]string
		Certificates map[string]string
		Settings     config
		Namespaces   map[string]namespace
		HelmRepos    map[string]string
		Apps         map[string]*release
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{
			name: "test case 1",
			fields: fields{
				Metadata: make(map[string]string),
				Certificates: map[string]string{
					"caCrt": "s3://some-bucket/12345.crt",
					"caKey": "s3://some-bucket/12345.key",
				},
				Settings: config{
					KubeContext: "minikube",
					Username:    "admin",
					Password:    "$K8S_PASSWORD",
					ClusterURI:  "https://192.168.99.100:8443",
				},
				Namespaces: map[string]namespace{
					"staging": namespace{false, false, false, "", "", "", "", "", "", "", (limits{}), make(map[string]string), make(map[string]string)},
				},
				HelmRepos: map[string]string{
					"stable": "https://kubernetes-charts.storage.googleapis.com",
					"myrepo": "s3://my-repo/charts",
				},
				Apps: make(map[string]*release),
			},
			want: true,
		}, {
			name: "test case 2 -- settings/nil_value is allowed",
			fields: fields{
				Metadata: make(map[string]string),
				Certificates: map[string]string{
					"caCrt": "s3://some-bucket/12345.crt",
					"caKey": "s3://some-bucket/12345.key",
				},
				Settings: config{},
				Namespaces: map[string]namespace{
					"staging": namespace{false, false, false, "", "", "", "", "", "", "", (limits{}), make(map[string]string), make(map[string]string)},
				},
				HelmRepos: map[string]string{
					"stable": "https://kubernetes-charts.storage.googleapis.com",
					"myrepo": "s3://my-repo/charts",
				},
				Apps: make(map[string]*release),
			},
			want: true,
		}, {
			name: "test case 3 -- settings/empty_context",
			fields: fields{
				Metadata: make(map[string]string),
				Certificates: map[string]string{
					"caCrt": "s3://some-bucket/12345.crt",
					"caKey": "s3://some-bucket/12345.key",
				},
				Settings: config{
					KubeContext: "",
					Username:    "admin",
					Password:    "$K8S_PASSWORD",
					ClusterURI:  "https://192.168.99.100:8443",
				},
				Namespaces: map[string]namespace{
					"staging": namespace{false, false, false, "", "", "", "", "", "", "", (limits{}), make(map[string]string), make(map[string]string)},
				},
				HelmRepos: map[string]string{
					"stable": "https://kubernetes-charts.storage.googleapis.com",
					"myrepo": "s3://my-repo/charts",
				},
				Apps: make(map[string]*release),
			},
			want: false,
		}, {
			name: "test case 4 -- settings/optional_params",
			fields: fields{
				Metadata: make(map[string]string),
				Certificates: map[string]string{
					"caCrt": "s3://some-bucket/12345.crt",
					"caKey": "s3://some-bucket/12345.key",
				},
				Settings: config{
					KubeContext: "minikube",
				},
				Namespaces: map[string]namespace{
					"staging": namespace{false, false, false, "", "", "", "", "", "", "", (limits{}), make(map[string]string), make(map[string]string)},
				},
				HelmRepos: map[string]string{
					"stable": "https://kubernetes-charts.storage.googleapis.com",
					"myrepo": "s3://my-repo/charts",
				},
				Apps: make(map[string]*release),
			},
			want: true,
		}, {
			name: "test case 5 -- settings/password-passed-directly",
			fields: fields{
				Metadata: make(map[string]string),
				Certificates: map[string]string{
					"caCrt": "s3://some-bucket/12345.crt",
					"caKey": "s3://some-bucket/12345.key",
				},
				Settings: config{
					KubeContext: "minikube",
					Username:    "admin",
					Password:    "K8S_PASSWORD",
					ClusterURI:  "https://192.168.99.100:8443",
				},
				Namespaces: map[string]namespace{
					"staging": namespace{false, false, false, "", "", "", "", "", "", "", (limits{}), make(map[string]string), make(map[string]string)},
				},
				HelmRepos: map[string]string{
					"stable": "https://kubernetes-charts.storage.googleapis.com",
					"myrepo": "s3://my-repo/charts",
				},
				Apps: make(map[string]*release),
			},
			want: true,
		}, {
			name: "test case 6 -- settings/clusterURI-empty-env-var",
			fields: fields{
				Metadata: make(map[string]string),
				Certificates: map[string]string{
					"caCrt": "s3://some-bucket/12345.crt",
					"caKey": "s3://some-bucket/12345.key",
				},
				Settings: config{
					KubeContext: "minikube",
					Username:    "admin",
					Password:    "K8S_PASSWORD",
					ClusterURI:  "$URI", // unset env
				},
				Namespaces: map[string]namespace{
					"staging": namespace{false, false, false, "", "", "", "", "", "", "", (limits{}), make(map[string]string), make(map[string]string)},
				},
				HelmRepos: map[string]string{
					"stable": "https://kubernetes-charts.storage.googleapis.com",
					"myrepo": "s3://my-repo/charts",
				},
				Apps: make(map[string]*release),
			},
			want: false,
		}, {
			name: "test case 7 -- settings/clusterURI-invalid",
			fields: fields{
				Metadata: make(map[string]string),
				Certificates: map[string]string{
					"caCrt": "s3://some-bucket/12345.crt",
					"caKey": "s3://some-bucket/12345.key",
				},
				Settings: config{
					KubeContext: "minikube",
					Username:    "admin",
					Password:    "K8S_PASSWORD",
					ClusterURI:  "https//192.168.99.100:8443", // invalid url
				},
				Namespaces: map[string]namespace{
					"staging": namespace{false, false, false, "", "", "", "", "", "", "", (limits{}), make(map[string]string), make(map[string]string)},
				},
				HelmRepos: map[string]string{
					"stable": "https://kubernetes-charts.storage.googleapis.com",
					"myrepo": "s3://my-repo/charts",
				},
				Apps: make(map[string]*release),
			},
			want: false,
		}, {
			name: "test case 8 -- certifications/missing key",
			fields: fields{
				Metadata: make(map[string]string),
				Certificates: map[string]string{
					"caCrt": "s3://some-bucket/12345.crt",
				},
				Settings: config{
					KubeContext: "minikube",
					Username:    "admin",
					Password:    "$K8S_PASSWORD",
					ClusterURI:  "https://192.168.99.100:8443",
				},
				Namespaces: map[string]namespace{
					"staging": namespace{false, false, false, "", "", "", "", "", "", "", (limits{}), make(map[string]string), make(map[string]string)},
				},
				HelmRepos: map[string]string{
					"stable": "https://kubernetes-charts.storage.googleapis.com",
					"myrepo": "s3://my-repo/charts",
				},
				Apps: make(map[string]*release),
			},
			want: false,
		}, {
			name: "test case 9 -- certifications/nil_value",
			fields: fields{
				Metadata:     make(map[string]string),
				Certificates: nil,
				Settings: config{
					KubeContext: "minikube",
					Username:    "admin",
					Password:    "$K8S_PASSWORD",
					ClusterURI:  "https://192.168.99.100:8443",
				},
				Namespaces: map[string]namespace{
					"staging": namespace{false, false, false, "", "", "", "", "", "", "", (limits{}), make(map[string]string), make(map[string]string)},
				},
				HelmRepos: map[string]string{
					"stable": "https://kubernetes-charts.storage.googleapis.com",
					"myrepo": "s3://my-repo/charts",
				},
				Apps: make(map[string]*release),
			},
			want: false,
		}, {
			name: "test case 10 -- certifications/invalid_s3",
			fields: fields{
				Metadata: make(map[string]string),
				Certificates: map[string]string{
					"caCrt": "s3://some-bucket/12345.crt",
					"caKey": "http://someurl.com/",
				},
				Settings: config{
					KubeContext: "minikube",
					Username:    "admin",
					Password:    "$K8S_PASSWORD",
					ClusterURI:  "https://192.168.99.100:8443",
				},
				Namespaces: map[string]namespace{
					"staging": namespace{false, false, false, "", "", "", "", "", "", "", (limits{}), make(map[string]string), make(map[string]string)},
				},
				HelmRepos: map[string]string{
					"stable": "https://kubernetes-charts.storage.googleapis.com",
					"myrepo": "s3://my-repo/charts",
				},
				Apps: make(map[string]*release),
			},
			want: false,
		}, {
			name: "test case 11 -- certifications/nil_value_pass",
			fields: fields{
				Metadata:     make(map[string]string),
				Certificates: nil,
				Settings: config{
					KubeContext: "minikube",
				},
				Namespaces: map[string]namespace{
					"staging": namespace{false, false, false, "", "", "", "", "", "", "", (limits{}), make(map[string]string), make(map[string]string)},
				},
				HelmRepos: map[string]string{
					"stable": "https://kubernetes-charts.storage.googleapis.com",
					"myrepo": "s3://my-repo/charts",
				},
				Apps: make(map[string]*release),
			},
			want: true,
		}, {
			name: "test case 12 -- namespaces/nil_value",
			fields: fields{
				Metadata:     make(map[string]string),
				Certificates: nil,
				Settings: config{
					KubeContext: "minikube",
				},
				Namespaces: nil,
				HelmRepos: map[string]string{
					"stable": "https://kubernetes-charts.storage.googleapis.com",
					"myrepo": "s3://my-repo/charts",
				},
				Apps: make(map[string]*release),
			},
			want: false,
		}, {
			name: "test case 13 -- namespaces/empty",
			fields: fields{
				Metadata:     make(map[string]string),
				Certificates: nil,
				Settings: config{
					KubeContext: "minikube",
				},
				Namespaces: map[string]namespace{},
				HelmRepos: map[string]string{
					"stable": "https://kubernetes-charts.storage.googleapis.com",
					"myrepo": "s3://my-repo/charts",
				},
				Apps: make(map[string]*release),
			},
			want: false,
		}, {
			name: "test case 14 -- namespaces/use and install tiller",
			fields: fields{
				Metadata:     make(map[string]string),
				Certificates: nil,
				Settings: config{
					KubeContext: "minikube",
				},
				Namespaces: map[string]namespace{
					"staging": namespace{false, true, true, "", "", "", "", "", "", "", (limits{}), make(map[string]string), make(map[string]string)},
				},
				HelmRepos: map[string]string{
					"stable": "https://kubernetes-charts.storage.googleapis.com",
					"myrepo": "s3://my-repo/charts",
				},
				Apps: make(map[string]*release),
			},
			want: false,
		}, {
			name: "test case 15 -- namespaces/use tiller with tls-valid",
			fields: fields{
				Metadata:     make(map[string]string),
				Certificates: nil,
				Settings: config{
					KubeContext: "minikube",
				},
				Namespaces: map[string]namespace{
					"staging": namespace{false, false, true, "", "", "s3://some-bucket/12345.crt", "", "", "s3://some-bucket/12345.crt", "s3://some-bucket/12345.crt", (limits{}), make(map[string]string), make(map[string]string)},
				},
				HelmRepos: map[string]string{
					"stable": "https://kubernetes-charts.storage.googleapis.com",
					"myrepo": "s3://my-repo/charts",
				},
				Apps: make(map[string]*release),
			},
			want: true,
		}, {
			name: "test case 16 -- namespaces/use tiller with tls-not enough certs",
			fields: fields{
				Metadata:     make(map[string]string),
				Certificates: nil,
				Settings: config{
					KubeContext: "minikube",
				},
				Namespaces: map[string]namespace{
					"staging": namespace{false, false, true, "", "", "", "", "", "s3://some-bucket/12345.crt", "s3://some-bucket/12345.crt", (limits{}), make(map[string]string), make(map[string]string)},
				},
				HelmRepos: map[string]string{
					"stable": "https://kubernetes-charts.storage.googleapis.com",
					"myrepo": "s3://my-repo/charts",
				},
				Apps: make(map[string]*release),
			},
			want: true,
		}, {
			name: "test case 17 -- namespaces/deploy tiller with tls- valid",
			fields: fields{
				Metadata:     make(map[string]string),
				Certificates: nil,
				Settings: config{
					KubeContext: "minikube",
				},
				Namespaces: map[string]namespace{
					"staging": namespace{false, true, false, "", "", "s3://some-bucket/12345.crt", "s3://some-bucket/12345.crt", "s3://some-bucket/12345.crt", "s3://some-bucket/12345.crt", "s3://some-bucket/12345.crt", (limits{}), make(map[string]string), make(map[string]string)},
				},
				HelmRepos: map[string]string{
					"stable": "https://kubernetes-charts.storage.googleapis.com",
					"myrepo": "s3://my-repo/charts",
				},
				Apps: make(map[string]*release),
			},
			want: true,
		}, {
			name: "test case 18 -- helmRepos/nil_value",
			fields: fields{
				Metadata:     make(map[string]string),
				Certificates: nil,
				Settings: config{
					KubeContext: "minikube",
				},
				Namespaces: map[string]namespace{
					"staging": namespace{false, false, false, "", "", "", "", "", "", "", (limits{}), make(map[string]string), make(map[string]string)},
				},
				HelmRepos: nil,
				Apps:      make(map[string]*release),
			},
			want: false,
		}, {
			name: "test case 19 -- helmRepos/empty",
			fields: fields{
				Metadata:     make(map[string]string),
				Certificates: nil,
				Settings: config{
					KubeContext: "minikube",
				},
				Namespaces: map[string]namespace{
					"staging": namespace{false, false, false, "", "", "", "", "", "", "", (limits{}), make(map[string]string), make(map[string]string)},
				},
				HelmRepos: map[string]string{},
				Apps:      make(map[string]*release),
			},
			want: false,
		}, {
			name: "test case 20 -- helmRepos/empty_repo_value",
			fields: fields{
				Metadata:     make(map[string]string),
				Certificates: nil,
				Settings: config{
					KubeContext: "minikube",
				},
				Namespaces: map[string]namespace{
					"staging": namespace{false, false, false, "", "", "", "", "", "", "", (limits{}), make(map[string]string), make(map[string]string)},
				},
				HelmRepos: map[string]string{
					"stable": "https://kubernetes-charts.storage.googleapis.com",
					"myrepo": "",
				},
				Apps: make(map[string]*release),
			},
			want: false,
		}, {
			name: "test case 21 -- helmRepos/invalid_repo_value",
			fields: fields{
				Metadata:     make(map[string]string),
				Certificates: nil,
				Settings: config{
					KubeContext: "minikube",
				},
				Namespaces: map[string]namespace{
					"staging": namespace{false, false, false, "", "", "", "", "", "", "", (limits{}), make(map[string]string), make(map[string]string)},
				},
				HelmRepos: map[string]string{
					"stable": "https://kubernetes-charts.storage.googleapis.com",
					"myrepo": "s3//my-repo/charts",
				},
				Apps: make(map[string]*release),
			},
			want: false,
		},
	}
	os.Setenv("K8S_PASSWORD", "my-fake-password")
	os.Setenv("SET_URI", "https://192.168.99.100:8443")
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := state{
				Metadata:     tt.fields.Metadata,
				Certificates: tt.fields.Certificates,
				Settings:     tt.fields.Settings,
				Namespaces:   tt.fields.Namespaces,
				HelmRepos:    tt.fields.HelmRepos,
				Apps:         tt.fields.Apps,
			}
			if got, _ := s.validate(); got != tt.want {
				t.Errorf("state.validate() = %v, want %v", got, tt.want)
			}
		})
	}
}
