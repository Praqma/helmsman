package main

import (
	"os"
	"reflect"
	"testing"
)

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
func Test_fromTOML_Expand(t *testing.T) {
	type args struct {
		file string
		s    *state
	}
	tests := []struct {
		name    string
		args    args
		section string
		field   string
		want    string
	}{
		{
			name: "test case 1 -- valid TOML expand ClusterURI",
			args: args{
				file: "example.toml",
				s:    new(state),
			},
			section: "Settings",
			field:   "ClusterURI",
			want:    "https://192.168.99.100:8443",
		},
		{
			name: "test case 2 -- valid TOML expand org",
			args: args{
				file: "example.toml",
				s:    new(state),
			},
			section: "Metadata",
			field:   "org",
			want:    "example.com/sample/",
		},
	}
	os.Setenv("SET_URI", "https://192.168.99.100:8443")
	os.Setenv("ORG_PATH", "sample")
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err, msg := fromTOML(tt.args.file, tt.args.s)
			if !err {
				t.Errorf("fromToml(), got: %v", msg)
			}

			tomlVal := reflect.ValueOf(tt.args.s).Elem()
			tomlType := reflect.TypeOf(tt.args.s)

			if tomlType.Kind() != reflect.Struct {

				section := tomlVal.FieldByName(tt.section)
				sectionType := reflect.TypeOf(section)

				if section.IsValid() && section.Kind() == reflect.Struct {
					field := section.FieldByName(tt.field)
					if sectionType.Kind() == reflect.String {
						if field.String() != tt.want {
							t.Errorf("fromToml().section.field = %v, got: %v", tt.want, field.String())
						}
					}
				} else if section.IsValid() && section.Kind() == reflect.Map {
					found := false
					value := ""
					for _, key := range section.MapKeys() {
						if key.String() == tt.field {
							found = true
							value = section.MapIndex(key).String()
						}
					}
					if !found {
						t.Errorf("fromToml().section.field = '%v' not found", tt.field)
					} else if value != tt.want {
						t.Errorf("fromToml().section.field = %v, got: %v", tt.want, value)
					}

				} else {
					t.Errorf("fromToml().section = struct, got: %v", sectionType.Kind())
				}

			} else {
				t.Errorf("fromToml() = struct, got: %v", tomlType.Kind())
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

func Test_fromYAML_Expand(t *testing.T) {
	type args struct {
		file string
		s    *state
	}
	tests := []struct {
		name    string
		args    args
		section string
		field   string
		want    string
	}{
		{
			name: "test case 1 -- valid YAML expand ClusterURI",
			args: args{
				file: "example.yaml",
				s:    new(state),
			},
			section: "Settings",
			field:   "ClusterURI",
			want:    "https://192.168.99.100:8443",
		},
		{
			name: "test case 2 -- valid YAML expand org",
			args: args{
				file: "example.yaml",
				s:    new(state),
			},
			section: "Metadata",
			field:   "org",
			want:    "example.com/sample/",
		},
	}
	os.Setenv("SET_URI", "https://192.168.99.100:8443")
	os.Setenv("ORG_PATH", "sample")
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err, msg := fromYAML(tt.args.file, tt.args.s)
			if !err {
				t.Errorf("fromYaml(), got: %v", msg)
			}

			yamlVal := reflect.ValueOf(tt.args.s).Elem()
			yamlType := reflect.TypeOf(tt.args.s)

			if yamlType.Kind() != reflect.Struct {

				section := yamlVal.FieldByName(tt.section)
				sectionType := reflect.TypeOf(section)

				if section.IsValid() && section.Kind() == reflect.Struct {
					field := section.FieldByName(tt.field)
					if sectionType.Kind() == reflect.String {
						if field.String() != tt.want {
							t.Errorf("fromYaml().section.field = %v, got: %v", tt.want, field.String())
						}
					}
				} else if section.IsValid() && section.Kind() == reflect.Map {
					found := false
					value := ""
					for _, key := range section.MapKeys() {
						if key.String() == tt.field {
							found = true
							value = section.MapIndex(key).String()
						}
					}
					if !found {
						t.Errorf("fromYaml().section.field = '%v' not found", tt.field)
					} else if value != tt.want {
						t.Errorf("fromYaml().section.field = %v, got: %v", tt.want, value)
					}

				} else {
					t.Errorf("fromYaml().section = struct, got: %v", sectionType.Kind())
				}

			} else {
				t.Errorf("fromYaml() = struct, got: %v", yamlType.Kind())
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
		filename  string
		filetypes []string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "test case 1 -- valid xml check",
			args: args{
				filename:  "name.xml",
				filetypes: []string{".xml"},
			},
			want: true,
		}, {
			name: "test case 2 -- valid yaml check",
			args: args{
				filename:  "another_name.yaml",
				filetypes: []string{".yaml", ".yml"},
			},
			want: true,
		}, {
			name: "test case 3 -- valid (short) yaml check",
			args: args{
				filename:  "another_name.yml",
				filetypes: []string{".yaml", ".yml"},
			},
			want: true,
		}, {
			name: "test case 4 -- invalid yaml check",
			args: args{
				filename:  "name.xml",
				filetypes: []string{".yaml", ".yml"},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isOfType(tt.args.filename, tt.args.filetypes); got != tt.want {
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
