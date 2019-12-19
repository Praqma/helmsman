package app

import (
	"strings"
	"testing"
)

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
					Purge:       true,
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
					Purge:       true,
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
					Purge:       true,
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
					Purge:       true,
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
					Purge:       true,
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
					Purge:       true,
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
					Purge:       true,
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
					Purge:       true,
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
					Purge:       true,
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
					Purge:       true,
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
					Purge:       true,
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
					Purge:       true,
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
					Purge:       true,
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
					Purge:       true,
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
			got, got1 := validateRelease("testApp", tt.args.r, names, tt.args.s)
			if got != tt.want {
				t.Errorf("validateRelease() got = %v, want %v", got, tt.want)
			}
			if strings.TrimSpace(got1) != tt.want1 {
				t.Errorf("validateRelease() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
