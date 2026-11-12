// Package repocheck keeps the workspace guidance honest: every top-level
// directory must be explained in AGENTS.md and every Go package must have a
// package comment.
package repocheck

import (
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const root = "../.."

func TestEveryTopLevelDirIsInAgentsGuide(t *testing.T) {
	guide, err := os.ReadFile(filepath.Join(root, "AGENTS.md"))
	if err != nil {
		t.Fatal(err)
	}
	entries, _ := os.ReadDir(root)
	for _, e := range entries {
		if !e.IsDir() || strings.HasPrefix(e.Name(), ".") {
			continue
		}
		if !strings.Contains(string(guide), "`"+e.Name()+"/`") {
			t.Errorf("AGENTS.md does not explain top-level directory %s/", e.Name())
		}
	}
}

func TestEveryPackageHasADocComment(t *testing.T) {
	seen := map[string]bool{}
	filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			if path != root && strings.HasPrefix(d.Name(), ".") {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") {
			return nil
		}
		dir := filepath.Dir(path)
		f, perr := parser.ParseFile(token.NewFileSet(), path, nil, parser.PackageClauseOnly|parser.ParseComments)
		if perr != nil {
			t.Errorf("%s: %v", path, perr)
			return nil
		}
		if _, ok := seen[dir]; !ok {
			seen[dir] = false
		}
		if f.Doc != nil {
			seen[dir] = true
		}
		return nil
	})
	if len(seen) < 10 {
		t.Fatalf("only %d packages found: the walk is broken", len(seen))
	}
	for dir, ok := range seen {
		if !ok {
			t.Errorf("package in %s has no package comment", strings.TrimPrefix(dir, root+"/"))
		}
	}
}
