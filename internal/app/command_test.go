package app

import (
	"strings"
	"testing"
)

func Test_command_exec(t *testing.T) {
	type fields struct {
		Cmd         string
		Args        []string
		Description string
	}
	tests := []struct {
		name   string
		fields fields
		want   int
		want1  string
	}{
		{
			name: "echo",
			fields: fields{
				Cmd:         "bash",
				Args:        []string{"-c", "echo this is fun"},
				Description: "A bash command execution test with echo.",
			},
			want:  0,
			want1: "this is fun",
		}, {
			name: "exitCode",
			fields: fields{
				Cmd:         "bash",
				Args:        []string{"-c", "echo $?"},
				Description: "A bash command execution test with exitCode.",
			},
			want:  0,
			want1: "0",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := Command{
				Cmd:         tt.fields.Cmd,
				Args:        tt.fields.Args,
				Description: tt.fields.Description,
			}
			got := c.Exec()
			if got.code != tt.want {
				t.Errorf("command.exec() got = %v, want %v", got.code, tt.want)
			}
			if strings.TrimSpace(got.output) != tt.want1 {
				t.Errorf("command.exec() got1 = %v, want %v", got.output, tt.want1)
			}
		})
	}
}
