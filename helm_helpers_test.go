package main

import (
	"os"
	"testing"
	"time"
)

func setupTestCase(t *testing.T) func(t *testing.T) {
	t.Log("setup test case")
	os.MkdirAll("/tmp/helmsman-tests/myapp", os.ModePerm)
	os.MkdirAll("/tmp/helmsman-tests/dir-with space/myapp", os.ModePerm)
	cmd := command{
		Cmd:         "bash",
		Args:        []string{"-c", "helm create  '/tmp/helmsman-tests/dir-with space/myapp'"},
		Description: "creating an empty local chart directory",
	}
	if exitCode, msg := cmd.exec(debug, verbose); exitCode != 0 {
		logError("Command returned with exit code: " + string(exitCode) + ". And error message: " + msg)
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
		name string
		args args
		want bool
	}{
		{
			name: "test case 1: valid local path with no chart",
			args: args{
				apps: map[string]*release{
					"app": &release{
						Name:            "",
						Description:     "",
						Namespace:       "",
						Enabled:         true,
						Chart:           "/tmp/helmsman-tests/myapp",
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
			name: "test case 2: invalid local path",
			args: args{
				apps: map[string]*release{
					"app": &release{
						Name:            "",
						Description:     "",
						Namespace:       "",
						Enabled:         true,
						Chart:           "/tmp/does-not-exist/myapp",
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
			name: "test case 3: valid chart local path with whitespace",
			args: args{
				apps: map[string]*release{
					"app": &release{
						Name:            "",
						Description:     "",
						Namespace:       "",
						Enabled:         true,
						Chart:           "/tmp/helmsman-tests/dir-with space/myapp",
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
			name: "test case 4: valid chart from repo",
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
		},
	}

	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if got, msg := validateReleaseCharts(tt.args.apps); got != tt.want {
				t.Errorf("getReleaseChartName() = %v, want %v , msg: %v", got, tt.want, msg)
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
				t.Errorf("getReleaseChartName() = %v, want %v", got, tt.want)
			}
		})
	}
}
