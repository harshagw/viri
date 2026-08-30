//go:build e2e

package test

import (
	"bytes"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// The compiler suite is split by what a program is supposed to do, because the
// three outcomes have different contracts:
//
// Within valid/ and invalid/, cases are grouped by language feature
// (classes/, collections/, control_flow/, functions/, imports/, operators/,
// strings/, types/, variables/, stdlib/) so a failure names the feature.
//
//	valid/    compiles and runs to completion. exit 0, stdout matches .out.
//	runtime/  compiles, then fails while running. non-zero exit, stdout matches
//	          .out (output produced before the failure), stderr matches .err.
//	invalid/  rejected before execution. non-zero exit, stderr matches .err,
//	          and stdout is empty — nothing ran.
//	modules/  imported by other cases, never executed directly.
//
// The empty-stdout assertion on invalid/ is the one that matters most: it is
// the end-to-end proof that an ill-typed program produces no effects.

type result struct {
	stdout string
	stderr string
	code   int
}

func runViri(t *testing.T, script string) result {
	t.Helper()

	bin, err := filepath.Abs("../viri")
	if err != nil {
		t.Fatalf("resolve viri path: %v", err)
	}
	if _, err := os.Stat(bin); err != nil {
		t.Fatalf("viri binary not built at %s: run 'make build' or 'make e2e'", bin)
	}

	cmd := exec.Command(bin, "--no-warning", "--engine=vm", script)
	var out, errBuf bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errBuf

	code := 0
	if err := cmd.Run(); err != nil {
		exit, ok := err.(*exec.ExitError)
		if !ok {
			t.Fatalf("running %s: %v", script, err)
		}
		code = exit.ExitCode()
	}
	return result{stdout: out.String(), stderr: errBuf.String(), code: code}
}

// golden reads an expected-output file. A case without one is a failure, not a
// skip: a missing .out used to make the test pass silently, so any program
// added without expectations looked like coverage it was not.
func golden(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("missing expectation file %s: every case needs one", path)
	}
	return string(data)
}

func equal(t *testing.T, what, got, want string) {
	t.Helper()
	if strings.TrimSpace(got) != strings.TrimSpace(want) {
		t.Errorf("%s mismatch\ngot:\n%s\nwant:\n%s", what, got, want)
	}
}

// cases lists every .viri file under dir, recursing into the feature
// subdirectories. The returned paths are relative to dir, so a subtest is named
// "classes/polymorphism.viri" and a failure says which feature broke.
func cases(t *testing.T, dir string) []string {
	t.Helper()
	var found []string
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(d.Name(), ".viri") {
			rel, err := filepath.Rel(dir, path)
			if err != nil {
				return err
			}
			found = append(found, rel)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk %s: %v", dir, err)
	}
	if len(found) == 0 {
		t.Fatalf("no cases in %s", dir)
	}
	sort.Strings(found)
	return found
}

func TestE2E_VM_Valid(t *testing.T) {
	dir := filepath.Join("testdata", "valid")
	for _, name := range cases(t, dir) {
		t.Run(name, func(t *testing.T) {
			path := filepath.Join(dir, name)
			got := runViri(t, path)

			if got.code != 0 {
				t.Fatalf("expected success, got exit %d\nstderr:\n%s", got.code, got.stderr)
			}
			if strings.TrimSpace(got.stderr) != "" {
				t.Errorf("expected no diagnostics, got:\n%s", got.stderr)
			}
			equal(t, "stdout", got.stdout, golden(t, strings.TrimSuffix(path, ".viri")+".out"))
		})
	}
}

func TestE2E_VM_Runtime(t *testing.T) {
	dir := filepath.Join("testdata", "runtime")
	for _, name := range cases(t, dir) {
		t.Run(name, func(t *testing.T) {
			path := filepath.Join(dir, name)
			base := strings.TrimSuffix(path, ".viri")
			got := runViri(t, path)

			if got.code == 0 {
				t.Fatalf("expected a runtime error, got exit 0\nstdout:\n%s", got.stdout)
			}
			equal(t, "stdout", got.stdout, golden(t, base+".out"))
			equal(t, "stderr", got.stderr, golden(t, base+".err"))
		})
	}
}

func TestE2E_VM_Invalid(t *testing.T) {
	dir := filepath.Join("testdata", "invalid")
	for _, name := range cases(t, dir) {
		t.Run(name, func(t *testing.T) {
			path := filepath.Join(dir, name)
			got := runViri(t, path)

			if got.code == 0 {
				t.Fatalf("expected rejection, got exit 0\nstdout:\n%s", got.stdout)
			}
			if got.stdout != "" {
				t.Errorf("a rejected program must not run, but it wrote to stdout:\n%s", got.stdout)
			}
			equal(t, "stderr", got.stderr, golden(t, strings.TrimSuffix(path, ".viri")+".err"))
		})
	}
}
