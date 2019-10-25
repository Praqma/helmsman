package main

import (
	"fmt"
	"os"
	"testing"
	"time"
)

func setupTestCase(t *testing.T) func(t *testing.T) {
	t.Log("setup test case")
	os.MkdirAll(os.TempDir()+"/helmsman-tests/myapp", os.ModePerm)
	os.MkdirAll(os.TempDir()+"/helmsman-tests/dir-with space/myapp", os.ModePerm)
	cmd := command{
		Cmd:         "helm",
		Args:        []string{"create", os.TempDir() + "/helmsman-tests/dir-with space/myapp"},
		Description: "creating an empty local chart directory",
	}
	if exitCode, msg := cmd.exec(debug, verbose); exitCode != 0 {
		logError(fmt.Sprintf("Command returned with exit code: %d. And error message: %s ", exitCode, msg))
	}

	return func(t *testing.T) {
		t.Log("teardown test case")
		//os.RemoveAll("/tmp/helmsman-tests/")
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
						Name:            "",
						Description:     "",
						Namespace:       "",
						Enabled:         true,
						Chart:           os.TempDir() + "/helmsman-tests/myapp",
						Version:         "",
						ValuesFile:      "",
						ValuesFiles:     []string{},
						SecretsFile:     "",
						SecretsFiles:    []string{},
						Purge:           false,
						Test:            false,
						Protected:       false,
						Wait:            false,
						Priority:        0,
						TillerNamespace: "",
						Set:             make(map[string]string),
						SetString:       make(map[string]string),
						HelmFlags:       []string{},
						NoHooks:         false,
						Timeout:         0,
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
						Name:            "",
						Description:     "",
						Namespace:       "",
						Enabled:         true,
						Chart:           os.TempDir() + "/does-not-exist/myapp",
						Version:         "",
						ValuesFile:      "",
						ValuesFiles:     []string{},
						SecretsFile:     "",
						SecretsFiles:    []string{},
						Purge:           false,
						Test:            false,
						Protected:       false,
						Wait:            false,
						Priority:        0,
						TillerNamespace: "",
						Set:             make(map[string]string),
						SetString:       make(map[string]string),
						HelmFlags:       []string{},
						NoHooks:         false,
						Timeout:         0,
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
						Name:            "",
						Description:     "",
						Namespace:       "",
						Enabled:         true,
						Chart:           os.TempDir() + "/helmsman-tests/dir-with space/myapp",
						Version:         "0.1.0",
						ValuesFile:      "",
						ValuesFiles:     []string{},
						SecretsFile:     "",
						SecretsFiles:    []string{},
						Purge:           false,
						Test:            false,
						Protected:       false,
						Wait:            false,
						Priority:        0,
						TillerNamespace: "",
						Set:             make(map[string]string),
						SetString:       make(map[string]string),
						HelmFlags:       []string{},
						NoHooks:         false,
						Timeout:         0,
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
						Name:            "",
						Description:     "",
						Namespace:       "",
						Enabled:         true,
						Chart:           "stable/prometheus",
						Version:         "",
						ValuesFile:      "",
						ValuesFiles:     []string{},
						SecretsFile:     "",
						SecretsFiles:    []string{},
						Purge:           false,
						Test:            false,
						Protected:       false,
						Wait:            false,
						Priority:        0,
						TillerNamespace: "",
						Set:             make(map[string]string),
						SetString:       make(map[string]string),
						HelmFlags:       []string{},
						NoHooks:         false,
						Timeout:         0,
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
						Name:            "app",
						Description:     "",
						Namespace:       "",
						Enabled:         true,
						Chart:           os.TempDir() + "/does-not-exist/myapp",
						Version:         "",
						ValuesFile:      "",
						ValuesFiles:     []string{},
						SecretsFile:     "",
						SecretsFiles:    []string{},
						Purge:           false,
						Test:            false,
						Protected:       false,
						Wait:            false,
						Priority:        0,
						TillerNamespace: "",
						Set:             make(map[string]string),
						SetString:       make(map[string]string),
						HelmFlags:       []string{},
						NoHooks:         false,
						Timeout:         0,
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
						Name:            "app",
						Description:     "",
						Namespace:       "",
						Enabled:         true,
						Chart:           os.TempDir() + "/does-not-exist/myapp",
						Version:         "",
						ValuesFile:      "",
						ValuesFiles:     []string{},
						SecretsFile:     "",
						SecretsFiles:    []string{},
						Purge:           false,
						Test:            false,
						Protected:       false,
						Wait:            false,
						Priority:        0,
						TillerNamespace: "",
						Set:             make(map[string]string),
						SetString:       make(map[string]string),
						HelmFlags:       []string{},
						NoHooks:         false,
						Timeout:         0,
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
			if got, msg := validateReleaseCharts(tt.args.apps); got != tt.want {
				t.Errorf("validateReleaseCharts() = %v, want %v , msg: %v", got, tt.want, msg)
			}
		})
	}
}

func Test_getReleaseChartVersion(t *testing.T) {
	// version string = the first semver-valid string after the last hypen in the chart string.

	type args struct {
		r releaseState
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "test case 1: there is a pre-release version",
			args: args{
				r: releaseState{
					Revision:        0,
					Updated:         time.Now(),
					Status:          "",
					Chart:           "elasticsearch-1.3.0-1",
					Namespace:       "",
					TillerNamespace: "",
				},
			},
			want: "1.3.0-1",
		}, {
			name: "test case 2: normal case",
			args: args{
				r: releaseState{
					Revision:        0,
					Updated:         time.Now(),
					Status:          "",
					Chart:           "elasticsearch-1.3.0",
					Namespace:       "",
					TillerNamespace: "",
				},
			},
			want: "1.3.0",
		}, {
			name: "test case 3: there is a hypen in the name",
			args: args{
				r: releaseState{
					Revision:        0,
					Updated:         time.Now(),
					Status:          "",
					Chart:           "elastic-search-1.3.0",
					Namespace:       "",
					TillerNamespace: "",
				},
			},
			want: "1.3.0",
		}, {
			name: "test case 4: there is meta information",
			args: args{
				r: releaseState{
					Revision:        0,
					Updated:         time.Now(),
					Status:          "",
					Chart:           "elastic-search-1.3.0+meta.info",
					Namespace:       "",
					TillerNamespace: "",
				},
			},
			want: "1.3.0+meta.info",
		}, {
			name: "test case 5: an invalid string",
			args: args{
				r: releaseState{
					Revision:        0,
					Updated:         time.Now(),
					Status:          "",
					Chart:           "foo",
					Namespace:       "",
					TillerNamespace: "",
				},
			},
			want: "",
		}, {
			name: "test case 6: version includes v",
			args: args{
				r: releaseState{
					Revision:        0,
					Updated:         time.Now(),
					Status:          "",
					Chart:           "cert-manager-v0.5.2",
					Namespace:       "",
					TillerNamespace: "",
				},
			},
			want: "v0.5.2",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.want)
			if got := getReleaseChartVersion(tt.args.r); got != tt.want {
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
					Chart:     "./test_files/chart-test",
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
			got, _ := getChartVersion(tt.args.r)
			if got != tt.want {
				t.Errorf("getChartVersion() = %v, want %v", got, tt.want)
			}
		})
	}
}
