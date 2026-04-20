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
