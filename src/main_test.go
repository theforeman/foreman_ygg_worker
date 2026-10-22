package main

import (
	"os"
	"testing"
)

func TestDetermineWorkdir(t *testing.T) {
	tests := []struct {
		description                string
		foremanYggWorkerWorkdir    string
		xdgRuntimeDir              string
		setForemanYggWorkerWorkdir bool
		setXdgRuntimeDir           bool
		want                       string
	}{
		{
			description:                "FOREMAN_YGG_WORKER_WORKDIR set",
			foremanYggWorkerWorkdir:    "/custom/workdir",
			setForemanYggWorkerWorkdir: true,
			want:                       "/custom/workdir",
		},
		{
			description:      "Only XDG_RUNTIME_DIR set",
			xdgRuntimeDir:    "/run/user/1000",
			setXdgRuntimeDir: true,
			want:             "/run/user/1000",
		},
		{
			description: "Neither set",
			want:        "/run",
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			if test.setForemanYggWorkerWorkdir {
				t.Setenv("FOREMAN_YGG_WORKER_WORKDIR", test.foremanYggWorkerWorkdir)
			} else {
				t.Setenv("FOREMAN_YGG_WORKER_WORKDIR", "")
				_ = os.Unsetenv("FOREMAN_YGG_WORKER_WORKDIR")
			}
			if test.setXdgRuntimeDir {
				t.Setenv("XDG_RUNTIME_DIR", test.xdgRuntimeDir)
			} else {
				t.Setenv("XDG_RUNTIME_DIR", "")
				_ = os.Unsetenv("XDG_RUNTIME_DIR")
			}

			got := determineWorkdir()
			if got != test.want {
				t.Errorf("|%+v| != |%+v|", got, test.want)
			}
		})
	}
}
