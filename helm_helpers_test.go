package main

import (
	"testing"
	"time"
)

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
