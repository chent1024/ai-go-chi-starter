package constraints

import (
	"go/ast"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestOnlyRuntimeLoggingConstructsSlogHandlers(t *testing.T) {
	repoRoot := findRepoRoot(t)
	allowedPrefix := filepath.ToSlash(filepath.Join("internal", "runtime", "logging")) + "/"
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
		if strings.HasPrefix(relPath, allowedPrefix) {
			return nil
		}

		file, fset, err := parseFile(path)
		if err != nil {
			return err
		}
		if len(importedAliases(file, "log/slog")) == 0 {
			return nil
		}
		for _, violation := range selectorCallViolations(
			relPath,
			file,
			fset,
			"log/slog",
			map[string]struct{}{
				"New":            {},
				"NewJSONHandler": {},
				"NewTextHandler": {},
			},
		) {
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
			"only internal/runtime/logging may construct slog loggers and handlers in production code",
			"use runtime/logging helpers such as NewLogger, NewBootstrapLogger, or NewNoopLogger instead of constructing slog handlers directly in feature packages.",
			violations,
		)
	}
}

func TestOnlyRuntimeLoggingDefinesLogFieldConstants(t *testing.T) {
	repoRoot := findRepoRoot(t)
	allowedPrefixes := []string{
		filepath.ToSlash(filepath.Join("internal", "runtime", "logging")) + "/",
		filepath.ToSlash(filepath.Join("internal", "runtime", "tracing")) + "/",
	}
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
		if hasAnyPrefix(relPath, allowedPrefixes) {
			return nil
		}

		file, fset, err := parseFile(path)
		if err != nil {
			return err
		}
		for _, decl := range file.Decls {
			genDecl, ok := decl.(*ast.GenDecl)
			if !ok || genDecl.Tok != token.CONST {
				continue
			}
			for _, spec := range genDecl.Specs {
				valueSpec, ok := spec.(*ast.ValueSpec)
				if !ok {
					continue
				}
				for _, name := range valueSpec.Names {
					if !strings.HasPrefix(name.Name, "LogField") {
						continue
					}
					position := fset.Position(name.Pos())
					violations = append(violations, formatViolation(relPath, position.Line, name.Name))
				}
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk repo: %v", err)
	}
	if len(violations) != 0 {
		failWithGuidance(
			t,
			"log field constants must only be declared in runtime observability packages",
			"reuse runtime/logging or runtime/tracing field constants instead of redefining LogField* names in transport, service, infra, or worker packages.",
			violations,
		)
	}
}

func TestRuntimeRootHasNoGoFiles(t *testing.T) {
	repoRoot := findRepoRoot(t)
	runtimeRoot := filepath.Join(repoRoot, "internal", "runtime")
	entries, err := os.ReadDir(runtimeRoot)
	if err != nil {
		t.Fatalf("read runtime root: %v", err)
	}

	var violations []string
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") {
			continue
		}
		violations = append(violations, filepath.ToSlash(filepath.Join("internal", "runtime", entry.Name())))
	}
	if len(violations) != 0 {
		failWithGuidance(
			t,
			"internal/runtime root must stay empty",
			"create or extend a concrete subpackage such as internal/runtime/logging or internal/runtime/tracing instead of adding new root-level Go files.",
			violations,
		)
	}
}
