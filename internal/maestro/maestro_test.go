package maestro

import (
	"archive/zip"
	"os"
	"path/filepath"
	"testing"
)

func TestParseVersion(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"cli version: 2.0.10", "2.0.10"},
		{"CLI version: 2.0.10", "2.0.10"},
		{"version: 1.39.0", "1.39.0"},
		{"Maestro CLI 2.0.10", "2.0.10"},
		{"2.0.10", "2.0.10"},
		{"  cli version: 2.0.10  ", "2.0.10"},
		{"random text 3.1.4 more text", "3.1.4"},
		{"no version here", "no version here"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := parseVersion(tt.input)
			if got != tt.want {
				t.Errorf("parseVersion(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestUnzip(t *testing.T) {
	// Create temp zip file
	tmpZip, err := os.CreateTemp("", "test-*.zip")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpZip.Name())

	// Create zip with test file
	zw := zip.NewWriter(tmpZip)
	fw, err := zw.Create("test.txt")
	if err != nil {
		t.Fatal(err)
	}
	fw.Write([]byte("hello"))
	zw.Close()
	tmpZip.Close()

	// Create temp dest dir
	destDir, err := os.MkdirTemp("", "test-dest-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(destDir)

	// Unzip
	if err := unzip(tmpZip.Name(), destDir); err != nil {
		t.Fatalf("unzip failed: %v", err)
	}

	// Verify file exists
	content, err := os.ReadFile(filepath.Join(destDir, "test.txt"))
	if err != nil {
		t.Fatalf("failed to read extracted file: %v", err)
	}
	if string(content) != "hello" {
		t.Errorf("got %q, want %q", content, "hello")
	}
}

func TestUnzip_ZipSlip(t *testing.T) {
	// Create temp zip with malicious path
	tmpZip, err := os.CreateTemp("", "test-*.zip")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpZip.Name())

	zw := zip.NewWriter(tmpZip)
	// Try to escape with ../
	fw, err := zw.Create("../../../etc/malicious.txt")
	if err != nil {
		t.Fatal(err)
	}
	fw.Write([]byte("bad"))
	zw.Close()
	tmpZip.Close()

	destDir, err := os.MkdirTemp("", "test-dest-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(destDir)

	// Should fail with zip slip protection
	err = unzip(tmpZip.Name(), destDir)
	if err == nil {
		t.Error("expected error for zip slip attack")
	}
}

func TestUnzip_WithDir(t *testing.T) {
	tmpZip, err := os.CreateTemp("", "test-*.zip")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpZip.Name())

	zw := zip.NewWriter(tmpZip)
	// Create directory entry
	zw.Create("subdir/")
	// Create file in directory
	fw, _ := zw.Create("subdir/file.txt")
	fw.Write([]byte("nested"))
	zw.Close()
	tmpZip.Close()

	destDir, err := os.MkdirTemp("", "test-dest-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(destDir)

	if err := unzip(tmpZip.Name(), destDir); err != nil {
		t.Fatalf("unzip failed: %v", err)
	}

	content, err := os.ReadFile(filepath.Join(destDir, "subdir", "file.txt"))
	if err != nil {
		t.Fatalf("failed to read nested file: %v", err)
	}
	if string(content) != "nested" {
		t.Errorf("got %q, want %q", content, "nested")
	}
}

func TestFindLibInScript(t *testing.T) {
	tests := []struct {
		name      string
		content   string
		scriptDir string
		want      string
	}{
		{
			name:      "absolute path",
			content:   `CLASSPATH="/opt/maestro/lib/maestro.jar"`,
			scriptDir: "/usr/bin",
			want:      "/opt/maestro/lib",
		},
		{
			name:      "variable path",
			content:   `CLASSPATH="$APP_HOME/lib/maestro.jar"`,
			scriptDir: "/opt/maestro/bin",
			want:      "/opt/maestro/lib",
		},
		{
			name:      "no classpath",
			content:   `echo "hello"`,
			scriptDir: "/usr/bin",
			want:      "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := findLibInScript(tt.content, tt.scriptDir)
			if got != tt.want {
				t.Errorf("findLibInScript() = %q, want %q", got, tt.want)
			}
		})
	}
}
