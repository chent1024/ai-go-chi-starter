package constraints

import (
	"os"
	"path/filepath"
	"regexp"
	"testing"
)

func TestDomainDirectoriesFollowRecipe(t *testing.T) {
	repoRoot := findRepoRoot(t)
	serviceRoot := filepath.Join(repoRoot, "internal", "service")
	entries, err := os.ReadDir(serviceRoot)
	if err != nil {
		t.Fatalf("read service root: %v", err)
	}

	var violations []string
	for _, entry := range entries {
		if !entry.IsDir() || entry.Name() == "shared" {
			continue
		}
		domain := entry.Name()
		for _, required := range []string{
			filepath.Join("internal", "service", domain, "model.go"),
			filepath.Join("internal", "service", domain, "repository.go"),
			filepath.Join("internal", "service", domain, "service.go"),
			filepath.Join("internal", "service", domain, "service_test.go"),
			filepath.Join("internal", "infra", "store", "postgres", domain+"_repository.go"),
			filepath.Join("internal", "infra", "store", "postgres", domain+"_repository_test.go"),
			filepath.Join("internal", "transport", "httpapi", "v1", domain+"_handler.go"),
			filepath.Join("internal", "transport", "httpapi", "v1", domain+"_handler_test.go"),
		} {
			if _, err := os.Stat(filepath.Join(repoRoot, required)); err == nil {
				continue
			}
			violations = append(violations, filepath.ToSlash(required))
		}
	}
	if len(violations) != 0 {
		failWithGuidance(
			t,
			"each service domain must follow the documented recipe",
			"for every app/internal/service/<domain> package, add model.go, repository.go, service.go, service_test.go, a postgres <domain>_repository(.go/_test.go), and a v1 <domain>_handler(.go/_test.go) before extending the router.",
			violations,
		)
	}
}

func TestInternalTopLevelDirectoriesAreAllowlisted(t *testing.T) {
	violations := unexpectedSubdirectories(t, filepath.Join("internal"), []string{
		"config",
		"constraints",
		"infra",
		"runtime",
		"service",
		"transport",
		"worker",
	})
	if len(violations) != 0 {
		failWithGuidance(
			t,
			"app/internal top-level directories must stay allowlisted",
			"if you truly need a new top-level internal directory, first decide its dependency boundary, then update docs/app/architecture.md, docs/app/codex-guide.md, and the allowlist or arch rules in app/internal/constraints/repo_rules_test.go plus .orch/rules/app/local.arch.rules.",
			violations,
		)
	}
}

func TestRuntimeSubdirectoriesAreAllowlisted(t *testing.T) {
	violations := unexpectedSubdirectories(t, filepath.Join("internal", "runtime"), []string{
		"logging",
		"tracing",
	})
	if len(violations) != 0 {
		failWithGuidance(
			t,
			"internal/runtime subdirectories must stay allowlisted",
			"put new logger behavior in internal/runtime/logging, put trace or span behavior in internal/runtime/tracing, and only add a brand-new runtime subpackage after documenting it in docs/app/architecture.md and docs/app/codex-guide.md.",
			violations,
		)
	}
}

func TestHTTPAPISubdirectoriesAreAllowlisted(t *testing.T) {
	violations := unexpectedSubdirectories(t, filepath.Join("internal", "transport", "httpapi"), []string{
		"drain",
		"httpx",
		"metrics",
		"middleware",
		"v1",
	})
	if len(violations) != 0 {
		failWithGuidance(
			t,
			"internal/transport/httpapi subdirectories must stay allowlisted",
			"new HTTP code should usually fit into v1, middleware, httpx, drain, or metrics; if a new protocol subpackage is necessary, document it in docs/app/architecture.md and docs/app/codex-guide.md and add matching arch rules before creating the directory.",
			violations,
		)
	}
}

func TestCmdWorkerDoesNotImplementTickerLoop(t *testing.T) {
	repoRoot := findRepoRoot(t)
	workerApp := filepath.Join(repoRoot, "cmd", "worker", "app.go")
	content, err := os.ReadFile(workerApp)
	if err != nil {
		t.Fatalf("read cmd/worker/app.go: %v", err)
	}
	if regexp.MustCompile(`\btime\.NewTicker\s*\(`).Match(content) {
		t.Fatal("cmd/worker/app.go must not implement ticker loops; move long-running loop logic to internal/worker and keep cmd/worker focused on wiring dependencies.")
	}
}
