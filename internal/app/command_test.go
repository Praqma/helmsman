package app

import (
	"testing"
)

func TestToolExists(t *testing.T) {
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
				tool: helmBin,
			},
			want: true,
		}, {
			name: "test case 2 -- checking kubectl exists.",
			args: args{
				tool: kubectlBin,
			},
			want: true,
		}, {
			name: "test case 3 -- checking nonExistingTool exists.",
			args: args{
				tool: "nonExistingTool",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ToolExists(tt.args.tool); got != tt.want {
				t.Errorf("toolExists() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCommandExec(t *testing.T) {
	type input struct {
		cmd  string
		args []string
		desc string
	}
	type expected struct {
		code   int
		err    error
		output string
	}
	tests := []struct {
		name  string
		input input
		want  expected
	}{
		{
			name: "echo",
			input: input{
				cmd:  "bash",
				args: []string{"-c", "echo this is fun"},
				desc: "A bash command execution test with echo.",
			},
			want: expected{
				code:   0,
				output: "this is fun",
				err:    nil,
			},
		}, {
			name: "exitCode",
			input: input{
				cmd:  "bash",
				args: []string{"-c", "echo $?"},
				desc: "A bash command execution test with exitCode.",
			},
			want: expected{
				code:   0,
				output: "0",
				err:    nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := Command{
				Cmd:         tt.input.cmd,
				Args:        tt.input.args,
				Description: tt.input.desc,
			}
			got, err := c.Exec()
			if err != tt.want.err {
				t.Errorf("command.exec() unexpected error got = %v, want %v", err, tt.want.err)
			}
			if got.code != tt.want.code {
				t.Errorf("command.exec() unexpected code got = %v, want %v", got.code, tt.want.code)
			}
			if got.output != tt.want.output {
				t.Errorf("command.exec() unexpected output got = %v, want %v", got.output, tt.want.output)
			}
		})
	}
}
