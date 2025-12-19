package test

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestE2E(t *testing.T) {
	testDataDir := "testdata"
	files, err := os.ReadDir(testDataDir)
	if err != nil {
		t.Fatalf("failed to read testdata dir: %v", err)
	}

	// Ensure the binary is built and available in the root
	// We assume 'viri' exists in the root directory relative to the project root
	viriPath, err := filepath.Abs("../viri")
	if err != nil {
		t.Fatalf("failed to get absolute path for viri: %v", err)
	}

	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".viri") {
			continue
		}

		// Skip module files that are just for importing
		if strings.Contains(file.Name(), "module_") {
			continue
		}

		t.Run(file.Name(), func(t *testing.T) {
			path := filepath.Join(testDataDir, file.Name())

			// Expected output is usually in a .out file with same name
			expectedPath := strings.TrimSuffix(path, ".viri") + ".out"
			expectedOutput, err := os.ReadFile(expectedPath)
			if err != nil {
				t.Logf("Warning: no .out file for %s", file.Name())
			}

			// Run the binary
			output := runViriBinary(t, viriPath, path)

			if expectedOutput != nil {
				if strings.TrimSpace(output) != strings.TrimSpace(string(expectedOutput)) {
					t.Errorf("output mismatch\ngot:\n%s\nwant:\n%s", output, string(expectedOutput))
				}
			}
		})
	}
}

func runViriBinary(t *testing.T, viriPath, scriptPath string) string {
	cmd := exec.Command(viriPath, scriptPath)
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		// Some tests might expect failure (e.g. runtime errors)
		// We combine stdout/stderr for these tests
		return strings.TrimSpace(out.String() + "\n" + stderr.String())
	}

	return out.String()
}
