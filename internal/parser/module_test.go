package parser

import (
	"path/filepath"
	"testing"
)

func TestResolveModulePath(t *testing.T) {
	tests := []struct {
		name       string
		baseDir    string
		importPath string
		output string
	}{
		{"relative same dir", "/abs/path", "mod.viri", "abs/path/mod.viri"},
		{"relative sub dir", "/abs/path", "sub/mod.viri", "abs/path/sub/mod.viri"},
		{"parent dir", "/abs/path/sub", "../mod.viri", "abs/path/mod.viri"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ResolveModulePath(tt.baseDir, tt.importPath)
			if err != nil {
				t.Fatalf("ResolveModulePath() error = %v", err)
			}
			want := tt.output
			if !filepath.IsAbs(got) {
				t.Errorf("expected absolute path, got %s", got)
			}
			if filepath.Base(got) != filepath.Base(want) {
				t.Errorf("got base %s, want base %s", filepath.Base(got), filepath.Base(want))
			}
		})
	}
}
