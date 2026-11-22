package shell

// Minimal silent-exec core extracted from the FeralOS utils/shell package
// (github.com/skea999/feralos, ISC). Only what feralos-dbus-audit uses:
// ExecSilentArgs (chroot-aware, never logs, never retries/prompting —
// boot-oneshot safety) and SetCaller.

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"sync"
	"time"
)

// commandTimeout bounds every single command execution. On expiry the
// process is killed and the timeout-wrapped error flows back to the caller.
const commandTimeout = 30 * time.Minute

var (
	currentCaller string
	callerMu      sync.RWMutex
)

// SetCaller sets the module label for subsequent log lines.
func SetCaller(caller string) {
	callerMu.Lock()
	currentCaller = caller
	callerMu.Unlock()
}

// maxSilentStderr bounds the stderr snippet attached to ExecSilent errors.
const maxSilentStderr = 500

// silentRun is the shared core of the silent flavors: builds the command with
// the package commandTimeout context, captures + trims stdout, attaches a
// trimmed+bounded stderr snippet to errors, logs nothing, and never retries /
// prompts (boot-oneshot safety).
func silentRun(buildCmd func(ctx context.Context) *exec.Cmd) (string, error) {
	var stdout, stderr bytes.Buffer
	ctx, cancel := context.WithTimeout(context.Background(), commandTimeout)
	defer cancel()
	cmd := buildCmd(ctx)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		errMsg := strings.TrimSpace(stderr.String())
		if len(errMsg) > maxSilentStderr {
			errMsg = errMsg[:maxSilentStderr]
		}
		callerMu.RLock()
		caller := currentCaller
		callerMu.RUnlock()
		if errMsg != "" {
			if caller != "" {
				err = fmt.Errorf("%w: %s: %s", err, caller, errMsg)
			} else {
				err = fmt.Errorf("%w: %s", err, errMsg)
			}
		}
		return strings.TrimSpace(stdout.String()), err
	}
	return strings.TrimSpace(stdout.String()), nil
}

// ExecSilentArgs runs name with args (optionally chrooted to dir when dir is
// non-empty), captures stdout, and attaches a bounded stderr snippet to
// errors. Logs nothing, never retries — boot-oneshot safety.
func ExecSilentArgs(dir string, name string, args ...string) (string, error) {
	return silentRun(func(ctx context.Context) *exec.Cmd {
		if dir != "" {
			return exec.CommandContext(ctx, "chroot", append([]string{dir, name}, args...)...) //nolint:gosec // G204: same posture as ExecuteCmdArgs — silent path, args logged-free
		}
		return exec.CommandContext(ctx, name, args...)
	})
}
