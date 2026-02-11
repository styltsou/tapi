# TAPI — Terminal API Client

> A keyboard-driven, Vim-style terminal API client built in Go with [Bubble Tea](https://github.com/charmbracelet/bubbletea).

```
 ████████╗ █████╗ ██████╗ ██╗
 ╚══██╔══╝██╔══██╗██╔══██╗██║
    ██║   ███████║██████╔╝██║
    ██║   ██╔══██║██╔═══╝ ██║
    ██║   ██║  ██║██║     ██║
    ╚═╝   ╚═╝  ╚═╝╚═╝     ╚═╝
```

## Features

- **Welcome Screen** — LazyVim-style dashboard with recent collections and quick actions
- **Collection Management** — Organize requests into collections stored as YAML in `~/.tapi/collections/`
- **Import/Export** — Import collections from Postman (v2.1), Insomnia (v4), or cURL commands. Export as YAML
- **Request Builder** — Full control over HTTP methods, URL, Headers, Params, Body, and Query Params
- **Response Viewer** — Syntax-highlighted JSON, collapsible headers, copy/save body
- **Environment Variables** — Manage environments in `~/.tapi/environments/`, use `{{var}}` syntax with autocomplete and validation
- **Vim-Style Modes** — Normal and Insert modes with a `Space` leader key for commands
- **Local-First** — All data stored as simple YAML files on your machine

## Installation

Requires **Go 1.21+**.

```bash
git clone https://github.com/styltsou/tapi
cd tapi
go build -o tapi ./cmd/tapi
sudo mv tapi /usr/local/bin/
```

On first run, a demo collection is created automatically so you can start exploring immediately.

## Usage

```bash
tapi
```

You'll land on the **welcome screen** showing your existing collections. Select one or create a new one to enter the workspace.

## Modes

TAPI uses Vim-style **Normal** and **Insert** modes. The current mode is shown in the status bar.

| Mode | Purpose | Enter | Exit |
|------|---------|-------|------|
| **NORMAL** | Navigate, run commands, leader chords | `Esc` | — |
| **INSERT** | Edit text fields (URL, body, headers) | `i` or `Enter` | `Esc` |
| **LEADER** | Waiting for chord key after `Space` | `Space` | auto |

## Keybindings

### Leader Chords (Normal mode: `Space` + key)

| Chord | Action |
|-------|--------|
| `Space e` | Toggle sidebar |
| `Space r` | Run request |
| `Space s` | Save request |
| `Space c` | Change collection |
| `Space v` | Toggle environments |
| `Space o` | Toggle preview (variable substitution) |
| `Space p` | Focus request pane |
| `Space k` | Open command menu |
| `Space q` | Quit |

### Navigation (Normal mode)

| Key | Action |
|-----|--------|
| `Tab` / `Shift+Tab` | Cycle between panes |
| `j` / `k` | Navigate lists |
| `←` / `→` | Cycle HTTP method (on method field) |
| `i` / `Enter` | Enter Insert mode |

### Editing (Insert mode)

| Key | Action |
|-----|--------|
| `Esc` | Return to Normal mode |
| `Tab` / `Shift+Tab` | Cycle fields within request pane |
| `Ctrl+a` | Add row (Headers / Query Params) |
| `Ctrl+d` | Delete row |
| `{{` | Autocomplete environment variable |

### Collections Sidebar

| Key | Action |
|-----|--------|
| `Enter` | Select request / Create new |
| `d` | Delete request (with confirmation) |
| `D` | Delete collection (with confirmation) |
| `y` | Duplicate request |
| `r` | Rename collection |

### Response Pane

| Key | Action |
|-----|--------|
| `j` / `k` | Scroll |
| `h` | Toggle headers |
| `c` | Copy body |

### Global

| Key | Action |
|-----|--------|
| `Ctrl+c` | Quit |

## Data Storage

```
~/.tapi/
├── collections/     # YAML files, one per collection
└── environments/    # YAML files, one per environment
```

## Import / Export

TAPI can import collections from other tools via the command menu (`Space k` → "Import Collection"):

- **Postman** — v2.1 JSON exports (folders are flattened)
- **Insomnia** — v4 JSON exports
- **cURL** — Paste a cURL command into a `.txt` file

To export, select "Export Collection" from the command menu — it saves the current collection as a portable YAML file.

## Tech Stack

- [Go](https://go.dev)
- [Bubble Tea](https://github.com/charmbracelet/bubbletea) — TUI framework
- [Lip Gloss](https://github.com/charmbracelet/lipgloss) — Styling
- [Chroma](https://github.com/alecthomas/chroma) — Syntax highlighting

## License

MIT
