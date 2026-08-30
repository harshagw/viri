package main

import (
	"os"
	"regexp"
	"strings"
	"testing"
)

// The playground's snippets ship with the site, so a snippet that no longer
// compiles is a broken front page. This reads them out of the TypeScript
// source rather than duplicating them, so the two cannot drift.
func TestSiteSnippetsRun(t *testing.T) {
	source, err := os.ReadFile("../../viri-web/lib/snippets.ts")
	if err != nil {
		t.Skipf("site sources not present: %v", err)
	}

	// Each snippet is a backtick-quoted `code:` field, plus DEFAULT_CODE.
	pattern := regexp.MustCompile("(?s)(?:code:|DEFAULT_CODE =)\\s*`(.*?)`")
	matches := pattern.FindAllStringSubmatch(string(source), -1)
	if len(matches) == 0 {
		t.Fatal("found no snippets to check; has snippets.ts changed shape?")
	}

	titles := regexp.MustCompile(`title: "([^"]+)"`).FindAllStringSubmatch(string(source), -1)

	for i, m := range matches {
		name := "DEFAULT_CODE"
		if i < len(titles) {
			name = titles[i][1]
		}
		t.Run(name, func(t *testing.T) {
			_, out, errs, _ := run(m[1])
			if len(errs) > 0 {
				t.Errorf("snippet does not run:\n  %s\n\nsource:\n%s",
					strings.Join(errs, "\n  "), m[1])
			}
			if out == "" {
				t.Errorf("snippet produced no output; it should show something")
			}
		})
	}
}
