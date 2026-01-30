package commands

import (
	"fmt"
	"strings"

	"github.com/barisercan/arcsii/internal/parser"
	"github.com/barisercan/arcsii/internal/renderer"
)

type Command struct {
	Name        string
	Aliases     []string
	Description string
	Handler     func(args []string) (string, string)
}

type Registry struct {
	targetDir string
	commands  map[string]*Command
}

func NewRegistry(targetDir string) *Registry {
	r := &Registry{
		targetDir: targetDir,
		commands:  make(map[string]*Command),
	}
	r.registerCommands()
	return r
}

func (r *Registry) registerCommands() {
	// Help command
	r.register(&Command{
		Name:        "help",
		Aliases:     []string{"h", "?"},
		Description: "Show available commands",
		Handler: func(args []string) (string, string) {
			return renderer.RenderHelp(), "Showing help"
		},
	})

	// Tree command - file structure
	r.register(&Command{
		Name:        "tree",
		Aliases:     []string{"t", "files"},
		Description: "Show file tree structure",
		Handler: func(args []string) (string, string) {
			tree := parser.ParseFileTree(r.targetDir)
			return renderer.RenderTree(tree), "File tree"
		},
	})

	// UML command - class diagrams
	r.register(&Command{
		Name:        "uml",
		Aliases:     []string{"class", "classes"},
		Description: "Show UML class diagram",
		Handler: func(args []string) (string, string) {
			classes := parser.ParseClasses(r.targetDir)
			return renderer.RenderUML(classes), "UML diagram"
		},
	})

	// ASCII art visualization
	r.register(&Command{
		Name:        "ascii",
		Aliases:     []string{"art", "a"},
		Description: "ASCII art architecture view",
		Handler: func(args []string) (string, string) {
			structure := parser.ParseStructure(r.targetDir)
			return renderer.RenderASCIIArt(structure), "ASCII art view"
		},
	})

	// Dependencies
	r.register(&Command{
		Name:        "deps",
		Aliases:     []string{"dependencies", "d"},
		Description: "Show dependency graph",
		Handler: func(args []string) (string, string) {
			deps := parser.ParseDependencies(r.targetDir)
			return renderer.RenderDeps(deps), "Dependencies"
		},
	})

	// Latest changes
	r.register(&Command{
		Name:        "changes",
		Aliases:     []string{"recent", "modified"},
		Description: "Show recently modified files",
		Handler: func(args []string) (string, string) {
			changes := parser.ParseRecentChanges(r.targetDir)
			return renderer.RenderChanges(changes), "Recent changes"
		},
	})

	// Stats command
	r.register(&Command{
		Name:        "stats",
		Aliases:     []string{"info", "summary"},
		Description: "Show project statistics",
		Handler: func(args []string) (string, string) {
			stats := parser.ParseStats(r.targetDir)
			return renderer.RenderStats(stats), "Project stats"
		},
	})

	// Functions command
	r.register(&Command{
		Name:        "funcs",
		Aliases:     []string{"functions", "fn"},
		Description: "List all functions/methods",
		Handler: func(args []string) (string, string) {
			funcs := parser.ParseFunctions(r.targetDir)
			return renderer.RenderFunctions(funcs), "Functions"
		},
	})
}

func (r *Registry) register(cmd *Command) {
	r.commands[cmd.Name] = cmd
	for _, alias := range cmd.Aliases {
		r.commands[alias] = cmd
	}
}

func (r *Registry) Execute(input string) (string, string) {
	input = strings.TrimPrefix(input, "/")
	parts := strings.Fields(input)

	if len(parts) == 0 {
		return renderer.RenderWelcome(), "Ready"
	}

	cmdName := strings.ToLower(parts[0])
	args := parts[1:]

	if cmd, ok := r.commands[cmdName]; ok {
		return cmd.Handler(args)
	}

	return fmt.Sprintf("Unknown command: %s\n\nType /help for available commands", cmdName), "Unknown command"
}
