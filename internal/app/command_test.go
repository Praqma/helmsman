package app

import (
	"testing"
)

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
		err    error
		output string
	}
	tests := []struct {
		name     string
		input    input
		expected expected
	}{
		{
			name: "echo",
			input: input{
				cmd:  "bash",
				args: []string{"-c", "echo this is fun"},
				desc: "A bash command execution test with echo.",
			},
			expected: expected{
				output: "this is fun",
			},
		}, {
			name: "exitCode",
			input: input{
				cmd:  "bash",
				args: []string{"-c", "echo $?"},
				desc: "A bash command execution test with exitCode.",
			},
			expected: expected{
				output: "0",
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
			res, err := c.Exec()
			if err != nil && tt.expected.err.Error() != err.Error() {
				t.Errorf("unexpected error running command.exec(): %v", err)
			}

			if res.output != tt.expected.output {
				t.Errorf("command.exec() expected: %v, got %v", tt.expected.output, res.output)
			}
		})
	}
}
