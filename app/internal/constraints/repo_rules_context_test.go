package constraints

import (
	"go/ast"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestProductionCodeDoesNotUseContextBackgroundOutsideEntryPoints(t *testing.T) {
	violations := collectPackageFunctionCallViolations(
		t,
		[]string{
			filepath.Join("internal", "service"),
			filepath.Join("internal", "infra"),
			filepath.Join("internal", "transport"),
			filepath.Join("internal", "worker"),
		},
		"context",
		map[string]struct{}{
			"Background": {},
			"TODO":       {},
		},
	)
	if len(violations) != 0 {
		failWithGuidance(
			t,
			"production code must not use context.Background or context.TODO outside entrypoints and runtime internals",
			"thread the incoming context through service, repository, transport, and worker code; only top-level process setup or dedicated runtime internals should create root contexts.",
			violations,
		)
	}
}

func TestPostgresAndMigrateUseContextAwareDatabaseCalls(t *testing.T) {
	violations := collectSelectorNameViolations(
		t,
		[]string{
			filepath.Join("internal", "infra", "store", "postgres"),
			filepath.Join("cmd", "migrate"),
		},
		map[string]struct{}{
			"Query":    {},
			"QueryRow": {},
			"Exec":     {},
			"Begin":    {},
		},
		func(file *ast.File) bool {
			return len(importedAliases(file, "database/sql")) != 0
		},
	)
	if len(violations) != 0 {
		failWithGuidance(
			t,
			"postgres and migration code must use context-aware database APIs",
			"use QueryContext, QueryRowContext, ExecContext, and BeginTx so request cancellation and shutdown propagate into database work.",
			violations,
		)
	}
}

func TestOutboundHTTPUsesNewRequestWithContext(t *testing.T) {
	violations := collectPackageFunctionCallViolations(
		t,
		[]string{
			filepath.Join("internal", "infra"),
			filepath.Join("internal", "transport"),
			filepath.Join("internal", "worker"),
		},
		"net/http",
		map[string]struct{}{
			"NewRequest": {},
		},
	)
	if len(violations) != 0 {
		failWithGuidance(
			t,
			"outbound HTTP must use http.NewRequestWithContext",
			"construct outbound requests with http.NewRequestWithContext so cancellation and request deadlines continue into downstream calls.",
			violations,
		)
	}
}

func TestPostgresRepositoriesUseSQLCGeneratedQueries(t *testing.T) {
	repoRoot := findRepoRoot(t)
	postgresRoot := filepath.Join(repoRoot, "internal", "infra", "store", "postgres")
	var violations []string

	err := filepath.WalkDir(postgresRoot, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if d.Name() == "sqlc" {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}

		file, fset, err := parseFile(path)
		if err != nil {
			return err
		}
		relPath, err := filepath.Rel(repoRoot, path)
		if err != nil {
			return err
		}
		violations = append(
			violations,
			selectorCallViolations(
				filepath.ToSlash(relPath),
				file,
				fset,
				"database/sql",
				map[string]struct{}{
					"ExecContext":     {},
					"QueryContext":    {},
					"QueryRowContext": {},
					"PrepareContext":  {},
				},
			)...,
		)
		return nil
	})
	if err != nil {
		t.Fatalf("walk postgres root: %v", err)
	}
	if len(violations) != 0 {
		failWithGuidance(
			t,
			"postgres repositories must execute SQL through generated sqlc queries",
			"move repository SQL into app/db/queries/*.sql, regenerate app/internal/infra/store/postgres/sqlc, and keep non-generated repository files limited to domain mapping, trace, and error translation.",
			violations,
		)
	}
}
