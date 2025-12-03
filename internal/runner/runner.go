package runner

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/anthropics/maestro-ios-device/internal/maestro"
)

const (
	DevicePort     = uint16(22087)
	buildTimeout   = 10 * time.Minute
	startupTimeout = 90 * time.Second
)

type Runner struct {
	deviceUDID string
	teamID     string
	buildDir   string
	cmd        *exec.Cmd
	logFile    *os.File
}

func New(deviceUDID, teamID string) *Runner {
	return &Runner{
		deviceUDID: deviceUDID,
		teamID:     teamID,
	}
}

func (r *Runner) Build(ctx context.Context) error {
	runnerPath, err := maestro.GetRunnerPath()
	if err != nil {
		return err
	}

	r.buildDir, err = os.MkdirTemp("", "maestro-build-*")
	if err != nil {
		return err
	}

	os.MkdirAll(filepath.Join(r.buildDir, "logs"), 0755)

	logPath := filepath.Join(r.buildDir, "logs", "build.log")
	logFile, err := os.Create(logPath)
	if err != nil {
		return err
	}
	defer logFile.Close()

	buildCtx, cancel := context.WithTimeout(ctx, buildTimeout)
	defer cancel()

	cmd := exec.CommandContext(buildCtx, "xcodebuild",
		"build-for-testing",
		"-project", filepath.Join(runnerPath, "maestro-driver-ios.xcodeproj"),
		"-scheme", "maestro-driver-ios",
		"-destination", r.destination(),
		"-derivedDataPath", r.buildOut(),
		fmt.Sprintf("DEVELOPMENT_TEAM=%s", r.teamID),
	)
	cmd.Stdout = logFile
	cmd.Stderr = logFile

	fmt.Println("🔨 Building (up to 10 min)...")

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("build failed (see %s)", logPath)
	}

	if _, err := r.findXctestrun(); err != nil {
		return err
	}

	fmt.Println("✅ Build complete")
	return nil
}

func (r *Runner) buildOut() string {
	return filepath.Join(r.buildDir, "build")
}

func (r *Runner) destination() string {
	return fmt.Sprintf("id=%s", r.deviceUDID)
}

func (r *Runner) findXctestrun() (string, error) {
	pattern := filepath.Join(r.buildOut(), "Build", "Products", "*.xctestrun")
	matches, _ := filepath.Glob(pattern)
	if len(matches) == 0 {
		return "", fmt.Errorf("no xctestrun file found")
	}
	return matches[0], nil
}

func (r *Runner) Start(ctx context.Context) error {
	xctestrun, err := r.findXctestrun()
	if err != nil {
		return err
	}

	logPath := filepath.Join(r.buildDir, "logs", "runner.log")
	r.logFile, err = os.Create(logPath)
	if err != nil {
		return fmt.Errorf("failed to create log file: %w", err)
	}

	r.cmd = exec.CommandContext(ctx, "xcodebuild",
		"test-without-building",
		"-xctestrun", xctestrun,
		"-destination", r.destination(),
		"-derivedDataPath", r.buildOut(),
	)
	r.cmd.Stdout = r.logFile
	r.cmd.Stderr = r.logFile

	fmt.Println("▶️  Starting runner...")

	if err := r.cmd.Start(); err != nil {
		return err
	}

	if err := r.waitForStartup(logPath); err != nil {
		r.Stop()
		return err
	}

	fmt.Println("✅ Runner started")
	return nil
}

func (r *Runner) Stop() {
	if r.cmd != nil && r.cmd.Process != nil {
		r.cmd.Process.Kill()
	}
	if r.logFile != nil {
		r.logFile.Close()
	}
}

func (r *Runner) Cleanup() {
	r.Stop()
	if r.buildDir != "" {
		os.RemoveAll(r.buildDir)
	}
}

func (r *Runner) waitForStartup(logPath string) error {
	timeout := time.After(startupTimeout)
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			content, err := os.ReadFile(logPath)
			if err != nil {
				continue
			}
			if err := checkLog(string(content), logPath); err != errNotReady {
				return err
			}
		case <-timeout:
			return fmt.Errorf("startup timeout (see %s)", logPath)
		}
	}
}

var errNotReady = fmt.Errorf("not ready")

func checkLog(log, logPath string) error {
	// Success indicators
	if strings.Contains(log, "started") {
		if strings.Contains(log, "Test Suite") || strings.Contains(log, "maestro-driver-ios") {
			return nil
		}
	}
	// Known errors
	if strings.Contains(log, "Developer App Certificate is not trusted") {
		return fmt.Errorf("certificate not trusted - trust it in Settings > General > VPN & Device Management")
	}
	if strings.Contains(log, "Testing failed:") {
		return fmt.Errorf("runner failed (see %s)", logPath)
	}
	return errNotReady
}
