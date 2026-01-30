package parser

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// FileNode represents a file or directory in the tree
type FileNode struct {
	Name     string
	Path     string
	IsDir    bool
	Children []*FileNode
	Size     int64
	ModTime  time.Time
}

// ClassInfo represents a struct/class
type ClassInfo struct {
	Name       string
	Package    string
	Fields     []FieldInfo
	Methods    []MethodInfo
	Implements []string
	File       string
}

// FieldInfo represents a struct field
type FieldInfo struct {
	Name string
	Type string
}

// MethodInfo represents a method
type MethodInfo struct {
	Name       string
	Receiver   string
	Parameters []string
	Returns    []string
}

// FunctionInfo represents a function
type FunctionInfo struct {
	Name       string
	Package    string
	File       string
	Parameters []string
	Returns    []string
	Line       int
}

// Dependency represents an import dependency
type Dependency struct {
	From    string
	To      string
	Package string
}

// ProjectStats holds project statistics
type ProjectStats struct {
	TotalFiles    int
	TotalLines    int
	TotalPackages int
	TotalFuncs    int
	TotalStructs  int
	Languages     map[string]int
	LargestFiles  []FileInfo
}

// FileInfo for stats
type FileInfo struct {
	Path  string
	Lines int
	Size  int64
}

// RecentChange represents a recently modified file
type RecentChange struct {
	Path    string
	ModTime time.Time
	Size    int64
}

// Structure represents the overall project structure
type Structure struct {
	Packages  []string
	MainFiles []string
	Modules   []ModuleInfo
}

// ModuleInfo represents a module/package
type ModuleInfo struct {
	Name    string
	Path    string
	Files   []string
	Structs []string
	Funcs   []string
}

// ParseFileTree builds a file tree structure
func ParseFileTree(root string) *FileNode {
	rootNode := &FileNode{
		Name:  filepath.Base(root),
		Path:  root,
		IsDir: true,
	}

	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		// Skip hidden files and common ignore patterns
		name := info.Name()
		if strings.HasPrefix(name, ".") || name == "node_modules" || name == "vendor" || name == "__pycache__" {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if path == root {
			return nil
		}

		rel, _ := filepath.Rel(root, path)
		parts := strings.Split(rel, string(os.PathSeparator))

		current := rootNode
		for i, part := range parts {
			found := false
			for _, child := range current.Children {
				if child.Name == part {
					current = child
					found = true
					break
				}
			}
			if !found {
				isLast := i == len(parts)-1
				newNode := &FileNode{
					Name:    part,
					Path:    filepath.Join(root, strings.Join(parts[:i+1], string(os.PathSeparator))),
					IsDir:   info.IsDir() && isLast || !isLast,
					Size:    info.Size(),
					ModTime: info.ModTime(),
				}
				current.Children = append(current.Children, newNode)
				current = newNode
			}
		}
		return nil
	})

	return rootNode
}

// ParseClasses extracts struct/class information from Go files
func ParseClasses(root string) []ClassInfo {
	var classes []ClassInfo
	fset := token.NewFileSet()

	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}
		if strings.Contains(path, "_test.go") {
			return nil
		}

		node, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
		if err != nil {
			return nil
		}

		for _, decl := range node.Decls {
			genDecl, ok := decl.(*ast.GenDecl)
			if !ok || genDecl.Tok != token.TYPE {
				continue
			}

			for _, spec := range genDecl.Specs {
				typeSpec, ok := spec.(*ast.TypeSpec)
				if !ok {
					continue
				}

				structType, ok := typeSpec.Type.(*ast.StructType)
				if !ok {
					continue
				}

				class := ClassInfo{
					Name:    typeSpec.Name.Name,
					Package: node.Name.Name,
					File:    path,
				}

				if structType.Fields != nil {
					for _, field := range structType.Fields.List {
						fieldType := exprToString(field.Type)
						if len(field.Names) > 0 {
							for _, name := range field.Names {
								class.Fields = append(class.Fields, FieldInfo{
									Name: name.Name,
									Type: fieldType,
								})
							}
						} else {
							// Embedded field
							class.Fields = append(class.Fields, FieldInfo{
								Name: fieldType,
								Type: "(embedded)",
							})
						}
					}
				}

				classes = append(classes, class)
			}
		}

		// Find methods
		for _, decl := range node.Decls {
			funcDecl, ok := decl.(*ast.FuncDecl)
			if !ok || funcDecl.Recv == nil {
				continue
			}

			receiverType := ""
			if len(funcDecl.Recv.List) > 0 {
				receiverType = exprToString(funcDecl.Recv.List[0].Type)
				receiverType = strings.TrimPrefix(receiverType, "*")
			}

			method := MethodInfo{
				Name:     funcDecl.Name.Name,
				Receiver: receiverType,
			}

			if funcDecl.Type.Params != nil {
				for _, param := range funcDecl.Type.Params.List {
					method.Parameters = append(method.Parameters, exprToString(param.Type))
				}
			}

			if funcDecl.Type.Results != nil {
				for _, result := range funcDecl.Type.Results.List {
					method.Returns = append(method.Returns, exprToString(result.Type))
				}
			}

			// Add method to corresponding class
			for i := range classes {
				if classes[i].Name == receiverType && classes[i].Package == node.Name.Name {
					classes[i].Methods = append(classes[i].Methods, method)
					break
				}
			}
		}

		return nil
	})

	return classes
}

// ParseFunctions extracts all functions
func ParseFunctions(root string) []FunctionInfo {
	var funcs []FunctionInfo
	fset := token.NewFileSet()

	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}
		if strings.Contains(path, "_test.go") {
			return nil
		}

		node, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
		if err != nil {
			return nil
		}

		for _, decl := range node.Decls {
			funcDecl, ok := decl.(*ast.FuncDecl)
			if !ok {
				continue
			}

			fn := FunctionInfo{
				Name:    funcDecl.Name.Name,
				Package: node.Name.Name,
				File:    path,
				Line:    fset.Position(funcDecl.Pos()).Line,
			}

			if funcDecl.Recv != nil {
				// Skip methods, only get standalone functions
				continue
			}

			if funcDecl.Type.Params != nil {
				for _, param := range funcDecl.Type.Params.List {
					fn.Parameters = append(fn.Parameters, exprToString(param.Type))
				}
			}

			if funcDecl.Type.Results != nil {
				for _, result := range funcDecl.Type.Results.List {
					fn.Returns = append(fn.Returns, exprToString(result.Type))
				}
			}

			funcs = append(funcs, fn)
		}

		return nil
	})

	return funcs
}

// ParseDependencies extracts import dependencies
func ParseDependencies(root string) []Dependency {
	var deps []Dependency
	fset := token.NewFileSet()

	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}

		node, err := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
		if err != nil {
			return nil
		}

		rel, _ := filepath.Rel(root, path)
		for _, imp := range node.Imports {
			importPath := strings.Trim(imp.Path.Value, `"`)
			deps = append(deps, Dependency{
				From:    rel,
				To:      importPath,
				Package: node.Name.Name,
			})
		}

		return nil
	})

	return deps
}

// ParseRecentChanges finds recently modified files
func ParseRecentChanges(root string) []RecentChange {
	var changes []RecentChange

	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}

		name := info.Name()
		if strings.HasPrefix(name, ".") || strings.Contains(path, "node_modules") {
			return nil
		}

		changes = append(changes, RecentChange{
			Path:    path,
			ModTime: info.ModTime(),
			Size:    info.Size(),
		})
		return nil
	})

	// Sort by modification time, newest first
	sort.Slice(changes, func(i, j int) bool {
		return changes[i].ModTime.After(changes[j].ModTime)
	})

	// Return top 20
	if len(changes) > 20 {
		changes = changes[:20]
	}

	return changes
}

// ParseStats gathers project statistics
func ParseStats(root string) ProjectStats {
	stats := ProjectStats{
		Languages: make(map[string]int),
	}

	packages := make(map[string]bool)
	fset := token.NewFileSet()

	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}

		name := info.Name()
		if strings.HasPrefix(name, ".") {
			return nil
		}

		ext := filepath.Ext(name)
		if ext != "" {
			stats.Languages[ext]++
		}

		stats.TotalFiles++

		// Count lines and parse Go files
		if strings.HasSuffix(path, ".go") && !strings.Contains(path, "_test.go") {
			data, err := os.ReadFile(path)
			if err == nil {
				lines := len(strings.Split(string(data), "\n"))
				stats.TotalLines += lines

				stats.LargestFiles = append(stats.LargestFiles, FileInfo{
					Path:  path,
					Lines: lines,
					Size:  info.Size(),
				})
			}

			node, err := parser.ParseFile(fset, path, nil, 0)
			if err == nil {
				packages[node.Name.Name] = true

				for _, decl := range node.Decls {
					switch d := decl.(type) {
					case *ast.FuncDecl:
						stats.TotalFuncs++
					case *ast.GenDecl:
						if d.Tok == token.TYPE {
							for _, spec := range d.Specs {
								if ts, ok := spec.(*ast.TypeSpec); ok {
									if _, ok := ts.Type.(*ast.StructType); ok {
										stats.TotalStructs++
									}
								}
							}
						}
					}
				}
			}
		}

		return nil
	})

	stats.TotalPackages = len(packages)

	// Sort largest files
	sort.Slice(stats.LargestFiles, func(i, j int) bool {
		return stats.LargestFiles[i].Lines > stats.LargestFiles[j].Lines
	})
	if len(stats.LargestFiles) > 5 {
		stats.LargestFiles = stats.LargestFiles[:5]
	}

	return stats
}

// ParseStructure analyzes overall project structure
func ParseStructure(root string) Structure {
	structure := Structure{}
	packageMap := make(map[string]*ModuleInfo)
	fset := token.NewFileSet()

	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}

		name := info.Name()
		if strings.HasPrefix(name, ".") || strings.Contains(path, "_test.go") {
			return nil
		}

		node, err := parser.ParseFile(fset, path, nil, 0)
		if err != nil {
			return nil
		}

		pkgName := node.Name.Name
		dir := filepath.Dir(path)

		if _, exists := packageMap[dir]; !exists {
			packageMap[dir] = &ModuleInfo{
				Name: pkgName,
				Path: dir,
			}
			structure.Packages = append(structure.Packages, pkgName)
		}

		mod := packageMap[dir]
		mod.Files = append(mod.Files, filepath.Base(path))

		if name == "main.go" {
			structure.MainFiles = append(structure.MainFiles, path)
		}

		for _, decl := range node.Decls {
			switch d := decl.(type) {
			case *ast.FuncDecl:
				if d.Recv == nil {
					mod.Funcs = append(mod.Funcs, d.Name.Name)
				}
			case *ast.GenDecl:
				if d.Tok == token.TYPE {
					for _, spec := range d.Specs {
						if ts, ok := spec.(*ast.TypeSpec); ok {
							if _, ok := ts.Type.(*ast.StructType); ok {
								mod.Structs = append(mod.Structs, ts.Name.Name)
							}
						}
					}
				}
			}
		}

		return nil
	})

	for _, mod := range packageMap {
		structure.Modules = append(structure.Modules, *mod)
	}

	return structure
}

func exprToString(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return "*" + exprToString(t.X)
	case *ast.SelectorExpr:
		return exprToString(t.X) + "." + t.Sel.Name
	case *ast.ArrayType:
		return "[]" + exprToString(t.Elt)
	case *ast.MapType:
		return "map[" + exprToString(t.Key) + "]" + exprToString(t.Value)
	case *ast.InterfaceType:
		return "interface{}"
	case *ast.FuncType:
		return "func"
	case *ast.ChanType:
		return "chan " + exprToString(t.Value)
	default:
		return "?"
	}
}
