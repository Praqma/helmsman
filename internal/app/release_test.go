package app

import (
	"testing"
)

func Test_release_validate(t *testing.T) {
	st := state{
		Metadata:     make(map[string]string),
		Certificates: make(map[string]string),
		Settings:     (config{}),
		Namespaces:   map[string]*namespace{"namespace": {false, limits{}, make(map[string]string), make(map[string]string), &quotas{}, false}},
		HelmRepos:    make(map[string]string),
		Apps:         make(map[string]*release),
	}

	type args struct {
		s state
		r *release
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "test case 1",
			args: args{
				r: &release{
					Name:        "release1",
					Description: "",
					Namespace:   "namespace",
					Enabled:     true,
					Chart:       "repo/chartX",
					Version:     "1.0",
					ValuesFile:  "../../tests/values.yaml",
					Test:        true,
				},
				s: st,
			},
			want: "",
		}, {
			name: "test case 2",
			args: args{
				r: &release{
					Name:        "release2",
					Description: "",
					Namespace:   "namespace",
					Enabled:     true,
					Chart:       "repo/chartX",
					Version:     "1.0",
					ValuesFile:  "xyz.yaml",
					Test:        true,
				},
				s: st,
			},
			want: "xyz.yaml must be valid relative (from dsf file) file path",
		}, {
			name: "test case 3",
			args: args{
				r: &release{
					Name:        "release3",
					Description: "",
					Namespace:   "namespace",
					Enabled:     true,
					Chart:       "repo/chartX",
					Version:     "1.0",
					ValuesFile:  "../../tests/values.xml",
					Test:        true,
				},
				s: st,
			},
			want: "../../tests/values.xml must be of one the following file formats: .yaml, .yml, .json",
		}, {
			name: "test case 4",
			args: args{
				r: &release{
					Name:        "release1",
					Description: "",
					Namespace:   "namespace",
					Enabled:     true,
					Chart:       "repo/chartX",
					Version:     "1.0",
					ValuesFile:  "../../tests/values.yaml",
					Test:        true,
				},
				s: st,
			},
			want: "release name must be unique within a given namespace",
		}, {
			name: "test case 5",
			args: args{
				r: &release{
					Name:        "",
					Description: "",
					Namespace:   "namespace",
					Enabled:     true,
					Chart:       "repo/chartX",
					Version:     "1.0",
					ValuesFile:  "../../tests/values.yaml",
					Test:        true,
				},
				s: st,
			},
			want: "",
		}, {
			name: "test case 6",
			args: args{
				r: &release{
					Name:        "release6",
					Description: "",
					Namespace:   "",
					Enabled:     true,
					Chart:       "repo/chartX",
					Version:     "1.0",
					ValuesFile:  "../../tests/values.yaml",
					Test:        true,
				},
				s: st,
			},
			want: "release targeted namespace can't be empty",
		}, {
			name: "test case 7",
			args: args{
				r: &release{
					Name:        "release7",
					Description: "",
					Namespace:   "namespace",
					Enabled:     true,
					Chart:       "chartX",
					Version:     "1.0",
					ValuesFile:  "../../tests/values.yaml",
					Test:        true,
				},
				s: st,
			},
			want: "chart can't be empty and must be of the format: repo/chart",
		}, {
			name: "test case 8",
			args: args{
				r: &release{
					Name:        "release8",
					Description: "",
					Namespace:   "namespace",
					Enabled:     true,
					Chart:       "",
					Version:     "1.0",
					ValuesFile:  "../../tests/values.yaml",
					Test:        true,
				},
				s: st,
			},
			want: "chart can't be empty and must be of the format: repo/chart",
		}, {
			name: "test case 9",
			args: args{
				r: &release{
					Name:        "release9",
					Description: "",
					Namespace:   "namespace",
					Enabled:     true,
					Chart:       "repo/chartX",
					Version:     "",
					ValuesFile:  "../../tests/values.yaml",
					Test:        true,
				},
				s: st,
			},
			want: "version can't be empty",
		}, {
			name: "test case 10",
			args: args{
				r: &release{
					Name:        "release10",
					Description: "",
					Namespace:   "namespace",
					Enabled:     true,
					Chart:       "repo/chartX",
					Version:     "1.0",
					ValuesFile:  "../../tests/values.yaml",
					Test:        true,
				},
				s: st,
			},
			want: "",
		}, {
			name: "test case 11",
			args: args{
				r: &release{
					Name:        "release11",
					Description: "",
					Namespace:   "namespace",
					Enabled:     true,
					Chart:       "repo/chartX",
					Version:     "1.0",
					ValuesFile:  "../../tests/values.yaml",
					ValuesFiles: []string{"xyz.yaml"},
					Test:        true,
				},
				s: st,
			},
			want: "valuesFile and valuesFiles should not be used together",
		}, {
			name: "test case 12",
			args: args{
				r: &release{
					Name:        "release12",
					Description: "",
					Namespace:   "namespace",
					Enabled:     true,
					Chart:       "repo/chartX",
					Version:     "1.0",
					ValuesFiles: []string{"xyz.yaml"},
					Test:        true,
				},
				s: st,
			},
			want: "xyz.yaml must be valid relative (from dsf file) file path",
		}, {
			name: "test case 13",
			args: args{
				r: &release{
					Name:        "release13",
					Description: "",
					Namespace:   "namespace",
					Enabled:     true,
					Chart:       "repo/chartX",
					Version:     "1.0",
					ValuesFiles: []string{"./../../tests/values.yaml", "../../tests/values2.yaml"},
					Test:        true,
				},
				s: st,
			},
			want: "",
		}, {
			name: "test case 14 - non-existing hook file",
			args: args{
				r: &release{
					Name:        "release14",
					Description: "",
					Namespace:   "namespace",
					Enabled:     true,
					Chart:       "repo/chartX",
					Version:     "1.0",
					ValuesFile:  "../../tests/values.yaml",
					Hooks:       map[string]interface{}{preInstall: "xyz.fake"},
				},
				s: st,
			},
			want: "xyz.fake must be valid relative (from dsf file) file path",
		}, {
			name: "test case 15 - invalid hook file type",
			args: args{
				r: &release{
					Name:        "release15",
					Description: "",
					Namespace:   "namespace",
					Enabled:     true,
					Chart:       "repo/chartX",
					Version:     "1.0",
					ValuesFile:  "../../tests/values.yaml",
					Hooks:       map[string]interface{}{preInstall: "../../tests/values.xml"},
				},
				s: st,
			},
			want: "../../tests/values.xml must be of one the following file formats: .yaml, .yml, .json",
		}, {
			name: "test case 16 - valid hook file type",
			args: args{
				r: &release{
					Name:        "release16",
					Description: "",
					Namespace:   "namespace",
					Enabled:     true,
					Chart:       "repo/chartX",
					Version:     "1.0",
					ValuesFile:  "../../tests/values.yaml",
					Hooks:       map[string]interface{}{preDelete: "../../tests/values.yaml"},
				},
				s: st,
			},
			want: "",
		}, {
			name: "test case 17 - valid hook file URL",
			args: args{
				r: &release{
					Name:        "release17",
					Description: "",
					Namespace:   "namespace",
					Enabled:     true,
					Chart:       "repo/chartX",
					Version:     "1.0",
					ValuesFile:  "../../tests/values.yaml",
					Hooks:       map[string]interface{}{postUpgrade: "https://raw.githubusercontent.com/jetstack/cert-manager/release-0.14/deploy/manifests/00-crds.yaml"},
				},
				s: st,
			},
			want: "",
		}, {
			name: "test case 18 - invalid hook file URL",
			args: args{
				r: &release{
					Name:        "release18",
					Description: "",
					Namespace:   "namespace",
					Enabled:     true,
					Chart:       "repo/chartX",
					Version:     "1.0",
					ValuesFile:  "../../tests/values.yaml",
					Hooks:       map[string]interface{}{preDelete: "https//raw.githubusercontent.com/jetstack/cert-manager/release-0.14/deploy/manifests/00-crds.yaml"},
				},
				s: st,
			},
			want: "https//raw.githubusercontent.com/jetstack/cert-manager/release-0.14/deploy/manifests/00-crds.yaml must be valid URL path to a raw file",
		}, {
			name: "test case 19 - invalid hook type 1",
			args: args{
				r: &release{
					Name:        "release19",
					Description: "",
					Namespace:   "namespace",
					Enabled:     true,
					Chart:       "repo/chartX",
					Version:     "1.0",
					ValuesFile:  "../../tests/values.yaml",
					Hooks:       map[string]interface{}{"afterDelete": "https://raw.githubusercontent.com/jetstack/cert-manager/release-0.14/deploy/manifests/00-crds.yaml"},
				},
				s: st,
			},
			want: "afterDelete is an Invalid hook type",
		}, {
			name: "test case 20 - invalid hook type 2",
			args: args{
				r: &release{
					Name:        "release20",
					Description: "",
					Namespace:   "namespace",
					Enabled:     true,
					Chart:       "repo/chartX",
					Version:     "1.0",
					ValuesFile:  "../../tests/values.yaml",
					Hooks:       map[string]interface{}{"PreDelete": "https://raw.githubusercontent.com/jetstack/cert-manager/release-0.14/deploy/manifests/00-crds.yaml"},
				},
				s: st,
			},
			want: "PreDelete is an Invalid hook type",
		}, {
			name: "test case 21",
			args: args{
				r: &release{
					Name:         "release21",
					Description:  "",
					Namespace:    "namespace",
					Enabled:      true,
					Chart:        "repo/chartX",
					Version:      "1.0",
					ValuesFile:   "../../tests/values.yaml",
					PostRenderer: "../../tests/post-renderer.sh",
					Test:         true,
				},
				s: st,
			},
			want: "",
		}, {
			name: "test case 22",
			args: args{
				r: &release{
					Name:         "release22",
					Description:  "",
					Namespace:    "namespace",
					Enabled:      true,
					Chart:        "repo/chartX",
					Version:      "1.0",
					ValuesFile:   "../../tests/values.yaml",
					PostRenderer: "doesnt-exist.sh",
					Test:         true,
				},
				s: st,
			},
			want: "doesnt-exist.sh must be valid relative (from dsf file) file path",
		}, {
			name: "test case 23 - executable hook type",
			args: args{
				r: &release{
					Name:        "release20",
					Description: "",
					Namespace:   "namespace",
					Enabled:     true,
					Chart:       "repo/chartX",
					Version:     "1.0",
					ValuesFile:  "../../tests/values.yaml",
					Hooks:       map[string]interface{}{preDelete: "../../tests/post-renderer.sh"},
				},
				s: st,
			},
			want: "",
		},
	}
	names := make(map[string]map[string]bool)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ""
			if r := tt.args.r.validate("testApp", names, &tt.args.s); r != nil {
				got = r.Error()
			}
			if got != tt.want {
				t.Errorf("release.validate() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_release_inheritHooks(t *testing.T) {
	st := state{
		Metadata:     make(map[string]string),
		Certificates: make(map[string]string),
		Settings: config{
			GlobalHooks: map[string]interface{}{
				preInstall:         "https://raw.githubusercontent.com/jetstack/cert-manager/release-0.14/deploy/manifests/00-crds.yaml",
				postInstall:        "https://raw.githubusercontent.com/jetstack/cert-manager/release-0.14/deploy/manifests/00-crds.yaml",
				"successCondition": "Complete",
				"successTimeout":   "60s",
			},
		},
		Namespaces: map[string]*namespace{"namespace": {false, limits{}, make(map[string]string), make(map[string]string), &quotas{}, false}},
		HelmRepos:  make(map[string]string),
		Apps:       make(map[string]*release),
	}

	type args struct {
		s state
		r *release
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "test case 1",
			args: args{
				r: &release{
					Name:        "release1 - Global hooks correctly inherited",
					Description: "",
					Namespace:   "namespace",
					Enabled:     true,
					Chart:       "repo/chartX",
					Version:     "1.0",
					ValuesFile:  "../../tests/values.yaml",
					Hooks: map[string]interface{}{
						postInstall:      "../../tests/values.yaml",
						preDelete:        "../../tests/values.yaml",
						"successTimeout": "360s",
					},
				},
				s: st,
			},
			want: "https://raw.githubusercontent.com/jetstack/cert-manager/release-0.14/deploy/manifests/00-crds.yaml -- " +
				"../../tests/values.yaml -- " +
				"../../tests/values.yaml -- " +
				"Complete -- 360s",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.args.r.inheritHooks(&tt.args.s)
			got := tt.args.r.Hooks[preInstall].(string) + " -- " + tt.args.r.Hooks[postInstall].(string) + " -- " + tt.args.r.Hooks[preDelete].(string) +
				" -- " + tt.args.r.Hooks["successCondition"].(string) + " -- " + tt.args.r.Hooks["successTimeout"].(string)
			if got != tt.want {
				t.Errorf("release.inheritHooks() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_release_getChartVersion(t *testing.T) {
	// version string = the first semver-valid string after the last hypen in the chart string.
	type args struct {
		r helmRelease
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "test case 1: there is a pre-release version",
			args: args{
				r: helmRelease{
					Revision:  0,
					Updated:   HelmTime{},
					Status:    "",
					Chart:     "elasticsearch-1.3.0-1",
					Namespace: "",
				},
			},
			want: "1.3.0-1",
		}, {
			name: "test case 2: normal case",
			args: args{
				r: helmRelease{
					Revision:  0,
					Updated:   HelmTime{},
					Status:    "",
					Chart:     "elasticsearch-1.3.0",
					Namespace: "",
				},
			},
			want: "1.3.0",
		}, {
			name: "test case 3: there is a hypen in the name",
			args: args{
				r: helmRelease{
					Revision:  0,
					Updated:   HelmTime{},
					Status:    "",
					Chart:     "elastic-search-1.3.0",
					Namespace: "",
				},
			},
			want: "1.3.0",
		}, {
			name: "test case 4: there is meta information",
			args: args{
				r: helmRelease{
					Revision:  0,
					Updated:   HelmTime{},
					Status:    "",
					Chart:     "elastic-search-1.3.0+meta.info",
					Namespace: "",
				},
			},
			want: "1.3.0+meta.info",
		}, {
			name: "test case 5: an invalid string",
			args: args{
				r: helmRelease{
					Revision:  0,
					Updated:   HelmTime{},
					Status:    "",
					Chart:     "foo",
					Namespace: "",
				},
			},
			want: "",
		}, {
			name: "test case 6: version includes v",
			args: args{
				r: helmRelease{
					Revision:  0,
					Updated:   HelmTime{},
					Status:    "",
					Chart:     "cert-manager-v0.5.2",
					Namespace: "",
				},
			},
			want: "v0.5.2",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.want)
			if got := tt.args.r.getChartVersion(); got != tt.want {
				t.Errorf("release.getChartVersion() = %v, want %v", got, tt.want)
			}
		})
	}
}
