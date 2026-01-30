package parser

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Language patterns for parsing different languages
type LanguagePattern struct {
	Extensions []string
	ClassRegex *regexp.Regexp
	FuncRegex  *regexp.Regexp
	ImportRegex *regexp.Regexp
	StructRegex *regexp.Regexp
	InterfaceRegex *regexp.Regexp
}

var languagePatterns = map[string]*LanguagePattern{
	"go": {
		Extensions:     []string{".go"},
		ClassRegex:     regexp.MustCompile(`type\s+(\w+)\s+struct\s*\{`),
		FuncRegex:      regexp.MustCompile(`func\s+(?:\([^)]+\)\s+)?(\w+)\s*\(`),
		ImportRegex:    regexp.MustCompile(`import\s+(?:\(\s*)?["']([^"']+)["']`),
		InterfaceRegex: regexp.MustCompile(`type\s+(\w+)\s+interface\s*\{`),
	},
	"java": {
		Extensions:     []string{".java"},
		ClassRegex:     regexp.MustCompile(`(?:public\s+|private\s+|protected\s+)?(?:abstract\s+|final\s+)?class\s+(\w+)`),
		FuncRegex:      regexp.MustCompile(`(?:public\s+|private\s+|protected\s+)?(?:static\s+)?(?:final\s+)?(?:synchronized\s+)?(?:\w+(?:<[^>]+>)?)\s+(\w+)\s*\(`),
		ImportRegex:    regexp.MustCompile(`import\s+(?:static\s+)?([^;]+);`),
		InterfaceRegex: regexp.MustCompile(`(?:public\s+|private\s+|protected\s+)?interface\s+(\w+)`),
	},
	"kotlin": {
		Extensions:     []string{".kt", ".kts"},
		ClassRegex:     regexp.MustCompile(`(?:data\s+|sealed\s+|open\s+|abstract\s+)?class\s+(\w+)`),
		FuncRegex:      regexp.MustCompile(`fun\s+(?:<[^>]+>\s+)?(\w+)\s*\(`),
		ImportRegex:    regexp.MustCompile(`import\s+([^\s]+)`),
		InterfaceRegex: regexp.MustCompile(`interface\s+(\w+)`),
	},
	"python": {
		Extensions:     []string{".py"},
		ClassRegex:     regexp.MustCompile(`class\s+(\w+)\s*[:\(]`),
		FuncRegex:      regexp.MustCompile(`def\s+(\w+)\s*\(`),
		ImportRegex:    regexp.MustCompile(`(?:from\s+(\S+)\s+)?import\s+([^#\n]+)`),
		InterfaceRegex: nil, // Python uses ABC
	},
	"typescript": {
		Extensions:     []string{".ts", ".tsx"},
		ClassRegex:     regexp.MustCompile(`(?:export\s+)?(?:abstract\s+)?class\s+(\w+)`),
		FuncRegex:      regexp.MustCompile(`(?:export\s+)?(?:async\s+)?function\s+(\w+)|(?:const|let|var)\s+(\w+)\s*=\s*(?:async\s+)?\(`),
		ImportRegex:    regexp.MustCompile(`import\s+(?:{[^}]+}|\*\s+as\s+\w+|\w+)\s+from\s+['"]([^'"]+)['"]`),
		InterfaceRegex: regexp.MustCompile(`(?:export\s+)?interface\s+(\w+)`),
	},
	"javascript": {
		Extensions:     []string{".js", ".jsx", ".mjs"},
		ClassRegex:     regexp.MustCompile(`(?:export\s+)?class\s+(\w+)`),
		FuncRegex:      regexp.MustCompile(`(?:export\s+)?(?:async\s+)?function\s+(\w+)|(?:const|let|var)\s+(\w+)\s*=\s*(?:async\s+)?\(`),
		ImportRegex:    regexp.MustCompile(`import\s+(?:{[^}]+}|\*\s+as\s+\w+|\w+)\s+from\s+['"]([^'"]+)['"]|require\s*\(\s*['"]([^'"]+)['"]\s*\)`),
		InterfaceRegex: nil,
	},
	"swift": {
		Extensions:     []string{".swift"},
		ClassRegex:     regexp.MustCompile(`(?:public\s+|private\s+|internal\s+|fileprivate\s+|open\s+)?(?:final\s+)?class\s+(\w+)`),
		FuncRegex:      regexp.MustCompile(`func\s+(\w+)\s*[<\(]`),
		ImportRegex:    regexp.MustCompile(`import\s+(\w+)`),
		InterfaceRegex: regexp.MustCompile(`protocol\s+(\w+)`),
		StructRegex:    regexp.MustCompile(`struct\s+(\w+)`),
	},
	"csharp": {
		Extensions:     []string{".cs"},
		ClassRegex:     regexp.MustCompile(`(?:public\s+|private\s+|protected\s+|internal\s+)?(?:static\s+|sealed\s+|abstract\s+|partial\s+)?class\s+(\w+)`),
		FuncRegex:      regexp.MustCompile(`(?:public\s+|private\s+|protected\s+|internal\s+)?(?:static\s+|virtual\s+|override\s+|async\s+)?(?:\w+(?:<[^>]+>)?)\s+(\w+)\s*\(`),
		ImportRegex:    regexp.MustCompile(`using\s+(?:static\s+)?([^;]+);`),
		InterfaceRegex: regexp.MustCompile(`(?:public\s+|private\s+|protected\s+|internal\s+)?interface\s+(\w+)`),
		StructRegex:    regexp.MustCompile(`(?:public\s+|private\s+)?struct\s+(\w+)`),
	},
	"rust": {
		Extensions:     []string{".rs"},
		ClassRegex:     regexp.MustCompile(`struct\s+(\w+)`),
		FuncRegex:      regexp.MustCompile(`(?:pub\s+)?(?:async\s+)?fn\s+(\w+)`),
		ImportRegex:    regexp.MustCompile(`use\s+([^;]+);`),
		InterfaceRegex: regexp.MustCompile(`trait\s+(\w+)`),
	},
}

// getLanguageForFile returns the language pattern for a file extension
func getLanguageForFile(filename string) *LanguagePattern {
	ext := strings.ToLower(filepath.Ext(filename))
	for _, pattern := range languagePatterns {
		for _, e := range pattern.Extensions {
			if e == ext {
				return pattern
			}
		}
	}
	return nil
}

// ParseClassesMultiLang extracts class/struct info from multiple languages
func ParseClassesMultiLang(root string) []ClassInfo {
	var classes []ClassInfo

	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}

		// Skip test files and hidden files
		name := info.Name()
		if strings.HasPrefix(name, ".") || strings.Contains(name, "_test.") || strings.Contains(name, ".test.") || strings.Contains(name, ".spec.") {
			return nil
		}

		// Skip common ignore dirs
		if strings.Contains(path, "node_modules") || strings.Contains(path, "vendor") || strings.Contains(path, "__pycache__") || strings.Contains(path, ".git") {
			return nil
		}

		lang := getLanguageForFile(name)
		if lang == nil {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return nil
		}
		defer file.Close()

		// Get relative package/module name
		rel, _ := filepath.Rel(root, path)
		pkg := filepath.Dir(rel)
		if pkg == "." {
			pkg = "root"
		}

		scanner := bufio.NewScanner(file)
		lineNum := 0
		var currentClass *ClassInfo

		for scanner.Scan() {
			line := scanner.Text()
			lineNum++

			// Find classes
			if lang.ClassRegex != nil {
				if matches := lang.ClassRegex.FindStringSubmatch(line); len(matches) > 1 {
					if currentClass != nil {
						classes = append(classes, *currentClass)
					}
					currentClass = &ClassInfo{
						Name:    matches[1],
						Package: pkg,
						File:    path,
					}
				}
			}

			// Find structs (for languages that have them separately)
			if lang.StructRegex != nil {
				if matches := lang.StructRegex.FindStringSubmatch(line); len(matches) > 1 {
					if currentClass != nil {
						classes = append(classes, *currentClass)
					}
					currentClass = &ClassInfo{
						Name:    matches[1],
						Package: pkg,
						File:    path,
					}
				}
			}

			// Find interfaces
			if lang.InterfaceRegex != nil {
				if matches := lang.InterfaceRegex.FindStringSubmatch(line); len(matches) > 1 {
					classes = append(classes, ClassInfo{
						Name:    matches[1] + " (interface)",
						Package: pkg,
						File:    path,
					})
				}
			}

			// Find methods for current class
			if currentClass != nil && lang.FuncRegex != nil {
				if matches := lang.FuncRegex.FindStringSubmatch(line); len(matches) > 1 {
					methodName := matches[1]
					if methodName == "" && len(matches) > 2 {
						methodName = matches[2]
					}
					if methodName != "" && methodName != currentClass.Name {
						currentClass.Methods = append(currentClass.Methods, MethodInfo{
							Name: methodName,
						})
					}
				}
			}
		}

		if currentClass != nil {
			classes = append(classes, *currentClass)
		}

		return nil
	})

	return classes
}

// ParseFunctionsMultiLang extracts functions from multiple languages
func ParseFunctionsMultiLang(root string) []FunctionInfo {
	var funcs []FunctionInfo

	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}

		name := info.Name()
		if strings.HasPrefix(name, ".") || strings.Contains(name, "_test.") || strings.Contains(name, ".test.") || strings.Contains(name, ".spec.") {
			return nil
		}

		if strings.Contains(path, "node_modules") || strings.Contains(path, "vendor") || strings.Contains(path, "__pycache__") || strings.Contains(path, ".git") {
			return nil
		}

		lang := getLanguageForFile(name)
		if lang == nil || lang.FuncRegex == nil {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return nil
		}
		defer file.Close()

		rel, _ := filepath.Rel(root, path)
		pkg := filepath.Dir(rel)
		if pkg == "." {
			pkg = "root"
		}

		scanner := bufio.NewScanner(file)
		lineNum := 0

		for scanner.Scan() {
			line := scanner.Text()
			lineNum++

			if matches := lang.FuncRegex.FindStringSubmatch(line); len(matches) > 1 {
				funcName := matches[1]
				if funcName == "" && len(matches) > 2 {
					funcName = matches[2]
				}
				if funcName != "" {
					funcs = append(funcs, FunctionInfo{
						Name:    funcName,
						Package: pkg,
						File:    path,
						Line:    lineNum,
					})
				}
			}
		}

		return nil
	})

	return funcs
}

// ParseDependenciesMultiLang extracts imports/dependencies from multiple languages
func ParseDependenciesMultiLang(root string) []Dependency {
	var deps []Dependency

	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}

		name := info.Name()
		if strings.HasPrefix(name, ".") {
			return nil
		}

		if strings.Contains(path, "node_modules") || strings.Contains(path, "vendor") || strings.Contains(path, "__pycache__") || strings.Contains(path, ".git") {
			return nil
		}

		lang := getLanguageForFile(name)
		if lang == nil || lang.ImportRegex == nil {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return nil
		}
		defer file.Close()

		rel, _ := filepath.Rel(root, path)
		pkg := filepath.Dir(rel)
		if pkg == "." {
			pkg = "root"
		}

		scanner := bufio.NewScanner(file)
		seen := make(map[string]bool)

		for scanner.Scan() {
			line := scanner.Text()

			if matches := lang.ImportRegex.FindStringSubmatch(line); len(matches) > 1 {
				importPath := matches[1]
				if importPath == "" && len(matches) > 2 {
					importPath = matches[2]
				}
				importPath = strings.TrimSpace(importPath)

				if importPath != "" && !seen[importPath] {
					seen[importPath] = true
					deps = append(deps, Dependency{
						From:    rel,
						To:      importPath,
						Package: pkg,
					})
				}
			}
		}

		return nil
	})

	return deps
}

// ParseStructureMultiLang analyzes project structure for multiple languages
func ParseStructureMultiLang(root string) Structure {
	structure := Structure{}
	packageMap := make(map[string]*ModuleInfo)

	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}

		name := info.Name()
		if strings.HasPrefix(name, ".") || strings.Contains(name, "_test.") || strings.Contains(name, ".test.") || strings.Contains(name, ".spec.") {
			return nil
		}

		if strings.Contains(path, "node_modules") || strings.Contains(path, "vendor") || strings.Contains(path, "__pycache__") || strings.Contains(path, ".git") || strings.Contains(path, "dist") || strings.Contains(path, "build") {
			return nil
		}

		lang := getLanguageForFile(name)
		if lang == nil {
			return nil
		}

		dir := filepath.Dir(path)
		rel, _ := filepath.Rel(root, dir)
		if rel == "." {
			rel = "root"
		}

		if _, exists := packageMap[dir]; !exists {
			packageMap[dir] = &ModuleInfo{
				Name: filepath.Base(dir),
				Path: dir,
			}
			structure.Packages = append(structure.Packages, rel)
		}

		mod := packageMap[dir]
		mod.Files = append(mod.Files, name)

		// Check for entry points
		if name == "main.go" || name == "main.py" || name == "index.js" || name == "index.ts" || name == "App.tsx" || name == "Main.java" || name == "Program.cs" || name == "main.swift" || name == "main.rs" {
			structure.MainFiles = append(structure.MainFiles, path)
		}

		// Parse for structs and functions
		file, err := os.Open(path)
		if err != nil {
			return nil
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Text()

			if lang.ClassRegex != nil {
				if matches := lang.ClassRegex.FindStringSubmatch(line); len(matches) > 1 {
					mod.Structs = append(mod.Structs, matches[1])
				}
			}

			if lang.StructRegex != nil {
				if matches := lang.StructRegex.FindStringSubmatch(line); len(matches) > 1 {
					mod.Structs = append(mod.Structs, matches[1])
				}
			}

			if lang.FuncRegex != nil {
				if matches := lang.FuncRegex.FindStringSubmatch(line); len(matches) > 1 {
					funcName := matches[1]
					if funcName == "" && len(matches) > 2 {
						funcName = matches[2]
					}
					if funcName != "" {
						mod.Funcs = append(mod.Funcs, funcName)
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
