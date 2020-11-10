package app

import (
	"os"
	"reflect"
	"testing"
)

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
				file: "../../../tests/invalid_example.toml",
				s:    new(state),
			},
			want: false,
		}, {
			name: "test case 2 -- valid TOML",
			args: args{
				file: "../../examples/example.toml",
				s:    new(state),
			},
			want: true,
		},
	}
	os.Setenv("ORG_PATH", "sample")
	os.Setenv("VALUE", "sample")
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, _ := fromTOML(tt.args.file, tt.args.s); got != tt.want {
				t.Errorf("fromToml() = %v, want %v", got, tt.want)
			}
		})
	}
	os.Unsetenv("ORG_PATH")
	os.Unsetenv("VALUE")
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
				file: "../../examples/example.toml",
				s:    new(state),
			},
			section: "Settings",
			field:   "ClusterURI",
			want:    "https://192.168.99.100:8443",
		},
		{
			name: "test case 2 -- valid TOML expand org",
			args: args{
				file: "../../examples/example.toml",
				s:    new(state),
			},
			section: "Metadata",
			field:   "org",
			want:    "example.com/sample/",
		},
	}
	os.Setenv("SET_URI", "https://192.168.99.100:8443")
	os.Setenv("ORG_PATH", "sample")
	os.Setenv("VALUE", "sample")
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
	os.Unsetenv("ORG_PATH")
	os.Unsetenv("SET_URI")
	os.Unsetenv("VALUE")
}

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
				file: "../../tests/invalid_example.yaml",
				s:    new(state),
			},
			want: false,
		}, {
			name: "test case 2 -- valid TOML",
			args: args{
				file: "../../examples/example.yaml",
				s:    new(state),
			},
			want: true,
		},
	}
	os.Setenv("VALUE", "sample")
	os.Setenv("ORG_PATH", "sample")
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, _ := fromYAML(tt.args.file, tt.args.s); got != tt.want {
				t.Errorf("fromYaml() = %v, want %v", got, tt.want)
			}
		})
	}
	os.Unsetenv("ORG_PATH")
	os.Unsetenv("VALUE")
}

func Test_fromYAML_UnsetVars(t *testing.T) {
	type args struct {
		file string
		s    *state
	}
	tests := []struct {
		name      string
		args      args
		targetVar string
		want      bool
	}{
		{
			name: "test case 1 -- unset ORG_PATH env var",
			args: args{
				file: "../../examples/example.yaml",
				s:    new(state),
			},
			targetVar: "ORG_PATH",
			want:      false,
		},
		{
			name: "test case 2 -- unset VALUE var",
			args: args{
				file: "../../examples/example.yaml",
				s:    new(state),
			},
			targetVar: "VALUE",
			want:      false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.targetVar == "ORG_PATH" {
				os.Setenv("VALUE", "sample")
			} else if tt.targetVar == "VALUE" {
				os.Setenv("ORG_PATH", "sample")
			}
			if got, _ := fromYAML(tt.args.file, tt.args.s); got != tt.want {
				t.Errorf("fromYaml() = %v, want %v", got, tt.want)
			}
		})
		os.Unsetenv("ORG_PATH")
		os.Unsetenv("VALUE")
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
				file: "../../examples/example.yaml",
				s:    new(state),
			},
			section: "Settings",
			field:   "ClusterURI",
			want:    "https://192.168.99.100:8443",
		},
		{
			name: "test case 2 -- valid YAML expand org",
			args: args{
				file: "../../examples/example.yaml",
				s:    new(state),
			},
			section: "Metadata",
			field:   "org",
			want:    "example.com/sample/",
		},
	}
	os.Setenv("SET_URI", "https://192.168.99.100:8443")
	os.Setenv("ORG_PATH", "sample")
	os.Setenv("VALUE", "sample")
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
	os.Unsetenv("ORG_PATH")
	os.Unsetenv("SET_URI")
	os.Unsetenv("VALUE")
}

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
				filepath: "../../tests/values.yaml",
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

func Test_eyamlSecrets(t *testing.T) {
	type args struct {
		r *release
		s *config
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "decryptSecrets - valid eyaml-based secrets decryption",
			args: args{
				s: &config{
					EyamlEnabled:        true,
					EyamlPublicKeyPath:  "./../../tests/keys/public_key.pkcs7.pem",
					EyamlPrivateKeyPath: "./../../tests/keys/private_key.pkcs7.pem",
				},
				r: &release{
					Name:        "release1",
					Namespace:   "namespace",
					Version:     "1.0.0",
					Enabled:     true,
					SecretsFile: "./../../tests/secrets/valid_eyaml_secrets.yaml",
				},
			},
			want: true,
		},
		{
			name: "decryptSecrets - not existing eyaml-based secrets file",
			args: args{
				s: &config{
					EyamlEnabled:        true,
					EyamlPublicKeyPath:  "./../../tests/keys/public_key.pkcs7.pem",
					EyamlPrivateKeyPath: "./../../tests/keys/private_key.pkcs7.pem",
				},
				r: &release{
					Name:        "release1",
					Namespace:   "namespace",
					Version:     "1.0.0",
					Enabled:     true,
					SecretsFile: "./../../tests/secrets/invalid_eyaml_secrets.yaml",
				},
			},
			want: false,
		},
		{
			name: "decryptSecrets - not existing eyaml key",
			args: args{
				s: &config{
					EyamlEnabled:        true,
					EyamlPublicKeyPath:  "./../../tests/keys/public_key.pkcs7.pem2",
					EyamlPrivateKeyPath: "./../../tests/keys/private_key.pkcs7.pem",
				},
				r: &release{
					Name:        "release1",
					Namespace:   "namespace",
					Version:     "1.0.0",
					Enabled:     true,
					SecretsFile: "./../../tests/secrets/valid_eyaml_secrets.yaml",
				},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.want)
			settings = &config{}
			settings.EyamlEnabled = tt.args.s.EyamlEnabled
			settings.EyamlPublicKeyPath = tt.args.s.EyamlPublicKeyPath
			settings.EyamlPrivateKeyPath = tt.args.s.EyamlPrivateKeyPath
			err := decryptSecret(tt.args.r.SecretsFile)
			switch err.(type) {
			case nil:
				if tt.want != true {
					t.Errorf("decryptSecret() = %v, want error", err)
				}
			case error:
				if tt.want != false {
					t.Errorf("decryptSecret() = %v, want nil", err)
				}
			}
			if _, err := os.Stat(tt.args.r.SecretsFile + ".dec"); err == nil {
				defer deleteFile(tt.args.r.SecretsFile + ".dec")
			}
		})
	}
}
