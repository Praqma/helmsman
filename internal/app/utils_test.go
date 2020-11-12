package app

import (
	"os"
	"testing"
)

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
