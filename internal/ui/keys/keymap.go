// internal/ui/keymap.go
package keys

import "github.com/charmbracelet/bubbles/key"

// KeyMap defines all key bindings
type KeyMap struct {
	Quit      key.Binding
	AddRow    key.Binding
	DeleteRow key.Binding
}

func DefaultKeyMap() KeyMap {
	return KeyMap{
		Quit: key.NewBinding(
			key.WithKeys("ctrl+c"),
			key.WithHelp("ctrl+c", "quit"),
		),
		AddRow: key.NewBinding(
			key.WithKeys("ctrl+a"),
			key.WithHelp("ctrl+a", "add row"),
		),
		DeleteRow: key.NewBinding(
			key.WithKeys("ctrl+d"),
			key.WithHelp("ctrl+d", "delete row"),
		),
	}
}

func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{
		key.NewBinding(key.WithKeys(""), key.WithHelp("SPC", "leader")),
		key.NewBinding(key.WithKeys(""), key.WithHelp("i", "insert")),
		key.NewBinding(key.WithKeys(""), key.WithHelp("ESC", "normal")),
		k.Quit,
	}
}

func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{
			key.NewBinding(key.WithKeys(""), key.WithHelp("SPC+r", "run")),
			key.NewBinding(key.WithKeys(""), key.WithHelp("SPC+s", "save")),
			key.NewBinding(key.WithKeys(""), key.WithHelp("SPC+e", "sidebar")),
			key.NewBinding(key.WithKeys(""), key.WithHelp("SPC+v", "envs")),
		},
		{
			key.NewBinding(key.WithKeys(""), key.WithHelp("SPC+c", "collection")),
			key.NewBinding(key.WithKeys(""), key.WithHelp("SPC+o", "preview")),
			key.NewBinding(key.WithKeys(""), key.WithHelp("SPC+m", "menu")),
			k.Quit,
		},
	}
}
