package cmd

import (
	"testing"
)

func TestParseWatchArgs(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		wantFlags   []string
		wantCommand []string
	}{
		{
			name:        "normal case with flags and command",
			args:        []string{"-i", "2", "--", "ls", "-la"},
			wantFlags:   []string{"-i", "2"},
			wantCommand: []string{"ls", "-la"},
		},
		{
			name:        "no flags before separator",
			args:        []string{"--", "echo", "hello"},
			wantFlags:   []string{},
			wantCommand: []string{"echo", "hello"},
		},
		{
			name:        "no separator",
			args:        []string{"-i", "2"},
			wantFlags:   []string{"-i", "2"},
			wantCommand: []string{},
		},
		{
			name:        "empty args",
			args:        []string{},
			wantFlags:   []string{},
			wantCommand: []string{},
		},
		{
			name:        "separator at end",
			args:        []string{"-i", "2", "--"},
			wantFlags:   []string{"-i", "2"},
			wantCommand: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotFlags, gotCommand := parseWatchArgs(tt.args)

			if len(gotFlags) != len(tt.wantFlags) {
				t.Errorf("flags length = %d, want %d", len(gotFlags), len(tt.wantFlags))
			}
			for i := range tt.wantFlags {
				if i >= len(gotFlags) {
					break
				}
				if gotFlags[i] != tt.wantFlags[i] {
					t.Errorf("flags[%d] = %q, want %q", i, gotFlags[i], tt.wantFlags[i])
				}
			}

			if len(gotCommand) != len(tt.wantCommand) {
				t.Errorf("command length = %d, want %d", len(gotCommand), len(tt.wantCommand))
			}
			for i := range tt.wantCommand {
				if i >= len(gotCommand) {
					break
				}
				if gotCommand[i] != tt.wantCommand[i] {
					t.Errorf("command[%d] = %q, want %q", i, gotCommand[i], tt.wantCommand[i])
				}
			}
		})
	}
}
