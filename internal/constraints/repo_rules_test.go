package constraints

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
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
		t.Fatalf("environment reads must stay in internal/config/config.go:\n%s", strings.Join(violations, "\n"))
	}
}

func TestServiceLayerDoesNotImportNetHTTP(t *testing.T) {
	repoRoot := findRepoRoot(t)
	serviceRoot := filepath.Join(repoRoot, "internal", "service")
	var violations []string

	err := filepath.WalkDir(serviceRoot, func(path string, d os.DirEntry, err error) error {
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
			if strings.Trim(importSpec.Path.Value, "\"") != "net/http" {
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
		t.Fatalf("walk service layer: %v", err)
	}
	if len(violations) != 0 {
		t.Fatalf("service layer must not import net/http:\n%s", strings.Join(violations, "\n"))
	}
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
