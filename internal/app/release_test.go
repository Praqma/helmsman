package app

import (
	"fmt"
	"os"
	"strings"
	"testing"
)

func setupTestCase(t *testing.T) func(t *testing.T) {
	t.Log("setup test case")
	os.MkdirAll(os.TempDir()+"/helmsman-tests/myapp", os.ModePerm)
	os.MkdirAll(os.TempDir()+"/helmsman-tests/dir-with space/myapp", os.ModePerm)
	cmd := helmCmd([]string{"create", os.TempDir() + "/helmsman-tests/dir-with space/myapp"}, "creating an empty local chart directory")
	if exitCode, msg, _ := cmd.exec(debug, verbose); exitCode != 0 {
		log.Fatal(fmt.Sprintf("Command returned with exit code: %d. And error message: %s ", exitCode, msg))
	}

	return func(t *testing.T) {
		t.Log("teardown test case")
		//os.RemoveAll("/tmp/helmsman-tests/")
	}
}

func Test_validateRelease(t *testing.T) {
	st := state{
		Metadata:     make(map[string]string),
		Certificates: make(map[string]string),
		Settings:     (config{}),
		Namespaces:   map[string]namespace{"namespace": namespace{false, limits{}, make(map[string]string), make(map[string]string)}},
		HelmRepos:    make(map[string]string),
		Apps:         make(map[string]*release),
	}

	type args struct {
		s state
		r *release
	}
	tests := []struct {
		name  string
		args  args
		want  bool
		want1 string
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
			want:  true,
			want1: "",
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
			want:  false,
			want1: "valuesFile must be a valid relative (from dsf file) file path for a yaml file, or can be left empty (provided path resolved to \"xyz.yaml\").",
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
			want:  false,
			want1: "valuesFile must be a valid relative (from dsf file) file path for a yaml file, or can be left empty (provided path resolved to \"../../tests/values.xml\").",
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
			want:  false,
			want1: "release name must be unique within a given Tiller.",
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
			want:  true,
			want1: "",
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
			want:  false,
			want1: "release targeted namespace can't be empty.",
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
			want:  false,
			want1: "chart can't be empty and must be of the format: repo/chart.",
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
			want:  false,
			want1: "chart can't be empty and must be of the format: repo/chart.",
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
			want:  false,
			want1: "version can't be empty.",
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
			want:  true,
			want1: "",
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
			want:  false,
			want1: "valuesFile and valuesFiles should not be used together.",
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
			want:  false,
			want1: "valuesFiles must be valid relative (from dsf file) file paths for a yaml file; path at index 0 provided path resolved to \"xyz.yaml\".",
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
			want:  true,
			want1: "",
		}, {
			name: "test case 14",
			args: args{
				r: &release{
					Name:        "release14",
					Description: "",
					Namespace:   "namespace",
					Enabled:     true,
					Chart:       "repo/chartX",
					Version:     "1.0",
					ValuesFiles: []string{"./../../tests/values.yaml", "../../tests/values2.yaml"},
					Test:        true,
					Set:         map[string]string{"some_var": "$SOME_VAR"},
				},
				s: st,
			},
			want:  false,
			want1: "env var [ $SOME_VAR ] is not set, but is wanted to be passed for [ some_var ] in [[ release14 ]]",
		},
	}
	names := make(map[string]map[string]bool)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := tt.args.r.validate("testApp", names, &tt.args.s)
			if got != tt.want {
				t.Errorf("validateRelease() got = %v, want %v", got, tt.want)
			}
			if strings.TrimSpace(got1) != tt.want1 {
				t.Errorf("validateRelease() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func Test_validateReleaseCharts(t *testing.T) {
	type args struct {
		apps map[string]*release
	}
	tests := []struct {
		name       string
		targetFlag []string
		groupFlag  []string
		args       args
		want       bool
	}{
		{
			name:       "test case 1: valid local path with no chart",
			targetFlag: []string{},
			args: args{
				apps: map[string]*release{
					"app": &release{
						Name:         "",
						Description:  "",
						Namespace:    "",
						Enabled:      true,
						Chart:        os.TempDir() + "/helmsman-tests/myapp",
						Version:      "",
						ValuesFile:   "",
						ValuesFiles:  []string{},
						SecretsFile:  "",
						SecretsFiles: []string{},
						Test:         false,
						Protected:    false,
						Wait:         false,
						Priority:     0,
						Set:          make(map[string]string),
						SetString:    make(map[string]string),
						HelmFlags:    []string{},
						NoHooks:      false,
						Timeout:      0,
					},
				},
			},
			want: false,
		}, {
			name:       "test case 2: invalid local path",
			targetFlag: []string{},
			args: args{
				apps: map[string]*release{
					"app": &release{
						Name:         "",
						Description:  "",
						Namespace:    "",
						Enabled:      true,
						Chart:        os.TempDir() + "/does-not-exist/myapp",
						Version:      "",
						ValuesFile:   "",
						ValuesFiles:  []string{},
						SecretsFile:  "",
						SecretsFiles: []string{},
						Test:         false,
						Protected:    false,
						Wait:         false,
						Priority:     0,
						Set:          make(map[string]string),
						SetString:    make(map[string]string),
						HelmFlags:    []string{},
						NoHooks:      false,
						Timeout:      0,
					},
				},
			},
			want: false,
		}, {
			name:       "test case 3: valid chart local path with whitespace",
			targetFlag: []string{},
			args: args{
				apps: map[string]*release{
					"app": &release{
						Name:         "",
						Description:  "",
						Namespace:    "",
						Enabled:      true,
						Chart:        os.TempDir() + "/helmsman-tests/dir-with space/myapp",
						Version:      "0.1.0",
						ValuesFile:   "",
						ValuesFiles:  []string{},
						SecretsFile:  "",
						SecretsFiles: []string{},
						Test:         false,
						Protected:    false,
						Wait:         false,
						Priority:     0,
						Set:          make(map[string]string),
						SetString:    make(map[string]string),
						HelmFlags:    []string{},
						NoHooks:      false,
						Timeout:      0,
					},
				},
			},
			want: true,
		}, {
			name:       "test case 4: valid chart from repo",
			targetFlag: []string{},
			args: args{
				apps: map[string]*release{
					"app": &release{
						Name:         "",
						Description:  "",
						Namespace:    "",
						Enabled:      true,
						Chart:        "stable/prometheus",
						Version:      "9.5.2",
						ValuesFile:   "",
						ValuesFiles:  []string{},
						SecretsFile:  "",
						SecretsFiles: []string{},
						Test:         false,
						Protected:    false,
						Wait:         false,
						Priority:     0,
						Set:          make(map[string]string),
						SetString:    make(map[string]string),
						HelmFlags:    []string{},
						NoHooks:      false,
						Timeout:      0,
					},
				},
			},
			want: true,
		}, {
			name:       "test case 5: invalid local path for chart ignored with -target flag, while other app was targeted",
			targetFlag: []string{"notThisOne"},
			args: args{
				apps: map[string]*release{
					"app": &release{
						Name:         "app",
						Description:  "",
						Namespace:    "",
						Enabled:      true,
						Chart:        os.TempDir() + "/does-not-exist/myapp",
						Version:      "",
						ValuesFile:   "",
						ValuesFiles:  []string{},
						SecretsFile:  "",
						SecretsFiles: []string{},
						Test:         false,
						Protected:    false,
						Wait:         false,
						Priority:     0,
						Set:          make(map[string]string),
						SetString:    make(map[string]string),
						HelmFlags:    []string{},
						NoHooks:      false,
						Timeout:      0,
					},
				},
			},
			want: true,
		}, {
			name:       "test case 6: invalid local path for chart included with -target flag",
			targetFlag: []string{"app"},
			args: args{
				apps: map[string]*release{
					"app": &release{
						Name:         "app",
						Description:  "",
						Namespace:    "",
						Enabled:      true,
						Chart:        os.TempDir() + "/does-not-exist/myapp",
						Version:      "",
						ValuesFile:   "",
						ValuesFiles:  []string{},
						SecretsFile:  "",
						SecretsFiles: []string{},
						Test:         false,
						Protected:    false,
						Wait:         false,
						Priority:     0,
						Set:          make(map[string]string),
						SetString:    make(map[string]string),
						HelmFlags:    []string{},
						NoHooks:      false,
						Timeout:      0,
					},
				},
			},
			want: false,
		},
	}

	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			targetMap = make(map[string]bool)
			groupMap = make(map[string]bool)
			for _, target := range tt.targetFlag {
				targetMap[target] = true
			}
			for _, group := range tt.groupFlag {
				groupMap[group] = true
			}
			err := validateReleaseCharts(tt.args.apps)
			switch err.(type) {
			case nil:
				if tt.want != true {
					t.Errorf("validateReleaseCharts() = %v, want error", err)
				}
			case error:
				if tt.want != false {
					t.Errorf("validateReleaseCharts() = %v, want nil", err)
				}
			}
		})
	}
}

func Test_getReleaseChartVersion(t *testing.T) {
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
				t.Errorf("getReleaseChartVersion() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getChartVersion(t *testing.T) {
	// version string = the first semver-valid string after the last hypen in the chart string.
	type args struct {
		r *release
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "getChartVersion - local chart should return given release version",
			args: args{
				r: &release{
					Name:      "release1",
					Namespace: "namespace",
					Version:   "1.0.0",
					Chart:     "./../../tests/chart-test",
					Enabled:   true,
				},
			},
			want: "1.0.0",
		},
		{
			name: "getChartVersion - unknown chart should error",
			args: args{
				r: &release{
					Name:      "release1",
					Namespace: "namespace",
					Version:   "1.0.0",
					Chart:     "random-chart-name-1f8147",
					Enabled:   true,
				},
			},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.want)
			got, _ := tt.args.r.getChartVersion()
			if got != tt.want {
				t.Errorf("getChartVersion() = %v, want %v", got, tt.want)
			}
		})
	}
}
