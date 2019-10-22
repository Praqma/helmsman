package main

import "testing"

func Test_toolExists(t *testing.T) {
	type args struct {
		tool string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "test case 1 -- checking helm exists.",
			args: args{
				tool: "helm",
			},
			want: true,
		}, {
			name: "test case 2 -- checking kubectl exists.",
			args: args{
				tool: "kubectl",
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := toolExists(tt.args.tool); got != tt.want {
				t.Errorf("toolExists() = %v, want %v", got, tt.want)
			}
		})
	}
}

// func Test_addNamespaces(t *testing.T) {
// 	type args struct {
// 		namespaces map[string]string
// 	}
// 	tests := []struct {
// 		name string
// 		args args
// 	}{
// 	// TODO: Add test cases.
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			addNamespaces(tt.args.namespaces)
// 		})
// 	}
// }

// func Test_validateReleaseCharts(t *testing.T) {
// 	type args struct {
// 		apps map[string]release
// 	}
// 	tests := []struct {
// 		name string
// 		args args
// 		want bool
// 	}{
// 	// TODO: Add test cases.
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			if got := validateReleaseCharts(tt.args.apps); got != tt.want {
// 				t.Errorf("validateReleaseCharts() = %v, want %v", got, tt.want)
// 			}
// 		})
// 	}
// }

// func Test_addHelmRepos(t *testing.T) {
// 	type args struct {
// 		repos map[string]string
// 	}
// 	tests := []struct {
// 		name string
// 		args args
// 		want bool
// 	}{
// 	// TODO: Add test cases.
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			if got := addHelmRepos(tt.args.repos); got != tt.want {
// 				t.Errorf("addHelmRepos() = %v, want %v", got, tt.want)
// 			}
// 		})
// 	}
// }

// func Test_setKubeContext(t *testing.T) {
// 	type args struct {
// 		context string
// 	}
// 	tests := []struct {
// 		name string
// 		args args
// 		want bool
// 	}{
// 	// TODO: Add test cases.
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			if got := setKubeContext(tt.args.context); got != tt.want {
// 				t.Errorf("setKubeContext() = %v, want %v", got, tt.want)
// 			}
// 		})
// 	}
// }

// func Test_createContext(t *testing.T) {
// 	tests := []struct {
// 		name string
// 		want bool
// 	}{
// 	// TODO: Add test cases.
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			if got := createContext(); got != tt.want {
// 				t.Errorf("createContext() = %v, want %v", got, tt.want)
// 			}
// 		})
// 	}
// }
