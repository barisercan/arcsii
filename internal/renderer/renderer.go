package renderer

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/barisercan/arcsii/internal/parser"
	"github.com/charmbracelet/lipgloss"
)

var (
	// Color palette
	cyan       = lipgloss.Color("#4ECDC4")
	pink       = lipgloss.Color("#FF6B6B")
	yellow     = lipgloss.Color("#FFE66D")
	purple     = lipgloss.Color("#A855F7")
	green      = lipgloss.Color("#10B981")
	blue       = lipgloss.Color("#3B82F6")
	orange     = lipgloss.Color("#F97316")
	gray       = lipgloss.Color("#6B7280")
	white      = lipgloss.Color("#FFFFFF")
	darkGray   = lipgloss.Color("#374151")

	// Styles
	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(cyan).
			BorderStyle(lipgloss.DoubleBorder()).
			BorderForeground(cyan).
			Padding(0, 2)

	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(purple).
			Padding(0, 1)

	classBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(blue).
			Padding(0, 1)

	methodStyle = lipgloss.NewStyle().
			Foreground(green)

	fieldStyle = lipgloss.NewStyle().
			Foreground(yellow)

	fileStyle = lipgloss.NewStyle().
			Foreground(cyan)

	dirStyle = lipgloss.NewStyle().
			Foreground(purple).
			Bold(true)

	labelStyle = lipgloss.NewStyle().
			Foreground(pink).
			Bold(true)

	dimStyle = lipgloss.NewStyle().
			Foreground(gray)

	highlightStyle = lipgloss.NewStyle().
			Foreground(white).
			Background(purple).
			Padding(0, 1)
)

// RenderWelcome renders the welcome screen
func RenderWelcome() string {
	logo := `
    ╔═══════════════════════════════════════════════════════════╗
    ║                                                           ║
    ║     ▄▀▄ █▀▄ ▄▀▀ ▄▀▀ █ █                                   ║
    ║     █▀█ █▀▄ █   ▀▀█ █ █                                   ║
    ║     ▀ ▀ ▀ ▀  ▀▀ ▀▀▀ ▀ ▀                                   ║
    ║                                                           ║
    ║         Terminal Architecture Visualizer                  ║
    ║                                                           ║
    ╚═══════════════════════════════════════════════════════════╝
`
	logoStyled := lipgloss.NewStyle().Foreground(cyan).Render(logo)

	commands := `
    ┌─────────────────────────────────────────────────────────────┐
    │  COMMANDS                                                   │
    ├─────────────────────────────────────────────────────────────┤
    │                                                             │
    │   /tree      ─────────────  File structure                  │
    │   /uml       ─────────────  Class diagrams                  │
    │   /ascii     ─────────────  ASCII architecture art          │
    │   /deps      ─────────────  Dependency graph                │
    │   /changes   ─────────────  Recent modifications            │
    │   /stats     ─────────────  Project statistics              │
    │   /funcs     ─────────────  List all functions              │
    │   /help      ─────────────  Show this help                  │
    │                                                             │
    └─────────────────────────────────────────────────────────────┘
`
	commandsStyled := lipgloss.NewStyle().Foreground(purple).Render(commands)

	tip := dimStyle.Render("\n    💡 Tip: Type a command and press Enter to explore your codebase\n")

	return logoStyled + commandsStyled + tip
}

// RenderHelp renders the help screen
func RenderHelp() string {
	return RenderWelcome()
}

// RenderTree renders a file tree
func RenderTree(root *parser.FileNode) string {
	var sb strings.Builder

	header := headerStyle.Render("📁 FILE TREE")
	sb.WriteString(header)
	sb.WriteString("\n\n")

	renderTreeNode(&sb, root, "", true)

	return sb.String()
}

func renderTreeNode(sb *strings.Builder, node *parser.FileNode, prefix string, isLast bool) {
	if node == nil {
		return
	}

	connector := "├── "
	if isLast {
		connector = "└── "
	}

	icon := getFileIcon(node.Name, node.IsDir)

	var name string
	if node.IsDir {
		name = dirStyle.Render(node.Name + "/")
	} else {
		name = fileStyle.Render(node.Name)
	}

	if prefix != "" || !node.IsDir {
		sb.WriteString(dimStyle.Render(prefix + connector))
		sb.WriteString(icon + " " + name)
		sb.WriteString("\n")
	} else {
		sb.WriteString(icon + " " + name)
		sb.WriteString("\n")
	}

	newPrefix := prefix
	if prefix != "" || !node.IsDir {
		if isLast {
			newPrefix = prefix + "    "
		} else {
			newPrefix = prefix + "│   "
		}
	}

	for i, child := range node.Children {
		isLastChild := i == len(node.Children)-1
		renderTreeNode(sb, child, newPrefix, isLastChild)
	}
}

func getFileIcon(name string, isDir bool) string {
	if isDir {
		return "📂"
	}

	ext := filepath.Ext(name)
	switch ext {
	case ".go":
		return "🔷"
	case ".js", ".ts", ".jsx", ".tsx":
		return "🟨"
	case ".py":
		return "🐍"
	case ".rs":
		return "🦀"
	case ".md":
		return "📝"
	case ".json":
		return "📋"
	case ".yaml", ".yml":
		return "⚙️"
	case ".html":
		return "🌐"
	case ".css", ".scss":
		return "🎨"
	case ".sql":
		return "🗄️"
	case ".sh":
		return "💻"
	case ".mod", ".sum":
		return "📦"
	default:
		return "📄"
	}
}

// RenderUML renders UML class diagrams
func RenderUML(classes []parser.ClassInfo) string {
	var sb strings.Builder

	header := headerStyle.Render("📐 UML CLASS DIAGRAM")
	sb.WriteString(header)
	sb.WriteString("\n\n")

	if len(classes) == 0 {
		sb.WriteString(dimStyle.Render("  No structs/classes found in this project.\n"))
		return sb.String()
	}

	for _, class := range classes {
		sb.WriteString(renderClassBox(class))
		sb.WriteString("\n")
	}

	// Render relationships
	if len(classes) > 1 {
		sb.WriteString(labelStyle.Render("  RELATIONSHIPS"))
		sb.WriteString("\n")
		sb.WriteString(dimStyle.Render("  ─────────────"))
		sb.WriteString("\n\n")

		for _, class := range classes {
			for _, field := range class.Fields {
				for _, other := range classes {
					if strings.Contains(field.Type, other.Name) && other.Name != class.Name {
						arrow := fmt.Sprintf("    %s ──────▶ %s",
							lipgloss.NewStyle().Foreground(blue).Render(class.Name),
							lipgloss.NewStyle().Foreground(green).Render(other.Name))
						relation := dimStyle.Render(fmt.Sprintf(" (has %s)", field.Name))
						sb.WriteString(arrow + relation + "\n")
					}
				}
			}
		}
	}

	return sb.String()
}

func renderClassBox(class parser.ClassInfo) string {
	var lines []string

	// Class name header
	nameWidth := len(class.Name) + 4
	minWidth := 30
	if nameWidth < minWidth {
		nameWidth = minWidth
	}

	// Package info
	pkgInfo := dimStyle.Render(fmt.Sprintf("pkg: %s", class.Package))

	// Class name
	className := lipgloss.NewStyle().
		Bold(true).
		Foreground(white).
		Background(blue).
		Padding(0, 1).
		Render(class.Name)

	lines = append(lines, className+"  "+pkgInfo)
	lines = append(lines, strings.Repeat("─", nameWidth))

	// Fields section
	if len(class.Fields) > 0 {
		lines = append(lines, labelStyle.Render("Fields:"))
		for _, field := range class.Fields {
			fieldLine := fmt.Sprintf("  %s %s",
				fieldStyle.Render(field.Name),
				dimStyle.Render(field.Type))
			lines = append(lines, fieldLine)
		}
	}

	// Methods section
	if len(class.Methods) > 0 {
		lines = append(lines, "")
		lines = append(lines, labelStyle.Render("Methods:"))
		for _, method := range class.Methods {
			params := strings.Join(method.Parameters, ", ")
			returns := strings.Join(method.Returns, ", ")

			methodLine := fmt.Sprintf("  %s(%s)",
				methodStyle.Render(method.Name),
				dimStyle.Render(params))

			if returns != "" {
				methodLine += dimStyle.Render(" → " + returns)
			}
			lines = append(lines, methodLine)
		}
	}

	content := strings.Join(lines, "\n")
	return classBoxStyle.Render(content)
}

// RenderASCIIArt renders ASCII art architecture view
func RenderASCIIArt(structure parser.Structure) string {
	var sb strings.Builder

	header := headerStyle.Render("🎨 ASCII ARCHITECTURE")
	sb.WriteString(header)
	sb.WriteString("\n\n")

	if len(structure.Modules) == 0 {
		sb.WriteString(dimStyle.Render("  No modules found in this project.\n"))
		return sb.String()
	}

	// ASCII art representation of architecture
	sb.WriteString(lipgloss.NewStyle().Foreground(cyan).Render(`
                    ╔════════════════════════════════════════╗
                    ║           PROJECT ARCHITECTURE         ║
                    ╚════════════════════════════════════════╝
                                       │
                                       ▼
`))

	// Entry points
	if len(structure.MainFiles) > 0 {
		sb.WriteString(lipgloss.NewStyle().Foreground(pink).Render(`
                    ┌────────────────────────────────────────┐
                    │            🚀 ENTRY POINTS             │
                    └────────────────────────────────────────┘
`))
		for _, main := range structure.MainFiles {
			sb.WriteString(fmt.Sprintf("                                  ◆ %s\n", fileStyle.Render(filepath.Base(main))))
		}
		sb.WriteString("\n                                       │\n                                       ▼\n")
	}

	// Modules as boxes
	sb.WriteString(lipgloss.NewStyle().Foreground(purple).Render(`
                    ┌────────────────────────────────────────┐
                    │              📦 MODULES                │
                    └────────────────────────────────────────┘
`))
	sb.WriteString("\n")

	for _, mod := range structure.Modules {
		box := renderModuleBox(mod)
		sb.WriteString(box)
		sb.WriteString("\n")
	}

	return sb.String()
}

func renderModuleBox(mod parser.ModuleInfo) string {
	var lines []string

	// Module header
	modName := highlightStyle.Render(mod.Name)
	lines = append(lines, "    "+modName)
	lines = append(lines, "    "+strings.Repeat("─", 40))

	// Files
	if len(mod.Files) > 0 {
		lines = append(lines, "    "+labelStyle.Render("Files:"))
		for _, f := range mod.Files {
			lines = append(lines, "      "+fileStyle.Render("◇ "+f))
		}
	}

	// Structs
	if len(mod.Structs) > 0 {
		lines = append(lines, "    "+labelStyle.Render("Structs:"))
		for _, s := range mod.Structs {
			lines = append(lines, "      "+lipgloss.NewStyle().Foreground(blue).Render("◈ "+s))
		}
	}

	// Functions
	if len(mod.Funcs) > 0 {
		lines = append(lines, "    "+labelStyle.Render("Functions:"))
		for _, f := range mod.Funcs {
			lines = append(lines, "      "+methodStyle.Render("◉ "+f))
		}
	}

	return strings.Join(lines, "\n")
}

// RenderDeps renders dependency graph
func RenderDeps(deps []parser.Dependency) string {
	var sb strings.Builder

	header := headerStyle.Render("🔗 DEPENDENCY GRAPH")
	sb.WriteString(header)
	sb.WriteString("\n\n")

	if len(deps) == 0 {
		sb.WriteString(dimStyle.Render("  No dependencies found.\n"))
		return sb.String()
	}

	// Group by package
	packages := make(map[string][]string)
	for _, dep := range deps {
		packages[dep.Package] = append(packages[dep.Package], dep.To)
	}

	// Deduplicate and render
	for pkg, imports := range packages {
		// Package header
		pkgBox := lipgloss.NewStyle().
			Foreground(white).
			Background(purple).
			Padding(0, 1).
			Render(pkg)
		sb.WriteString("  " + pkgBox + "\n")

		// Dedupe imports
		seen := make(map[string]bool)
		var unique []string
		for _, imp := range imports {
			if !seen[imp] {
				seen[imp] = true
				unique = append(unique, imp)
			}
		}

		// Render imports as tree
		for i, imp := range unique {
			connector := "├──"
			if i == len(unique)-1 {
				connector = "└──"
			}

			// Color based on type
			var impStyled string
			if strings.HasPrefix(imp, "github.com/barisercan/arcsii") {
				impStyled = lipgloss.NewStyle().Foreground(green).Render(imp)
			} else if strings.Contains(imp, ".") {
				impStyled = lipgloss.NewStyle().Foreground(orange).Render(imp)
			} else {
				impStyled = lipgloss.NewStyle().Foreground(cyan).Render(imp)
			}

			sb.WriteString(fmt.Sprintf("  %s %s\n", dimStyle.Render(connector), impStyled))
		}
		sb.WriteString("\n")
	}

	// Legend
	sb.WriteString("\n")
	sb.WriteString(dimStyle.Render("  Legend: "))
	sb.WriteString(lipgloss.NewStyle().Foreground(green).Render("internal"))
	sb.WriteString(dimStyle.Render(" │ "))
	sb.WriteString(lipgloss.NewStyle().Foreground(orange).Render("external"))
	sb.WriteString(dimStyle.Render(" │ "))
	sb.WriteString(lipgloss.NewStyle().Foreground(cyan).Render("stdlib"))
	sb.WriteString("\n")

	return sb.String()
}

// RenderChanges renders recent changes
func RenderChanges(changes []parser.RecentChange) string {
	var sb strings.Builder

	header := headerStyle.Render("🕐 RECENT CHANGES")
	sb.WriteString(header)
	sb.WriteString("\n\n")

	if len(changes) == 0 {
		sb.WriteString(dimStyle.Render("  No recent changes found.\n"))
		return sb.String()
	}

	now := time.Now()

	for _, change := range changes {
		ago := now.Sub(change.ModTime)
		agoStr := formatDuration(ago)

		// Time indicator
		var timeStyle lipgloss.Style
		if ago < time.Hour {
			timeStyle = lipgloss.NewStyle().Foreground(green)
		} else if ago < 24*time.Hour {
			timeStyle = lipgloss.NewStyle().Foreground(yellow)
		} else {
			timeStyle = lipgloss.NewStyle().Foreground(gray)
		}

		timeBadge := timeStyle.Render(fmt.Sprintf("%-12s", agoStr))

		// File path
		filePath := fileStyle.Render(change.Path)

		// Size
		size := dimStyle.Render(fmt.Sprintf("(%s)", formatSize(change.Size)))

		sb.WriteString(fmt.Sprintf("  %s  %s %s\n", timeBadge, filePath, size))
	}

	return sb.String()
}

func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return "just now"
	} else if d < time.Hour {
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	} else if d < 24*time.Hour {
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	} else {
		return fmt.Sprintf("%dd ago", int(d.Hours()/24))
	}
}

func formatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// RenderStats renders project statistics
func RenderStats(stats parser.ProjectStats) string {
	var sb strings.Builder

	header := headerStyle.Render("📊 PROJECT STATISTICS")
	sb.WriteString(header)
	sb.WriteString("\n\n")

	// Main stats box
	statsContent := fmt.Sprintf(`
  %s %d
  %s %d
  %s %d
  %s %d
  %s %d
`,
		labelStyle.Render("Total Files:    "), stats.TotalFiles,
		labelStyle.Render("Total Lines:    "), stats.TotalLines,
		labelStyle.Render("Packages:       "), stats.TotalPackages,
		labelStyle.Render("Functions:      "), stats.TotalFuncs,
		labelStyle.Render("Structs:        "), stats.TotalStructs)

	sb.WriteString(boxStyle.Render(statsContent))
	sb.WriteString("\n\n")

	// Languages breakdown
	if len(stats.Languages) > 0 {
		sb.WriteString(labelStyle.Render("  Languages:"))
		sb.WriteString("\n")
		for ext, count := range stats.Languages {
			bar := strings.Repeat("█", min(count, 30))
			barStyled := lipgloss.NewStyle().Foreground(cyan).Render(bar)
			sb.WriteString(fmt.Sprintf("    %-8s %s %d\n", ext, barStyled, count))
		}
		sb.WriteString("\n")
	}

	// Largest files
	if len(stats.LargestFiles) > 0 {
		sb.WriteString(labelStyle.Render("  Largest Files:"))
		sb.WriteString("\n")
		for _, f := range stats.LargestFiles {
			sb.WriteString(fmt.Sprintf("    %s %s\n",
				dimStyle.Render(fmt.Sprintf("%5d lines", f.Lines)),
				fileStyle.Render(f.Path)))
		}
	}

	return sb.String()
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// RenderFunctions renders a list of all functions
func RenderFunctions(funcs []parser.FunctionInfo) string {
	var sb strings.Builder

	header := headerStyle.Render("⚡ FUNCTIONS")
	sb.WriteString(header)
	sb.WriteString("\n\n")

	if len(funcs) == 0 {
		sb.WriteString(dimStyle.Render("  No functions found.\n"))
		return sb.String()
	}

	// Group by package
	packages := make(map[string][]parser.FunctionInfo)
	for _, fn := range funcs {
		packages[fn.Package] = append(packages[fn.Package], fn)
	}

	for pkg, fns := range packages {
		// Package header
		pkgBox := lipgloss.NewStyle().
			Foreground(white).
			Background(blue).
			Padding(0, 1).
			Render(pkg)
		sb.WriteString("  " + pkgBox + "\n\n")

		for _, fn := range fns {
			params := strings.Join(fn.Parameters, ", ")
			returns := strings.Join(fn.Returns, ", ")

			// Function signature
			sig := fmt.Sprintf("    %s(%s)",
				methodStyle.Render(fn.Name),
				dimStyle.Render(params))

			if returns != "" {
				sig += dimStyle.Render(" → " + returns)
			}

			// Location
			loc := dimStyle.Render(fmt.Sprintf(" :%d", fn.Line))

			sb.WriteString(sig + loc + "\n")
		}
		sb.WriteString("\n")
	}

	return sb.String()
}
