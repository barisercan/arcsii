# arcsii

Beautiful terminal-based code architecture visualizer with ASCII art.

```
    ╔═══════════════════════════════════════════════════════════╗
    ║                                                           ║
    ║     ▄▀▄ █▀▄ ▄▀▀ ▄▀▀ █ █                                   ║
    ║     █▀█ █▀▄ █   ▀▀█ █ █                                   ║
    ║     ▀ ▀ ▀ ▀  ▀▀ ▀▀▀ ▀ ▀                                   ║
    ║                                                           ║
    ║         Terminal Architecture Visualizer                  ║
    ║                                                           ║
    ╚═══════════════════════════════════════════════════════════╝
```

## Installation

### Homebrew (macOS)

```bash
brew install barisercan/tap/arcsii
```

### Download Binary

Download the latest release from [GitHub Releases](https://github.com/barisercan/arcsii/releases).

### Build from Source

```bash
go install github.com/barisercan/arcsii@latest
```

## Usage

```bash
# Run in current directory
arcsii

# Run on a specific project
arcsii /path/to/project
```

## Commands

| Command | Aliases | Description |
|---------|---------|-------------|
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
- `↑↓` - Scroll content
- `Esc` / `Ctrl+C` - Quit

## Screenshots

### File Tree (`/tree`)
```
📂 arcsii/
├── 📂 internal/
│   ├── 📂 commands/
│   │   └── 🔷 registry.go
│   ├── 📂 parser/
│   │   └── 🔷 parser.go
│   ├── 📂 renderer/
│   │   └── 🔷 renderer.go
│   └── 📂 ui/
│       └── 🔷 model.go
├── 🔷 main.go
└── 📦 go.mod
```

### UML Class Diagram (`/uml`)
```
╭──────────────────────────────────────╮
│ Model                    pkg: ui     │
│──────────────────────────────────────│
│ Fields:                              │
│   targetDir string                   │
│   input textinput.Model              │
│   viewport viewport.Model            │
│                                      │
│ Methods:                             │
│   Init() → Cmd                       │
│   Update(Msg) → Model, Cmd           │
│   View() → string                    │
╰──────────────────────────────────────╯
```

### Dependency Graph (`/deps`)
```
  ui
  ├── github.com/barisercan/arcsii/internal/commands
  ├── github.com/barisercan/arcsii/internal/renderer
  ├── github.com/charmbracelet/bubbletea
  └── github.com/charmbracelet/lipgloss
```

## License

MIT
