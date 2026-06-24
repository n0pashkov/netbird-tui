package client

import (
	"context"
	"os/exec"
)

var runNetbirdDebug = func(ctx context.Context, args ...string) ([]byte, error) {
	cmdArgs := append([]string{"debug"}, args...)
	return exec.CommandContext(ctx, "netbird", cmdArgs...).CombinedOutput()
}

func RunDebugCommand(ctx context.Context, args ...string) (string, error) {
	out, err := runNetbirdDebug(ctx, args...)
	return string(out), err
}
