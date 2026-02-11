// internal/ui/response_model.go
package ui

import (
	"fmt"
	"strings"

	"github.com/alecthomas/chroma/v2/formatters"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/styltsou/tapi/internal/http"
	"github.com/styltsou/tapi/internal/storage"
)

// ResponseModel handles the response viewer
type ResponseModel struct {
	width    int
	height   int
	loading  bool
	response *http.ProcessedResponse
	request  storage.Request
	viewport viewport.Model
}

func NewResponseModel() ResponseModel {
	vp := viewport.New(80, 20)
	return ResponseModel{
		viewport: vp,
	}
}

func (m *ResponseModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.viewport.Width = width - 4
	m.viewport.Height = height - 10
}

func (m *ResponseModel) SetLoading(loading bool) {
	m.loading = loading
}

func (m *ResponseModel) SetResponse(resp *http.ProcessedResponse, req storage.Request) {
	m.response = resp
	m.request = req

	// Format response content for viewport
	content := m.formatResponse()
	m.viewport.SetContent(content)
}

func (m *ResponseModel) formatResponse() string {
	if m.response == nil {
		return DimStyle.Render("No response yet. Execute a request to see results.")
	}

	var sb strings.Builder

	// Status line
	statusColor := SecondaryColor
	if m.response.StatusCode >= 400 {
		statusColor = ErrorColor
	}
	statusStyle := lipgloss.NewStyle().Foreground(White).Background(statusColor).Bold(true).Padding(0, 1)

	sb.WriteString(statusStyle.Render(m.response.Status))
	sb.WriteString("  ")
	sb.WriteString(DimStyle.Render(fmt.Sprintf("%v  %s", m.response.Duration, m.response.FormatSize())))
	sb.WriteString("\n\n")

	if m.response.Truncated {
		sb.WriteString(ErrorStatusStyle.Render("⚠️  Response truncated (>10MB)") + "\n\n")
	}

	// Headers
	sb.WriteString(HeaderStyle.Render("HEADERS"))
	sb.WriteString("\n")
	for key, values := range m.response.Headers {
		for _, value := range values {
			sb.WriteString(DimStyle.Render(key+": ") + value + "\n")
		}
	}
	sb.WriteString("\n")

	// Body
	sb.WriteString(HeaderStyle.Render("BODY"))
	sb.WriteString("\n")
	
	body := m.response.BodyString()
	contentType := m.response.GetHeader("Content-Type")

	highlighted := highlight(body, contentType)
	sb.WriteString(highlighted)

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

	style := styles.Get("monokai")
	if style == nil {
		style = styles.Fallback
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
		switch msg.String() {
		case "esc", "q":
			// No-op in dashboard mode, handled by main model focusing
			return m, nil
		
		case "c":
			if m.response != nil {
				err := clipboard.WriteAll(m.response.BodyString())
				if err != nil {
					return m, showStatusCmd("Failed to copy body", true)
				}
				return m, showStatusCmd("Body copied to clipboard", false)
			}
		}
	}

	// Update viewport
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func (m ResponseModel) View() string {
	if m.loading {
		return DimStyle.Render("Loading...\n\nExecuting request...")
	}

	if m.response == nil {
		return DimStyle.Render("No response yet.\nExecute a request to see the response.")
	}

	return m.viewport.View()
}
