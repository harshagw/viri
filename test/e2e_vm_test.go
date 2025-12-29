//go:build e2e

package test

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Files that the VM engine supports
var vmSupportedFiles = []string{
	"arithmetic.viri",
	"conditionals.viri",
	"loops.viri",
	"functions.viri",
	"anonymous_functions.viri",
	"block_scope.viri",
	"classes.viri",
	"classes_inheritance.viri",
	"classes_advanced.viri",
}

func TestE2E_VM(t *testing.T) {
	testDataDir := "testdata"

	// Ensure the binary is built and available in the root
	viriPath, err := filepath.Abs("../viri")
	if err != nil {
		t.Fatalf("failed to get absolute path for viri: %v", err)
	}

	for _, fileName := range vmSupportedFiles {
		t.Run(fileName, func(t *testing.T) {
			path := filepath.Join(testDataDir, fileName)

			// Check if the file exists
			if _, err := os.Stat(path); os.IsNotExist(err) {
				t.Skipf("test file %s does not exist", fileName)
			}

			// Expected output is in a .out file with same name
			expectedPath := strings.TrimSuffix(path, ".viri") + ".out"
			expectedOutput, err := os.ReadFile(expectedPath)
			if err != nil {
				t.Logf("Warning: no .out file for %s", fileName)
			}

			// Run the binary with VM engine
			output := runViriBinaryVM(t, viriPath, path)

			if expectedOutput != nil {
				if strings.TrimSpace(output) != strings.TrimSpace(string(expectedOutput)) {
					t.Errorf("output mismatch\ngot:\n%s\nwant:\n%s", output, string(expectedOutput))
				}
			}
		})
	}
}

func runViriBinaryVM(t *testing.T, viriPath, scriptPath string) string {
	cmd := exec.Command(viriPath, "--engine=vm", scriptPath)
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
