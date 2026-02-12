package ui

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/styltsou/tapi/internal/config"
)

// ApplyTheme overrides the default color palette from config
func ApplyTheme(theme config.ThemeConfig) {
	if theme.Primary != "" {
		PrimaryColor = lipgloss.Color(theme.Primary)
	}
	if theme.Secondary != "" {
		SecondaryColor = lipgloss.Color(theme.Secondary)
	}
	if theme.Accent != "" {
		AccentColor = lipgloss.Color(theme.Accent)
	}
	if theme.Error != "" {
		ErrorColor = lipgloss.Color(theme.Error)
	}
}

var (
	// Colors
	PrimaryColor   = lipgloss.Color("#7D56F4") // Indigo/Purple
	SecondaryColor = lipgloss.Color("#04B575") // Green
	AccentColor    = lipgloss.Color("#EE6FF8") // Pink/Magenta
	White          = lipgloss.Color("#FFFFFF")
	Gray           = lipgloss.Color("#777777")
	MidGray        = lipgloss.Color("#444444")
	DarkGray       = lipgloss.Color("#222222")
	Black          = lipgloss.Color("#000000")
	ErrorColor     = lipgloss.Color("#FF4C4C")

	// Method Colors
	MethodGetColor    = lipgloss.Color("#04B575") // Green
	MethodPostColor   = lipgloss.Color("#7D56F4") // Violet
	MethodPutColor    = lipgloss.Color("#FFA500") // Orange
	MethodDeleteColor = lipgloss.Color("#FF4C4C") // Red
	MethodPatchColor  = lipgloss.Color("#FFA500") // Orange (same as PUT)
	MethodOptionsColor = lipgloss.Color("#ADD8E6") // Light Blue

	// Styles
	TitleStyle = lipgloss.NewStyle().
			Foreground(White).
			Background(PrimaryColor).
			Padding(0, 1).
			Bold(true)

	HeaderStyle = lipgloss.NewStyle().
			Foreground(PrimaryColor).
			Bold(true).
			Padding(0, 1)

	FooterStyle = lipgloss.NewStyle().
			Foreground(Gray).
			Padding(0, 1)

	StatusStyle = lipgloss.NewStyle().
			Foreground(White).
			Background(DarkGray).
			Padding(0, 1)

	ErrorStatusStyle = StatusStyle.Background(ErrorColor)

	// Status Bar Styles
	StatusBarLogoStyle = lipgloss.NewStyle().
			Foreground(White).
			Background(PrimaryColor).
			Padding(0, 1).
			Bold(true)

	StatusBarContextStyle = lipgloss.NewStyle().
			Foreground(White).
			Background(MidGray).
			Padding(0, 1)

	StatusBarInfoStyle = lipgloss.NewStyle().
			Foreground(Gray).
			Background(DarkGray).
			Padding(0, 1)

	// Dashboard Pane Styles
	ActivePaneStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(PrimaryColor).
			Padding(0, 1)

	InactivePaneStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(MidGray).
			Padding(0, 1)

	// Modal Styles
	ModalStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(PrimaryColor).
			Padding(1, 2).
			Background(DarkGray)

	// Main Layout Style
	MainLayoutStyle = lipgloss.NewStyle().
			Padding(1, 2)

	// Item styles
	SelectedStyle = lipgloss.NewStyle().
			Foreground(PrimaryColor).
			Bold(true).
			Border(lipgloss.ThickBorder(), false, false, false, true).
			BorderForeground(PrimaryColor).
			PaddingLeft(1)

	DimStyle = lipgloss.NewStyle().
			Foreground(Gray)

	BoldStyle = lipgloss.NewStyle().Bold(true)

	ParamStyle = lipgloss.NewStyle().
			Foreground(AccentColor).
			Bold(true)
)

// MethodBadge returns a styled badge for an HTTP method
func MethodBadge(method string) string {
	style := lipgloss.NewStyle().
		Bold(true).
		Padding(0, 1).
		Foreground(White)

	switch method {
	case "GET":
		style = style.Background(MethodGetColor)
	case "POST":
		style = style.Background(MethodPostColor)
	case "PUT":
		style = style.Background(MethodPutColor)
	case "DELETE":
		style = style.Background(MethodDeleteColor)
	case "PATCH":
		style = style.Background(MethodPatchColor)
	case "OPTIONS":
		style = style.Background(MethodOptionsColor).Foreground(Black) // Dark text for light background
	case "HEAD":
		style = style.Background(Gray).Foreground(White)
	default:
		style = style.Background(Gray)
	}

	return style.Render(method)
}
