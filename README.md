# arcsii

Beautiful terminal-based code architecture visualizer with ASCII art, live file monitoring, and git operation animations.

![arcsii demo](arcsii-demo.gif)

## Features

- **Live File Monitor** - Watch file changes in real-time with animated previews
- **Git Operation Animations** - Cool ASCII art animations for commit, push, pull, merge, rebase, and more
- **Multi-Language Support** - Works with Go, Java, Python, TypeScript, JavaScript, Swift, Kotlin, C#, Rust
- **ASCII Architecture View** - Beautiful ASCII art visualization of your project structure
- **UML Class Diagrams** - View classes, structs, interfaces, and their relationships
- **Dependency Graphs** - See your project's import dependencies
- **Command History** - Use `↑↓` to cycle through commands

## Installation

### Homebrew (macOS/Linux)

```bash
brew install barisercan/tap/arcsii
```

### Download Binary

Download the latest release from [GitHub Releases](https://github.com/barisercan/arcsii/releases).

Available for:
- macOS (Apple Silicon & Intel)
- Linux (amd64 & arm64)
- Windows (amd64 & arm64)

### Build from Source

```bash
go install github.com/barisercan/arcsii@latest
```

## Usage

```bash
# Run in current directory (starts in live watch mode)
arcsii

# Run on a specific project
arcsii /path/to/project
```

## Commands

| Command | Aliases | Description |
|---------|---------|-------------|
| `/watch` | `/live`, `/w` | Live file monitor mode (default) |
| `/tree` | `/t`, `/files` | Show file tree structure |
| `/uml` | `/class`, `/classes` | Show UML class diagram |
| `/ascii` | `/art`, `/a` | ASCII art architecture view |
| `/deps` | `/dependencies`, `/d` | Show dependency graph |
| `/changes` | `/recent`, `/modified` | Show recently modified files |
| `/stats` | `/info`, `/summary` | Show project statistics |
| `/funcs` | `/functions`, `/fn` | List all functions/methods |
| `/help` | `/h`, `/?` | Show help |

## Controls

- `Enter` - Execute command
- `↑↓` - Cycle through command history
- `Esc` / `Ctrl+C` - Quit

## Live File Monitor

The default mode watches your project for file changes in real-time:

```
◐ LIVE FILE MONITOR

    ✎ modified   🔷  internal/ui/model.go  2s ago
        │ return sb.String()
        │ }
        │ func (m Model) renderEvent...

    ✚ created    📄  new-file.txt  5s ago
        │ Hello World
```

## Git Animations

When you perform git operations, arcsii shows animated ASCII art:

**Commit:**
```
╔═══════════════════════════════════════════════════════╗
║      ██████╗ ██████╗ ███╗   ███╗███╗   ███╗██╗████████╗║
║     ██╔════╝██╔═══██╗████╗ ████║████╗ ████║██║╚══██╔══╝║
║     ██║     ██║   ██║██╔████╔██║██╔████╔██║██║   ██║   ║
║     ╚██████╗╚██████╔╝██║ ╚═╝ ██║██║ ╚═╝ ██║██║   ██║   ║
║              [  ✓  ]  Changes saved!                  ║
╚═══════════════════════════════════════════════════════╝
```

Also supports: **push**, **pull**, **merge**, **checkout**, **rebase**, **stash**

## ASCII Architecture View

```
╔═══════════════════════════════════════════════════════════════════════════╗
║     █████╗ ██████╗  ██████╗██╗  ██╗██╗████████╗███████╗ ██████╗████████╗  ║
║    ██╔══██╗██╔══██╗██╔════╝██║  ██║██║╚══██╔══╝██╔════╝██╔════╝╚══██╔══╝  ║
║    ███████║██████╔╝██║     ███████║██║   ██║   █████╗  ██║        ██║     ║
╚═══════════════════════════════════════════════════════════════════════════╝

    ┌─────────────────────────────────────────────────────────────────┐
    │  📊 SYSTEM OVERVIEW                                             │
    │    Modules: 4       Classes: 12      Functions: 45              │
    └─────────────────────────────────────────────────────────────────┘

    ╔══════════════════════════════════════════════════════════╗
    ║                    ◈ COMMANDS ◈                          ║
    ╠══════════════════════════════════════════════════════════╣
    ║  ◆ Classes/Structs                                       ║
    ║    └── Registry                                          ║
    ║  ƒ Functions                                             ║
    ║    └── NewRegistry                                       ║
    ║  ◈ Files                                                 ║
    ║    └── 🔷 registry.go                                    ║
    ╚══════════════════════════════════════════════════════════╝
```

## Supported Languages

| Language | Extensions | Features |
|----------|------------|----------|
| Go | `.go` | Classes, functions, imports |
| Java | `.java` | Classes, interfaces, methods, imports |
| Kotlin | `.kt`, `.kts` | Classes, interfaces, functions, imports |
| Python | `.py` | Classes, functions, imports |
| TypeScript | `.ts`, `.tsx` | Classes, interfaces, functions, imports |
| JavaScript | `.js`, `.jsx`, `.mjs` | Classes, functions, imports |
| Swift | `.swift` | Classes, structs, protocols, functions, imports |
| C# | `.cs` | Classes, interfaces, structs, methods, imports |
| Rust | `.rs` | Structs, traits, functions, imports |

## Built With

- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - TUI framework
- [Lip Gloss](https://github.com/charmbracelet/lipgloss) - Style definitions
- [fsnotify](https://github.com/fsnotify/fsnotify) - File system notifications

## License

MIT
