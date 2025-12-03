package maestro

import (
	"archive/zip"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

const (
	releaseURL = "https://github.com/devicelab-dev/maestro-ios-device/releases/latest/download"
	jarsZip    = "maestro-jars.zip"
	runnerZip  = "maestro-ios-runner.zip"
	runnerDir  = ".maestro/maestro-ios-xctest-runner"
)

var supportedVersions = []string{"2.0.9", "2.0.10"}

func isSupportedVersion(version string) bool {
	for _, v := range supportedVersions {
		if version == v {
			return true
		}
	}
	return false
}

func runnerBasePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, runnerDir), nil
}

func RunSetup() error {
	fmt.Println("🔧 Maestro iOS Device Setup")
	fmt.Println()

	// Check Maestro installed
	installed, version, err := checkInstalled()
	if err != nil {
		return err
	}
	if !installed {
		return fmt.Errorf("Maestro not found. Install from: https://maestro.mobile.dev/")
	}

	fmt.Printf("Detected Maestro: %s\n", version)

	if !isSupportedVersion(version) {
		return fmt.Errorf("Unsupported Maestro version: %s\n\nSupported versions: 2.0.9, 2.0.10\n\nPlease upgrade or downgrade Maestro to a supported version.", version)
	}

	libPath, err := getLibPath()
	if err != nil {
		return fmt.Errorf("failed to find Maestro lib: %w", err)
	}

	runnerPath, err := runnerBasePath()
	if err != nil {
		return err
	}

	fmt.Println()
	fmt.Println("📦 Backing up existing JARs...")
	if err := backupJars(libPath); err != nil {
		return fmt.Errorf("failed to backup JARs: %w", err)
	}
	fmt.Println("✅ Backup complete")

	fmt.Println("📥 Downloading JARs...")
	if err := downloadAndExtract(releaseURL+"/"+jarsZip, libPath); err != nil {
		return fmt.Errorf("failed to download JARs: %w", err)
	}
	fmt.Println("✅ JARs installed")

	fmt.Println("📥 Downloading iOS runner...")
	if err := os.MkdirAll(runnerPath, 0755); err != nil {
		return err
	}
	if err := downloadAndExtract(releaseURL+"/"+runnerZip, runnerPath); err != nil {
		return fmt.Errorf("failed to download runner: %w", err)
	}
	fmt.Println("✅ iOS runner installed")

	fmt.Println()
	fmt.Println("✅ Setup complete!")
	fmt.Println()
	fmt.Println("Run tests with:")
	fmt.Println("  maestro-ios-device --team-id <ID> --device <UDID>")

	return nil
}

func IsPatched() (bool, error) {
	out, err := exec.Command("maestro", "--help").Output()
	if err != nil {
		return false, err
	}
	return strings.Contains(string(out), "driver-host-port"), nil
}

func GetRunnerPath() (string, error) {
	basePath, err := runnerBasePath()
	if err != nil {
		return "", err
	}

	path := filepath.Join(basePath, "maestro-driver-ios.xcodeproj")
	if _, err := os.Stat(path); err != nil {
		return "", fmt.Errorf("iOS runner not found. Run: maestro-ios-device setup")
	}

	return basePath, nil
}

func checkInstalled() (bool, string, error) {
	out, err := exec.Command("maestro", "--version").Output()
	if err != nil {
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "executable file not found") {
			return false, "", nil
		}
		return false, "", err
	}
	return true, parseVersion(string(out)), nil
}

func parseVersion(out string) string {
	out = strings.TrimSpace(out)

	for _, prefix := range []string{"cli version: ", "version: ", "CLI "} {
		if idx := strings.Index(strings.ToLower(out), strings.ToLower(prefix)); idx != -1 {
			rest := out[idx+len(prefix):]
			if fields := strings.Fields(rest); len(fields) > 0 {
				return fields[0]
			}
		}
	}

	if m := regexp.MustCompile(`\d+\.\d+\.\d+`).FindString(out); m != "" {
		return m
	}

	return out
}

func getLibPath() (string, error) {
	scriptPath, err := exec.Command("which", "maestro").Output()
	if err != nil {
		return "", fmt.Errorf("maestro not in PATH")
	}

	script := strings.TrimSpace(string(scriptPath))
	if resolved, err := filepath.EvalSymlinks(script); err == nil {
		script = resolved
	}

	if content, err := os.ReadFile(script); err == nil {
		if libPath := findLibInScript(string(content), filepath.Dir(script)); libPath != "" {
			return libPath, nil
		}
	}

	// Fallback 1: lib is sibling to bin
	binDir := filepath.Dir(script)
	libPath := filepath.Join(filepath.Dir(binDir), "lib")
	if _, err := os.Stat(libPath); err == nil {
		return libPath, nil
	}

	// Fallback 2: ~/.maestro/lib
	if home, err := os.UserHomeDir(); err == nil {
		libPath = filepath.Join(home, ".maestro", "lib")
		if _, err := os.Stat(libPath); err == nil {
			return libPath, nil
		}
	}

	return "", fmt.Errorf("lib directory not found")
}

func findLibInScript(content, scriptDir string) string {
	for _, line := range strings.Split(content, "\n") {
		if strings.Contains(line, "CLASSPATH") && strings.Contains(line, "/lib/") {
			if idx := strings.Index(line, "/lib/"); idx > 0 {
				start := strings.LastIndexAny(line[:idx], "=\"'") + 1
				pathPart := line[start : idx+4]

				if strings.HasPrefix(pathPart, "$") {
					return filepath.Join(filepath.Dir(scriptDir), "lib")
				}
				if filepath.IsAbs(pathPart) {
					return pathPart
				}
			}
		}
	}
	return ""
}

func backupJars(libPath string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	backupDir := filepath.Join(home, ".maestro", "backup")
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return err
	}

	entries, err := os.ReadDir(libPath)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		name := entry.Name()
		if !strings.HasPrefix(name, "maestro") || !strings.HasSuffix(name, ".jar") {
			continue
		}
		src := filepath.Join(libPath, name)
		dst := filepath.Join(backupDir, name)
		if err := copyFile(src, dst); err != nil {
			return fmt.Errorf("failed to backup %s: %w", name, err)
		}
	}
	return nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}

func downloadAndExtract(url, dest string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("download failed: %s", resp.Status)
	}

	tmp, err := os.CreateTemp("", "maestro-*.zip")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)

	if _, err := io.Copy(tmp, resp.Body); err != nil {
		tmp.Close()
		return err
	}
	tmp.Close()

	return unzip(tmpName, dest)
}

func unzip(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		// Security: prevent zip slip
		target := filepath.Join(dest, f.Name)
		if !strings.HasPrefix(filepath.Clean(target), filepath.Clean(dest)+string(os.PathSeparator)) {
			return fmt.Errorf("invalid file path: %s", f.Name)
		}

		if f.FileInfo().IsDir() {
			os.MkdirAll(target, 0755)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
			return err
		}

		if err := extractFile(f, target); err != nil {
			return err
		}
	}

	return nil
}

func extractFile(f *zip.File, target string) error {
	rc, err := f.Open()
	if err != nil {
		return err
	}
	defer rc.Close()

	out, err := os.OpenFile(target, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, rc)
	return err
}

