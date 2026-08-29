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

// TestE2E_Diff runs every test program through both engines and requires
// their outputs to be identical
func TestE2E_Diff(t *testing.T) {
	testDataDir := "testdata"
	files, err := os.ReadDir(testDataDir)
	if err != nil {
		t.Fatalf("failed to read testdata dir: %v", err)
	}

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

			interpOut := runEngine(t, viriPath, "interpreter", path)
			vmOut := runEngine(t, viriPath, "vm", path)

			if interpOut != vmOut {
				t.Errorf("engines disagree\ninterpreter:\n%s\nvm:\n%s", interpOut, vmOut)
			}
		})
	}
}

func runEngine(t *testing.T, viriPath, engine, scriptPath string) string {
	t.Helper()
	cmd := exec.Command(viriPath, "--no-warning", "--engine="+engine, scriptPath)
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	_ = cmd.Run() // error cases are part of the comparison

	return strings.TrimSpace(out.String() + "\n" + stderr.String())
}
