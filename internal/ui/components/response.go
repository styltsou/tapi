// internal/ui/response_model.go
package components

import (
	"github.com/styltsou/tapi/internal/ui/commands"
	"github.com/styltsou/tapi/internal/ui/styles"

	"fmt"
	"strings"

	"github.com/alecthomas/chroma/v2/formatters"
	"github.com/alecthomas/chroma/v2/lexers"
	chromaStyles "github.com/alecthomas/chroma/v2/styles"
	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/styltsou/tapi/internal/http"
	"github.com/styltsou/tapi/internal/storage"
)

// ResponseModel handles the response viewer
type ResponseModel struct {
	Width    int
	Height   int
	loading  bool
	response *http.ProcessedResponse
	request  storage.Request
	viewport viewport.Model

	// Search state
	searchInput   textinput.Model
	searching     bool          // search bar is Visible and focused
	searchActive  bool          // matches are highlighted (bar may be dismissed)
	searchQuery   string        // current search query
	matches       []SearchMatch // all match positions
	currentMatch  int           // index of current highlighted match (0-based)
}

// SearchMatch represents a single match in the body text
type SearchMatch struct {
	StartByte int // byte offset in raw body
	EndByte   int // byte offset end (exclusive)
	Line      int // 0-based line number
}

func NewResponseModel() ResponseModel {
	vp := viewport.New(80, 20)

	si := textinput.New()
	si.Placeholder = "Search..."
	si.Width = 30

	return ResponseModel{
		viewport:    vp,
		searchInput: si,
	}
}

func (m *ResponseModel) SetSize(width, height int) {
	m.Width = width
	m.Height = height
	m.viewport.Width = width - 4
	// Adjust viewport height. We reserve a bit of space for search bar if needed,
	// but mostly it's just the pane.
	m.viewport.Height = max(1, height-2)
}

func (m *ResponseModel) SetLoading(loading bool) {
	m.loading = loading
}

func (m *ResponseModel) SetResponse(resp *http.ProcessedResponse, req storage.Request) {
	m.response = resp
	m.request = req

	// Clear any previous search
	m.clearSearch()

	// Format response content for viewport
	content := m.formatResponse()
	m.viewport.SetContent(content)
}

// clearSearch resets all search state
func (m *ResponseModel) clearSearch() {
	m.searching = false
	m.searchActive = false
	m.searchQuery = ""
	m.matches = nil
	m.currentMatch = 0
	m.searchInput.SetValue("")
}

// Clear resets the response view
func (m *ResponseModel) Clear() {
	m.response = nil
	m.request = storage.Request{}
	m.clearSearch()
	m.viewport.SetContent(styles.DimStyle.Render("No response yet. Execute a request to see results."))
}

// findMatches performs case-insensitive search on the raw body text
func findMatches(body string, query string) []SearchMatch {
	if query == "" || body == "" {
		return nil
	}

	lowerBody := strings.ToLower(body)
	lowerQuery := strings.ToLower(query)
	queryLen := len(lowerQuery)

	var matches []SearchMatch
	offset := 0
	for {
		idx := strings.Index(lowerBody[offset:], lowerQuery)
		if idx < 0 {
			break
		}
		absIdx := offset + idx
		// Calculate line number
		line := strings.Count(body[:absIdx], "\n")
		matches = append(matches, SearchMatch{
			StartByte: absIdx,
			EndByte:   absIdx + queryLen,
			Line:      line,
		})
		offset = absIdx + 1 // advance by 1 to find overlapping matches
	}
	return matches
}

// highlightMatches applies highlight markers to body text for matched regions.
// Returns the body with ANSI-colored highlights.
func highlightMatches(body string, matches []SearchMatch, currentIdx int) string {
	if len(matches) == 0 {
		return body
	}

	matchStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("#FFFF00")).
		Foreground(lipgloss.Color("#000000"))

	currentStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("#FF6600")).
		Foreground(lipgloss.Color("#FFFFFF")).
		Bold(true)

	var sb strings.Builder
	prev := 0

	for i, m := range matches {
		// Write text before this match
		if m.StartByte > prev {
			sb.WriteString(body[prev:m.StartByte])
		}

		// Write the match with highlight
		matchText := body[m.StartByte:m.EndByte]
		if i == currentIdx {
			sb.WriteString(currentStyle.Render(matchText))
		} else {
			sb.WriteString(matchStyle.Render(matchText))
		}
		prev = m.EndByte
	}

	// Write remaining text
	if prev < len(body) {
		sb.WriteString(body[prev:])
	}

	return sb.String()
}

// executeSearch searches the body and updates match state
func (m *ResponseModel) executeSearch() {
	if m.response == nil {
		return
	}
	query := m.searchInput.Value()
	m.searchQuery = query
	body := m.response.BodyString()
	m.matches = findMatches(body, query)
	m.currentMatch = 0

	if len(m.matches) > 0 {
		m.searchActive = true
	} else {
		m.searchActive = query != ""
	}

	// Re-render content with highlights
	content := m.formatResponse()
	m.viewport.SetContent(content)

	// Scroll to first match
	if len(m.matches) > 0 {
		m.scrollToMatch(0)
	}
}

// scrollToMatch scrolls the viewport to show the given match
func (m *ResponseModel) scrollToMatch(idx int) {
	if idx < 0 || idx >= len(m.matches) {
		return
	}
	match := m.matches[idx]

	// The viewport content has header stuff before the body.
	// We need to estimate the header line count to offset.
	headerLines := m.countHeaderLines()
	targetLine := headerLines + match.Line

	// Set viewport to show the target line near the top
	vpHeight := m.viewport.Height
	scrollTo := targetLine - vpHeight/3
	if scrollTo < 0 {
		scrollTo = 0
	}
	m.viewport.SetYOffset(scrollTo)
}

// countHeaderLines counts lines in the formatted response before the body content
func (m *ResponseModel) countHeaderLines() int {
	if m.response == nil {
		return 0
	}
	lines := 0
	// Status line + blank line
	lines += 2
	// Truncated warning
	if m.response.Truncated {
		lines += 2
	}
	// HEADERS label + header lines + blank
	lines += 1
	for _, values := range m.response.Headers {
		lines += len(values)
	}
	lines += 1
	// BODY label
	lines += 1
	return lines
}

func (m *ResponseModel) formatResponse() string {
	if m.response == nil {
		return styles.DimStyle.Render("No response yet. Execute a request to see results.")
	}

	var sb strings.Builder

	// Status line
	statusColor := styles.SecondaryColor
	if m.response.StatusCode >= 400 {
		statusColor = styles.ErrorColor
	}
	statusStyle := lipgloss.NewStyle().Foreground(styles.White).Background(statusColor).Bold(true).Padding(0, 1)

	sb.WriteString(statusStyle.Render(m.response.Status))
	sb.WriteString("  ")
	sb.WriteString(styles.DimStyle.Render(fmt.Sprintf("%v  %s", m.response.Duration, m.response.FormatSize())))
	sb.WriteString("\n\n")

	if m.response.Truncated {
		sb.WriteString(styles.ErrorStatusStyle.Render("⚠️  Response truncated (>10MB)") + "\n\n")
	}

	// Headers
	sb.WriteString(styles.HeaderStyle.Render("HEADERS"))
	sb.WriteString("\n")
	for key, values := range m.response.Headers {
		for _, value := range values {
			sb.WriteString(styles.DimStyle.Render(key+": ") + value + "\n")
		}
	}
	sb.WriteString("\n")

	// Body
	sb.WriteString(styles.HeaderStyle.Render("BODY"))
	sb.WriteString("\n")

	body := m.response.BodyString()
	contentType := m.response.GetHeader("Content-Type")

	// If search is active, highlight matches on raw body THEN apply syntax highlighting
	// Note: We highlight on raw text because syntax highlighting adds ANSI codes
	// that would interfere with match positions
	if m.searchActive && len(m.matches) > 0 {
		// Apply search highlighting (on raw body, no syntax highlighting to avoid conflicts)
		sb.WriteString(highlightMatches(body, m.matches, m.currentMatch))
	} else {
		highlighted := highlight(body, contentType)
		sb.WriteString(highlighted)
	}

	return sb.String()
}

func highlight(content string, contentType string) string {
	lexer := lexers.Get(contentType)
	if lexer == nil {
		lexer = lexers.Analyse(content)
	}
	if lexer == nil {
		lexer = lexers.Fallback
	}

	style := chromaStyles.Get("monokai")
	if style == nil {
		style = chromaStyles.Fallback
	}

	formatter := formatters.TTY256
	iterator, err := lexer.Tokenise(nil, content)
	if err != nil {
		return content
	}

	var sb strings.Builder
	err = formatter.Format(&sb, style, iterator)
	if err != nil {
		return content
	}

	return sb.String()
}

func (m ResponseModel) Update(msg tea.Msg) (ResponseModel, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// While search input is focused
		if m.searching {
			switch msg.String() {
			case "esc":
				// Close search completely
				m.clearSearch()
				content := m.formatResponse()
				m.viewport.SetContent(content)
				return m, nil

			case "enter":
				// Dismiss search bar but keep highlights
				m.searching = false
				m.searchInput.Blur()
				return m, nil

			default:
				// Update search input
				m.searchInput, cmd = m.searchInput.Update(msg)
				// Live search on every keystroke
				m.executeSearch()
				return m, cmd
			}
		}

		// Normal mode (search bar not focused)
		switch msg.String() {
		case "esc", "q":
			if m.searchActive {
				// Clear search highlights
				m.clearSearch()
				content := m.formatResponse()
				m.viewport.SetContent(content)
				return m, nil
			}
			return m, nil

		case "/":
			if m.response != nil {
				m.searching = true
				m.searchInput.Focus()
				m.searchInput.SetValue("")
				return m, nil
			}

		case "n":
			if m.searchActive && len(m.matches) > 0 {
				m.currentMatch = (m.currentMatch + 1) % len(m.matches)
				content := m.formatResponse()
				m.viewport.SetContent(content)
				m.scrollToMatch(m.currentMatch)
				return m, nil
			}

		case "N":
			if m.searchActive && len(m.matches) > 0 {
				m.currentMatch = (m.currentMatch - 1 + len(m.matches)) % len(m.matches)
				content := m.formatResponse()
				m.viewport.SetContent(content)
				m.scrollToMatch(m.currentMatch)
				return m, nil
			}

		case "c":
			if m.response != nil {
				err := clipboard.WriteAll(m.response.BodyString())
				if err != nil {
					return m, commands.ShowStatusCmd("Failed to copy body", true)
				}
				return m, commands.ShowStatusCmd("Body copied to clipboard", false)
			}
		}
	}

	// Update viewport
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func (m ResponseModel) View() string {
	if m.loading {
		return styles.DimStyle.Render("Loading...\n\nExecuting request...")
	}

	if m.response == nil {
		return styles.DimStyle.Render("No response yet.\nExecute a request to see the response.")
	}

	var view strings.Builder

	// Search bar (shown when searching or when matches are active)
	if m.searching || m.searchActive {
		searchBar := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(styles.SecondaryColor).
			Padding(0, 1).
			Width(m.Width - 6)

		var barContent string
		if m.searching {
			barContent = "🔍 " + m.searchInput.View()
		} else {
			barContent = styles.DimStyle.Render("🔍 " + m.searchQuery)
		}

		// Match counter
		if m.searchActive && len(m.matches) > 0 {
			counter := fmt.Sprintf("  %d/%d", m.currentMatch+1, len(m.matches))
			barContent += styles.DimStyle.Render(counter)
		} else if m.searchActive && m.searchQuery != "" {
			barContent += styles.ErrorStatusStyle.Render("  No matches")
		}

		view.WriteString(searchBar.Render(barContent))
		view.WriteString("\n")
	}

	view.WriteString(m.viewport.View())

	return view.String()
}
