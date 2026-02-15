package styles

import (
	"strings"

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
	PaneStyle = lipgloss.NewStyle().
			Padding(0, 1)

	SidebarStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), false, true, false, false). // Right border
			BorderForeground(MidGray).
			Padding(0, 0)

	RequestPaneStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), false, false, true, false). // Bottom border
			BorderForeground(MidGray).
			Padding(0, 1)

	ResponsePaneStyle = lipgloss.NewStyle().
			Padding(0, 1)

	// Keep these for backward compatibility or strict focus highlighting if needed,
	// but we might just change the border color of the specific styles dynamically.
	ActiveBorderColor = PrimaryColor
	InactiveBorderColor = MidGray

	// Modal Styles
	ModalStyle = lipgloss.NewStyle().
			Border(lipgloss.ThickBorder()).
			BorderForeground(PrimaryColor).
			Padding(0, 1)

	SelectorModalStyle = lipgloss.NewStyle().
			Border(lipgloss.ThickBorder()).
			BorderForeground(PrimaryColor).
			Padding(0, 1)

	ConfirmationModalStyle = lipgloss.NewStyle().
			Border(lipgloss.ThickBorder()).
			BorderForeground(ErrorColor).
			Background(DarkGray).
			Padding(1, 1)

	CommandPaletteStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(PrimaryColor).
			Padding(0, 1)

	NotificationStyle = lipgloss.NewStyle().
				Background(ErrorColor).
				Foreground(White).
				Padding(0, 2).
				Bold(true)

	// Main Layout Style
	MainLayoutStyle = lipgloss.NewStyle().
			Padding(0, 1)

	// Item styles
	SelectedStyle = lipgloss.NewStyle().
			Foreground(PrimaryColor).
			Bold(true).
			Border(lipgloss.ThickBorder(), false, false, false, true).
			BorderForeground(PrimaryColor).
			PaddingLeft(1)

	ModalSelectedStyle = SelectedStyle.Copy()

	DimStyle = lipgloss.NewStyle().
			Foreground(Gray)

	BoldStyle = lipgloss.NewStyle().Bold(true)

	ParamStyle = lipgloss.NewStyle().
			Foreground(AccentColor).
			Bold(true)

	PaneHeaderStyle = lipgloss.NewStyle().
			Foreground(White).
			Background(MidGray).
			Padding(0, 1).
			Bold(true).
			Width(100) // Width will be overridden

	PrimaryColorStyle = lipgloss.NewStyle().Foreground(PrimaryColor)
	ErrorColorStyle   = lipgloss.NewStyle().Foreground(ErrorColor)
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

// Solidify ensures every line in s is padded to width with the style's background.
// It also ensures that internal ANSI resets (from nested styles) don't create "holes"
// by re-injecting the background sequence after every reset.
func Solidify(s string, width int, style lipgloss.Style) string {
	// Get the background escape sequence by rendering a space and stripping it.
	renderedSpace := style.Render(" ")
	spaceIdx := strings.Index(renderedSpace, " ")
	if spaceIdx == -1 {
		return style.Width(width).Render(s)
	}
	bgSeq := renderedSpace[:spaceIdx]

	lines := strings.Split(s, "\n")
	var res []string
	for _, l := range lines {
		// 1. Pad the raw line to target width
		w := lipgloss.Width(l)
		padded := l
		if w < width {
			padded += strings.Repeat(" ", width-w)
		}

		// 2. Render with style
		rendered := style.Render(padded)

		// 3. Re-inject background after every reset, but avoid at the very end
		// Handle \x1b[0m, \x1b[m, and other variants if they occur.
		resets := []string{"\x1b[0m", "\x1b[m", "\x1b[0;0m"}
		fixed := rendered
		for _, r := range resets {
			fixed = strings.ReplaceAll(fixed, r, r+bgSeq)
		}

		// 4. Clean up any trailing background sequence to prevent leakage
		fixed = strings.TrimSuffix(fixed, bgSeq)

		res = append(res, fixed)
	}
	return strings.Join(res, "\n")
}

// WithBorderTitle renders a box with the title embedded centered on the top border.
func WithBorderTitle(style lipgloss.Style, content string, titleStr string) string {
	border, _, right, bottom, left := style.GetBorder()
	borderColor := style.GetBorderTopForeground()
	if borderColor == nil {
		borderColor = PrimaryColor
	}

	// Remove top border from original style for content rendering
	contentStyle := style.Copy().Border(border, false, right, bottom, left)
	renderedContent := contentStyle.Render(content)

	// Measure ACTUAL width
	lines := strings.Split(renderedContent, "\n")
	actualWidth := 0
	if len(lines) > 0 {
		actualWidth = lipgloss.Width(lines[0])
	}

	// Prepare title
	titleStyle := lipgloss.NewStyle().Foreground(borderColor).Bold(true)
	renderedTitle := titleStyle.Render(" " + strings.TrimSpace(titleStr) + " ")
	titleWidth := lipgloss.Width(renderedTitle)

	// Construct top border
	borderStyle := lipgloss.NewStyle().Foreground(borderColor)
	
	leftLen := (actualWidth - titleWidth) / 2
	rightLen := actualWidth - titleWidth - leftLen

	topLine := borderStyle.Render(border.TopLeft + strings.Repeat(border.Top, leftLen-1)) +
		renderedTitle +
		borderStyle.Render(strings.Repeat(border.Top, rightLen-1) + border.TopRight)

	return lipgloss.JoinVertical(lipgloss.Left, topLine, renderedContent)
}
