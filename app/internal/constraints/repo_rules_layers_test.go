package constraints

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestServiceLayerDoesNotImportNetHTTP(t *testing.T) {
	violations := collectImportViolations(t, filepath.Join("internal", "service"), "net/http")
	if len(violations) != 0 {
		failWithGuidance(
			t,
			"service layer must not import net/http",
			"keep HTTP request/response handling in internal/transport/httpapi; service should receive plain inputs and return domain results.",
			violations,
		)
	}
}

func TestServiceLayerDoesNotImportChi(t *testing.T) {
	violations := collectImportViolations(t, filepath.Join("internal", "service"), "github.com/go-chi/chi/v5")
	if len(violations) != 0 {
		failWithGuidance(
			t,
			"service layer must not import chi",
			"route params and chi-specific request plumbing belong in internal/transport/httpapi; pass resolved values into service methods.",
			violations,
		)
	}
}

func TestServiceLayerDoesNotImportLogSlog(t *testing.T) {
	violations := collectImportViolations(t, filepath.Join("internal", "service"), "log/slog")
	if len(violations) != 0 {
		failWithGuidance(
			t,
			"service layer must not import log/slog",
			"pass structured loggers into transport, worker, infra, or runtime helpers; keep service packages free from direct observability wiring.",
			violations,
		)
	}
}

func TestServiceLayerDoesNotImportConfig(t *testing.T) {
	violations := collectPrefixImportViolations(t, filepath.Join("internal", "service"), "ai-go-chi-starter/internal/config")
	if len(violations) != 0 {
		failWithGuidance(
			t,
			"service layer must not import config",
			"load config in cmd/*, translate it into service dependencies or options there, and keep service packages independent from repository-wide configuration types.",
			violations,
		)
	}
}

func TestServiceLayerDoesNotImportRuntimePackages(t *testing.T) {
	violations := collectPrefixImportViolations(t, filepath.Join("internal", "service"), "ai-go-chi-starter/internal/runtime")
	if len(violations) != 0 {
		failWithGuidance(
			t,
			"service layer must not import runtime packages",
			"move logging, tracing, spans, and request context handling to transport, infra, worker, or runtime subpackages; service should stay focused on domain rules.",
			violations,
		)
	}
}

func TestTransportHandlersDoNotImportDatabaseSQL(t *testing.T) {
	handlerRoot := filepath.Join("internal", "transport", "httpapi", "v1")
	violations := collectImportViolations(t, handlerRoot, "database/sql")
	if len(violations) != 0 {
		failWithGuidance(
			t,
			"http handlers must not import database/sql",
			"move SQL access into internal/infra/store/* and expose it through service or repository interfaces; handlers should only do protocol translation.",
			violations,
		)
	}
}

func TestRepositoryPackagesDoNotImportTransportPackages(t *testing.T) {
	repoRoot := filepath.Join("internal", "infra", "store")
	violations := collectPrefixImportViolations(t, repoRoot, "ai-go-chi-starter/internal/transport")
	if len(violations) != 0 {
		failWithGuidance(
			t,
			"repository packages must not import transport packages",
			"return domain models from repositories and keep HTTP DTOs in internal/transport/httpapi only.",
			violations,
		)
	}
}

func TestNoPackageImportsRuntimeRoot(t *testing.T) {
	repoRoot := findRepoRoot(t)
	targetImport := "ai-go-chi-starter/internal/runtime"
	var violations []string

	err := filepath.WalkDir(repoRoot, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() && shouldSkipDir(d.Name()) {
			return filepath.SkipDir
		}
		if d.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}

		file, _, err := parseFile(path)
		if err != nil {
			return err
		}
		for _, importSpec := range file.Imports {
			if strings.Trim(importSpec.Path.Value, "\"") != targetImport {
				continue
			}
			relPath, err := filepath.Rel(repoRoot, path)
			if err != nil {
				return err
			}
			violations = append(violations, filepath.ToSlash(relPath))
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk repo: %v", err)
	}
	if len(violations) != 0 {
		failWithGuidance(
			t,
			"packages must not import internal/runtime root package",
			"import a concrete subpackage instead: use internal/runtime/logging for logger concerns or internal/runtime/tracing for trace and span concerns.",
			violations,
		)
	}
}
