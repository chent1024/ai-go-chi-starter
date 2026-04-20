package constraints

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
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

func unexpectedSubdirectories(t *testing.T, relDir string, allowed []string) []string {
	t.Helper()
	repoRoot := findRepoRoot(t)
	root := filepath.Join(repoRoot, relDir)
	entries, err := os.ReadDir(root)
	if err != nil {
		t.Fatalf("read %s: %v", relDir, err)
	}

	allowedSet := make(map[string]struct{}, len(allowed))
	for _, name := range allowed {
		allowedSet[name] = struct{}{}
	}

	var violations []string
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		if _, ok := allowedSet[entry.Name()]; ok {
			continue
		}
		violations = append(violations, filepath.ToSlash(filepath.Join(relDir, entry.Name())))
	}
	slices.Sort(violations)
	return violations
}

func failWithGuidance(t *testing.T, rule string, guidance string, violations []string) {
	t.Helper()
	t.Fatalf("%s\nrecommended fix: %s\nviolations:\n%s", rule, guidance, strings.Join(violations, "\n"))
}

func collectImportViolations(t *testing.T, relDir string, targetImport string) []string {
	t.Helper()
	return collectImportMatches(t, relDir, func(importPath string) bool {
		return importPath == targetImport
	})
}

func collectPrefixImportViolations(t *testing.T, relDir string, importPrefix string) []string {
	t.Helper()
	return collectImportMatches(t, relDir, func(importPath string) bool {
		return importPath == importPrefix || strings.HasPrefix(importPath, importPrefix+"/")
	})
}

func collectImportMatches(t *testing.T, relDir string, match func(string) bool) []string {
	t.Helper()
	repoRoot := findRepoRoot(t)
	root := filepath.Join(repoRoot, relDir)
	var violations []string

	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}

		file, _, err := parseFile(path)
		if err != nil {
			return err
		}
		for _, importSpec := range file.Imports {
			importPath := strings.Trim(importSpec.Path.Value, "\"")
			if !match(importPath) {
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
		t.Fatalf("walk %s: %v", relDir, err)
	}
	return violations
}

func collectPackageFunctionCallViolations(t *testing.T, relDirs []string, importPath string, selectors map[string]struct{}) []string {
	t.Helper()
	return collectASTViolations(t, relDirs, func(relPath string, file *ast.File, fset *token.FileSet) []string {
		return selectorCallViolations(relPath, file, fset, importPath, selectors)
	})
}

func collectSelectorNameViolations(
	t *testing.T,
	relDirs []string,
	selectors map[string]struct{},
	fileFilter func(*ast.File) bool,
) []string {
	t.Helper()
	return collectASTViolations(t, relDirs, func(relPath string, file *ast.File, fset *token.FileSet) []string {
		if fileFilter != nil && !fileFilter(file) {
			return nil
		}
		var violations []string
		ast.Inspect(file, func(node ast.Node) bool {
			call, ok := node.(*ast.CallExpr)
			if !ok {
				return true
			}
			selector, ok := call.Fun.(*ast.SelectorExpr)
			if !ok {
				return true
			}
			if _, ok := selectors[selector.Sel.Name]; !ok {
				return true
			}
			position := fset.Position(selector.Pos())
			violations = append(violations, formatViolation(relPath, position.Line, selector.Sel.Name))
			return true
		})
		return violations
	})
}

func collectASTViolations(
	t *testing.T,
	relDirs []string,
	collect func(relPath string, file *ast.File, fset *token.FileSet) []string,
) []string {
	t.Helper()
	repoRoot := findRepoRoot(t)
	var violations []string

	for _, relDir := range relDirs {
		root := filepath.Join(repoRoot, relDir)
		err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
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
			violations = append(violations, collect(filepath.ToSlash(relPath), file, fset)...)
			return nil
		})
		if err != nil {
			t.Fatalf("walk %s: %v", relDir, err)
		}
	}
	return violations
}

func selectorCallViolations(
	relPath string,
	file *ast.File,
	fset *token.FileSet,
	importPath string,
	selectors map[string]struct{},
) []string {
	aliases := importedAliases(file, importPath)
	if len(aliases) == 0 {
		return nil
	}

	var violations []string
	ast.Inspect(file, func(node ast.Node) bool {
		call, ok := node.(*ast.CallExpr)
		if !ok {
			return true
		}
		selector, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		target, ok := selector.X.(*ast.Ident)
		if !ok || !slices.Contains(aliases, target.Name) {
			return true
		}
		if _, ok := selectors[selector.Sel.Name]; !ok {
			return true
		}
		position := fset.Position(selector.Pos())
		violations = append(violations, formatViolation(relPath, position.Line, target.Name+"."+selector.Sel.Name))
		return true
	})
	return violations
}

func parseFile(path string) (*ast.File, *token.FileSet, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, path, nil, 0)
	if err != nil {
		return nil, nil, err
	}
	return file, fset, nil
}

func envReadViolations(relPath string, file *ast.File, fset *token.FileSet) []string {
	aliases := importedAliases(file, "os")
	syscallAliases := importedAliases(file, "syscall")
	var violations []string

	ast.Inspect(file, func(node ast.Node) bool {
		call, ok := node.(*ast.CallExpr)
		if !ok {
			return true
		}
		selector, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		target, ok := selector.X.(*ast.Ident)
		if !ok {
			return true
		}

		if slices.Contains(aliases, target.Name) && (selector.Sel.Name == "Getenv" || selector.Sel.Name == "LookupEnv") {
			position := fset.Position(selector.Pos())
			violations = append(violations, formatViolation(relPath, position.Line, "os."+selector.Sel.Name))
			return true
		}
		if slices.Contains(syscallAliases, target.Name) && selector.Sel.Name == "Getenv" {
			position := fset.Position(selector.Pos())
			violations = append(violations, formatViolation(relPath, position.Line, "syscall.Getenv"))
		}
		return true
	})
	return violations
}

func importedAliases(file *ast.File, importPath string) []string {
	var aliases []string
	for _, importSpec := range file.Imports {
		if strings.Trim(importSpec.Path.Value, "\"") != importPath {
			continue
		}
		switch {
		case importSpec.Name != nil && importSpec.Name.Name != "_":
			aliases = append(aliases, importSpec.Name.Name)
		default:
			parts := strings.Split(importPath, "/")
			aliases = append(aliases, parts[len(parts)-1])
		}
	}
	return aliases
}

func formatViolation(path string, line int, symbol string) string {
	return path + ":" + strconv.Itoa(line) + " uses " + symbol
}

func hasAnyPrefix(value string, prefixes []string) bool {
	for _, prefix := range prefixes {
		if strings.HasPrefix(value, prefix) {
			return true
		}
	}
	return false
}

func shouldSkipDir(name string) bool {
	switch name {
	case ".git", ".orch", ".tools", "bin", "dist", "vendor":
		return true
	default:
		return false
	}
}

func findRepoRoot(t *testing.T) string {
	t.Helper()

	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("repo root not found")
		}
		dir = parent
	}
}
