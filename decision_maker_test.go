package main

import (
	"testing"
)

func Test_getValuesFiles(t *testing.T) {
	type args struct {
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
					ValuesFile:  "test_files/values.yaml",
					Purge:       true,
					Test:        true,
				},
				//s: st,
			},
			want: " -f " + pwd + "/" + relativeDir + "/test_files/values.yaml",
		},
		{
			name: "test case 2",
			args: args{
				r: &release{
					Name:        "release1",
					Description: "",
					Namespace:   "namespace",
					Enabled:     true,
					Chart:       "repo/chartX",
					Version:     "1.0",
					ValuesFiles: []string{"test_files/values.yaml"},
					Purge:       true,
					Test:        true,
				},
				//s: st,
			},
			want: " -f " + pwd + "/" + relativeDir + "/test_files/values.yaml",
		},
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
					ValuesFiles: []string{"test_files/values.yaml", "test_files/values2.yaml"},
					Purge:       true,
					Test:        true,
				},
				//s: st,
			},
			want: " -f " + pwd + "/" + relativeDir + "/test_files/values.yaml -f " + pwd + "/" + relativeDir + "/test_files/values2.yaml",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getValuesFiles(tt.args.r); got != tt.want {
				t.Errorf("getValuesFiles() = %v, want %v", got, tt.want)
			}
		})
	}
}
