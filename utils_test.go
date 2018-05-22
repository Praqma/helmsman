package main

import "testing"

// func Test_printMap(t *testing.T) {
// 	type args struct {
// 		m map[string]string
// 	}
// 	tests := []struct {
// 		name string
// 		args args
// 	}{
// 	// TODO: Add test cases.
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			printMap(tt.args.m)
// 		})
// 	}
// }

func Test_fromTOML(t *testing.T) {
	type args struct {
		file string
		s    *state
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "test case 1 -- invalid TOML",
			args: args{
				file: "test_files/invalid_example.toml",
				s:    new(state),
			},
			want: false,
		}, {
			name: "test case 2 -- valid TOML",
			args: args{
				file: "example.toml",
				s:    new(state),
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, _ := fromTOML(tt.args.file, tt.args.s); got != tt.want {
				t.Errorf("fromToml() = %v, want %v", got, tt.want)
			}
		})
	}
}

// func Test_toTOML(t *testing.T) {
// 	type args struct {
// 		file string
// 		s    *state
// 	}
// 	tests := []struct {
// 		name string
// 		args args
// 	}{
// 	// TODO: Add test cases.
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			toTOML(tt.args.file, tt.args.s)
// 		})
// 	}
// }

func Test_fromYAML(t *testing.T) {
	type args struct {
		file string
		s    *state
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "test case 1 -- invalid YAML",
			args: args{
				file: "test_files/invalid_example.yaml",
				s:    new(state),
			},
			want: false,
		}, {
			name: "test case 2 -- valid TOML",
			args: args{
				file: "example.yaml",
				s:    new(state),
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, _ := fromYAML(tt.args.file, tt.args.s); got != tt.want {
				t.Errorf("fromYaml() = %v, want %v", got, tt.want)
			}
		})
	}
}

// func Test_toYAML(t *testing.T) {
// 	type args struct {
// 		file string
// 		s    *state
// 	}
// 	tests := []struct {
// 		name string
// 		args args
// 	}{
// 	// TODO: Add test cases.
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			toYAML(tt.args.file, tt.args.s)
// 		})
// 	}
// }

func Test_isOfType(t *testing.T) {
	type args struct {
		filename string
		filetype string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "test case 1 -- valid xml check",
			args: args{
				filename: "name.xml",
				filetype: ".xml",
			},
			want: true,
		}, {
			name: "test case 2 -- valid yaml check",
			args: args{
				filename: "another_name.yaml",
				filetype: ".yaml",
			},
			want: true,
		}, {
			name: "test case 3 -- invalid yaml check",
			args: args{
				filename: "name.xml",
				filetype: ".yaml",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isOfType(tt.args.filename, tt.args.filetype); got != tt.want {
				t.Errorf("isOfType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_readFile(t *testing.T) {
	type args struct {
		filepath string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "test case 1 -- successful reading.",
			args: args{
				filepath: "test_files/values.yaml",
			},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := readFile(tt.args.filepath); got != tt.want {
				t.Errorf("readFile() = %v, want %v", got, tt.want)
			}
		})
	}
}

// func Test_printHelp(t *testing.T) {
// 	tests := []struct {
// 		name string
// 	}{
// 	// TODO: Add test cases.
// 	}
// 	for range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			printHelp()
// 		})
// 	}
// }
