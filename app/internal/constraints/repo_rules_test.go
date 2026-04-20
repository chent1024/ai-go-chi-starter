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
