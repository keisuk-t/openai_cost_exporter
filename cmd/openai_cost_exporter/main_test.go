package main

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"
)

func TestFlags(t *testing.T) {
	for _, tc := range []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{name: "version", args: []string{"--version"}},
		{name: "help", args: []string{"--help"}},
		{name: "unknown flag", args: []string{"--unknown"}, wantErr: true},
		{name: "positional argument", args: []string{"unexpected"}, wantErr: true},
		{name: "empty address", args: []string{"--web.listen-address="}, wantErr: true},
		{name: "invalid address", args: []string{"--web.listen-address=not-an-address"}, wantErr: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			err := run(context.Background(), tc.args, &stdout, &stderr)
			if (err != nil) != tc.wantErr {
				t.Fatalf("run error = %v, wantErr = %v", err, tc.wantErr)
			}
			if tc.name == "version" && strings.TrimSpace(stdout.String()) != version {
				t.Fatalf("version output = %q", stdout.String())
			}
			if tc.name == "help" && !strings.Contains(stderr.String(), "web.listen-address") {
				t.Fatalf("help output = %q", stderr.String())
			}
		})
	}
}

func TestRunShutsDownOnCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := run(ctx, []string{"--web.listen-address=127.0.0.1:0"}, io.Discard, io.Discard); err != nil {
		t.Fatalf("run: %v", err)
	}
}
