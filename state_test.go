package main

import "testing"

func Test_state_validate(t *testing.T) {
	type fields struct {
		Metadata     map[string]string
		Certificates map[string]string
		Settings     map[string]string
		Namespaces   map[string]string
		HelmRepos    map[string]string
		Apps         map[string]release
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
				Settings: map[string]string{
					"kubeContext": "minikube",
					"username":    "admin",
					"password":    "$K8S_PASSWORD",
					"clusterURI":  "https://192.168.99.100:8443",
				},
				Namespaces: map[string]string{
					"staging": "staging",
				},
				HelmRepos: map[string]string{
					"stable": "https://kubernetes-charts.storage.googleapis.com",
					"myrepo": "s3://my-repo/charts",
				},
				Apps: make(map[string]release),
			},
			want: true,
		}, {
			name: "test case 2 -- settings/nil_value",
			fields: fields{
				Metadata: make(map[string]string),
				Certificates: map[string]string{
					"caCrt": "s3://some-bucket/12345.crt",
					"caKey": "s3://some-bucket/12345.key",
				},
				Settings: nil,
				Namespaces: map[string]string{
					"staging": "staging",
				},
				HelmRepos: map[string]string{
					"stable": "https://kubernetes-charts.storage.googleapis.com",
					"myrepo": "s3://my-repo/charts",
				},
				Apps: make(map[string]release),
			},
			want: false,
		}, {
			name: "test case 3 -- settings/empty_context",
			fields: fields{
				Metadata: make(map[string]string),
				Certificates: map[string]string{
					"caCrt": "s3://some-bucket/12345.crt",
					"caKey": "s3://some-bucket/12345.key",
				},
				Settings: map[string]string{
					"kubeContext": "",
					"username":    "admin",
					"password":    "$K8S_PASSWORD",
					"clusterURI":  "https://192.168.99.100:8443",
				},
				Namespaces: map[string]string{
					"staging": "staging",
				},
				HelmRepos: map[string]string{
					"stable": "https://kubernetes-charts.storage.googleapis.com",
					"myrepo": "s3://my-repo/charts",
				},
				Apps: make(map[string]release),
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
				Settings: map[string]string{
					"kubeContext": "minikube",
				},
				Namespaces: map[string]string{
					"staging": "staging",
				},
				HelmRepos: map[string]string{
					"stable": "https://kubernetes-charts.storage.googleapis.com",
					"myrepo": "s3://my-repo/charts",
				},
				Apps: make(map[string]release),
			},
			want: true,
		}, {
			name: "test case 5 -- settings/password",
			fields: fields{
				Metadata: make(map[string]string),
				Certificates: map[string]string{
					"caCrt": "s3://some-bucket/12345.crt",
					"caKey": "s3://some-bucket/12345.key",
				},
				Settings: map[string]string{
					"kubeContext": "minikube",
					"username":    "admin",
					"password":    "K8S_PASSWORD",
					"clusterURI":  "https://192.168.99.100:8443",
				},
				Namespaces: map[string]string{
					"staging": "staging",
				},
				HelmRepos: map[string]string{
					"stable": "https://kubernetes-charts.storage.googleapis.com",
					"myrepo": "s3://my-repo/charts",
				},
				Apps: make(map[string]release),
			},
			want: false,
		}, {
			name: "test case 6 -- settings/clusterURI",
			fields: fields{
				Metadata: make(map[string]string),
				Certificates: map[string]string{
					"caCrt": "s3://some-bucket/12345.crt",
					"caKey": "s3://some-bucket/12345.key",
				},
				Settings: map[string]string{
					"kubeContext": "minikube",
					"username":    "admin",
					"password":    "K8S_PASSWORD",
					"clusterURI":  "https//192.168.99.100:8443",
				},
				Namespaces: map[string]string{
					"staging": "staging",
				},
				HelmRepos: map[string]string{
					"stable": "https://kubernetes-charts.storage.googleapis.com",
					"myrepo": "s3://my-repo/charts",
				},
				Apps: make(map[string]release),
			},
			want: false,
		}, {
			name: "test case 7 -- certifications/missing key",
			fields: fields{
				Metadata: make(map[string]string),
				Certificates: map[string]string{
					"caCrt": "s3://some-bucket/12345.crt",
				},
				Settings: map[string]string{
					"kubeContext": "minikube",
					"username":    "admin",
					"password":    "$K8S_PASSWORD",
					"clusterURI":  "https://192.168.99.100:8443",
				},
				Namespaces: map[string]string{
					"staging": "staging",
				},
				HelmRepos: map[string]string{
					"stable": "https://kubernetes-charts.storage.googleapis.com",
					"myrepo": "s3://my-repo/charts",
				},
				Apps: make(map[string]release),
			},
			want: false,
		}, {
			name: "test case 8 -- certifications/nil_value",
			fields: fields{
				Metadata:     make(map[string]string),
				Certificates: nil,
				Settings: map[string]string{
					"kubeContext": "minikube",
					"username":    "admin",
					"password":    "$K8S_PASSWORD",
					"clusterURI":  "https://192.168.99.100:8443",
				},
				Namespaces: map[string]string{
					"staging": "staging",
				},
				HelmRepos: map[string]string{
					"stable": "https://kubernetes-charts.storage.googleapis.com",
					"myrepo": "s3://my-repo/charts",
				},
				Apps: make(map[string]release),
			},
			want: false,
		}, {
			name: "test case 9 -- certifications/invalid_s3",
			fields: fields{
				Metadata: make(map[string]string),
				Certificates: map[string]string{
					"caCrt": "s3://some-bucket/12345.crt",
					"caKey": "http://someurl.com/",
				},
				Settings: map[string]string{
					"kubeContext": "minikube",
					"username":    "admin",
					"password":    "$K8S_PASSWORD",
					"clusterURI":  "https://192.168.99.100:8443",
				},
				Namespaces: map[string]string{
					"staging": "staging",
				},
				HelmRepos: map[string]string{
					"stable": "https://kubernetes-charts.storage.googleapis.com",
					"myrepo": "s3://my-repo/charts",
				},
				Apps: make(map[string]release),
			},
			want: false,
		}, {
			name: "test case 10 -- certifications/nil_value_pass",
			fields: fields{
				Metadata:     make(map[string]string),
				Certificates: nil,
				Settings: map[string]string{
					"kubeContext": "minikube",
				},
				Namespaces: map[string]string{
					"staging": "staging",
				},
				HelmRepos: map[string]string{
					"stable": "https://kubernetes-charts.storage.googleapis.com",
					"myrepo": "s3://my-repo/charts",
				},
				Apps: make(map[string]release),
			},
			want: true,
		}, {
			name: "test case 11 -- namespaces/nil_value",
			fields: fields{
				Metadata:     make(map[string]string),
				Certificates: nil,
				Settings: map[string]string{
					"kubeContext": "minikube",
				},
				Namespaces: nil,
				HelmRepos: map[string]string{
					"stable": "https://kubernetes-charts.storage.googleapis.com",
					"myrepo": "s3://my-repo/charts",
				},
				Apps: make(map[string]release),
			},
			want: false,
		}, {
			name: "test case 12 -- namespaces/empty",
			fields: fields{
				Metadata:     make(map[string]string),
				Certificates: nil,
				Settings: map[string]string{
					"kubeContext": "minikube",
				},
				Namespaces: map[string]string{},
				HelmRepos: map[string]string{
					"stable": "https://kubernetes-charts.storage.googleapis.com",
					"myrepo": "s3://my-repo/charts",
				},
				Apps: make(map[string]release),
			},
			want: false,
		}, {
			name: "test case 13 -- namespaces/empty_namespace_value",
			fields: fields{
				Metadata:     make(map[string]string),
				Certificates: nil,
				Settings: map[string]string{
					"kubeContext": "minikube",
				},
				Namespaces: map[string]string{
					"staging": "staging",
					"x":       "y",
					"z":       "",
				},
				HelmRepos: map[string]string{
					"stable": "https://kubernetes-charts.storage.googleapis.com",
					"myrepo": "s3://my-repo/charts",
				},
				Apps: make(map[string]release),
			},
			want: false,
		}, {
			name: "test case 14 -- helmRepos/nil_value",
			fields: fields{
				Metadata:     make(map[string]string),
				Certificates: nil,
				Settings: map[string]string{
					"kubeContext": "minikube",
				},
				Namespaces: map[string]string{
					"staging": "staging",
				},
				HelmRepos: nil,
				Apps:      make(map[string]release),
			},
			want: false,
		}, {
			name: "test case 15 -- helmRepos/empty",
			fields: fields{
				Metadata:     make(map[string]string),
				Certificates: nil,
				Settings: map[string]string{
					"kubeContext": "minikube",
				},
				Namespaces: map[string]string{
					"staging": "staging",
				},
				HelmRepos: map[string]string{},
				Apps:      make(map[string]release),
			},
			want: false,
		}, {
			name: "test case 16 -- helmRepos/empty_repo_value",
			fields: fields{
				Metadata:     make(map[string]string),
				Certificates: nil,
				Settings: map[string]string{
					"kubeContext": "minikube",
				},
				Namespaces: map[string]string{
					"staging": "staging",
				},
				HelmRepos: map[string]string{
					"stable": "https://kubernetes-charts.storage.googleapis.com",
					"myrepo": "",
				},
				Apps: make(map[string]release),
			},
			want: false,
		}, {
			name: "test case 17 -- helmRepos/invalid_repo_value",
			fields: fields{
				Metadata:     make(map[string]string),
				Certificates: nil,
				Settings: map[string]string{
					"kubeContext": "minikube",
				},
				Namespaces: map[string]string{
					"staging": "staging",
				},
				HelmRepos: map[string]string{
					"stable": "https://kubernetes-charts.storage.googleapis.com",
					"myrepo": "s3//my-repo/charts",
				},
				Apps: make(map[string]release),
			},
			want: false,
		},
	}
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

// func Test_state_print(t *testing.T) {
// 	type fields struct {
// 		Metadata       map[string]string
// 		Certifications map[string]string
// 		Settings       map[string]string
// 		Namespaces     map[string]string
// 		HelmRepos      map[string]string
// 		Apps           map[string]release
// 	}
// 	tests := []struct {
// 		name   string
// 		fields fields
// 	}{
// 	// TODO: Add test cases.
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			s := state{
// 				Metadata:       tt.fields.Metadata,
// 				Certifications: tt.fields.Certifications,
// 				Settings:       tt.fields.Settings,
// 				Namespaces:     tt.fields.Namespaces,
// 				HelmRepos:      tt.fields.HelmRepos,
// 				Apps:           tt.fields.Apps,
// 			}
// 			s.print()
// 		})
// 	}
// }
