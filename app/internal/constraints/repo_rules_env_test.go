package constraints

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestOnlyConfigReadsEnvironmentVariables(t *testing.T) {
	repoRoot := findRepoRoot(t)
	allowed := filepath.Join("internal", "config", "config.go")
	var violations []string

	err := filepath.WalkDir(repoRoot, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() && shouldSkipDir(d.Name()) {
			return filepath.SkipDir
		}
		if d.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}

		relPath, err := filepath.Rel(repoRoot, path)
		if err != nil {
			return err
		}
		relPath = filepath.ToSlash(relPath)
		if relPath == allowed {
			return nil
		}

		file, fset, err := parseFile(path)
		if err != nil {
			return err
		}
		for _, violation := range envReadViolations(relPath, file, fset) {
			violations = append(violations, violation)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk repo: %v", err)
	}
	if len(violations) != 0 {
		failWithGuidance(
			t,
			"environment reads must stay in internal/config/config.go",
			"move env parsing to internal/config/config.go and pass typed config downward instead of calling os.Getenv or os.LookupEnv from feature packages.",
			violations,
		)
	}
}
