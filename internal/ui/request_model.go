package ui

import (
	"net/url"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/styltsou/tapi/internal/storage"
)

type RequestSection int

const (
	SectionURL RequestSection = iota
	SectionPathParams
	SectionQueryParams
	SectionHeaders
	SectionBody
)

// KVInput is a generic key-value pair of text inputs
type KVInput struct {
	key   textinput.Model
	value textinput.Model
}

// RequestModel handles the request builder view
type RequestModel struct {
	width  int
	height int

	// Current request being edited
	request storage.Request
	baseURL string

	// Static Fields
	methods     []string
	methodIndex int
	pathInput   textinput.Model
	bodyInput   textarea.Model

	// Dynamic Fields
	pathParamsInputs []KVInput
	headerInputs     []KVInput
	queryInputs      []KVInput

	// State
	focusedSection RequestSection
	focusedIndex   int // index within a section's list (e.g., which header)
	isKeyFocused   bool // true if key is focused, false if value is focused in KV sections
	vars           map[string]string
	preview        bool
	
	// Autocomplete
	suggestions SuggestionModel
}

func NewRequestModel() RequestModel {
	pathInput := textinput.New()
	pathInput.Placeholder = "/api/endpoint"
	pathInput.CharLimit = 200
	pathInput.Width = 50

	bodyInput := textarea.New()
	bodyInput.Placeholder = "Body..."
	bodyInput.SetWidth(50)
	bodyInput.SetHeight(10)

	return RequestModel{
		methods:          []string{"GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS"},
		methodIndex:      0,
		pathInput:        pathInput,
		bodyInput:        bodyInput,
		focusedSection:   SectionURL,
		focusedIndex:     1, // Start on URL
		pathParamsInputs: []KVInput{},
		headerInputs:     []KVInput{newEmptyKVInput()},
		queryInputs:      []KVInput{newEmptyKVInput()},
		suggestions:      NewSuggestionModel(),
	}
}

func newEmptyKVInput() KVInput {
	ki := textinput.New()
	ki.Placeholder = "Key"
	ki.Width = 20
	vi := textinput.New()
	vi.Placeholder = "Value"
	vi.Width = 40
	return KVInput{key: ki, value: vi}
}

func (m *RequestModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.bodyInput.SetWidth(width - 4)
}

func (m *RequestModel) LoadRequest(req storage.Request, baseURL string) {
	m.request = req
	m.baseURL = baseURL

	for i, mthd := range m.methods {
		if mthd == req.Method {
			m.methodIndex = i
			break
		}
	}

	// Parse URL for query params and path params
	m.pathInput.SetValue(req.URL)
	m.parseURLParams()

	m.bodyInput.SetValue(req.Body)

	// Load headers
	m.headerInputs = []KVInput{}
	for key, value := range req.Headers {
		m.addRow(&m.headerInputs, key, value)
	}

	// Add empty rows if needed
	if len(m.headerInputs) == 0 {
		m.addRow(&m.headerInputs, "", "")
	}
	// Path param rows are managed by parseURLParams
	// Query params rows are managed by parseURLParams

	m.focusedSection = SectionURL
	m.focusedIndex = 1 // Focus URL
	m.updateFocus()
}

func (m *RequestModel) addRow(list *[]KVInput, k, v string) {
	ki := textinput.New()
	ki.SetValue(k)
	ki.Placeholder = "Key"
	ki.Width = 20

	vi := textinput.New()
	vi.SetValue(v)
	vi.Placeholder = "Value"
	vi.Width = 40

	*list = append(*list, KVInput{key: ki, value: vi})
}

// parseURLParams parses the current URL in pathInput and updates pathParamsInputs/queryInputs
func (m *RequestModel) parseURLParams() {
	rawURL := m.pathInput.Value()

	// 1. Parse Path Params (:param)
	// We extract params but keep the existing values if defined
	newPathParams := []KVInput{}
	segments := strings.Split(rawURL, "/")
	for _, seg := range segments {
		if strings.HasPrefix(seg, ":") && len(seg) > 1 {
			key := seg[1:]
			// Check if we already have a value for this key
			val := ""
			for _, old := range m.pathParamsInputs {
				if old.key.Value() == key {
					val = old.value.Value()
					break
				}
			}
			
			// Create new input pair
			ki := textinput.New()
			ki.SetValue(key)
			ki.Placeholder = "Key"
			ki.Width = 20
			// Key is read-only essentially as it comes from URL
			
			vi := textinput.New()
			vi.SetValue(val)
			vi.Placeholder = "Value"
			vi.Width = 40
			
			newPathParams = append(newPathParams, KVInput{key: ki, value: vi})
		}
	}
	m.pathParamsInputs = newPathParams

	// 2. Parse Query Params
	// This is trickier because we want bidirectional sync.
	// If the user typed in the URL, we update the list.
	// If the user typed in the list, we update the URL (handled elsewhere).
	// Here we assume URL is source of truth.
	
	// Check if URL has query
	if idx := strings.Index(rawURL, "?"); idx != -1 {
		// queryStr := rawURL[idx+1:] // Unused
		// Simple split for now, robust parsing requires net/url
		// But robust parsing might mess up partial capabilities while typing.
		// Let's use net/url for robustness but handle errors gracefully.
		
		// To parse query properly, we need a valid URL structure.
		// dummy scheme if missing
		parseURL := rawURL
		if !strings.HasPrefix(parseURL, "http") {
			parseURL = "http://example.com/" + strings.TrimLeft(parseURL, "/")
		}
		
		u, err := url.Parse(parseURL)
		if err == nil {
			q := u.Query()
			newQueryParams := []KVInput{}
			for k, vs := range q {
				for _, v := range vs {
					ki := textinput.New()
					ki.SetValue(k)
					ki.Placeholder = "Key"
					ki.Width = 20
					
					vi := textinput.New()
					vi.SetValue(v)
					vi.Placeholder = "Value"
					vi.Width = 40
					
					newQueryParams = append(newQueryParams, KVInput{key: ki, value: vi})
				}
			}
			m.queryInputs = newQueryParams
		}
	} else {
		m.queryInputs = []KVInput{}
	}
	
	if len(m.queryInputs) == 0 {
		m.addRow(&m.queryInputs, "", "")
	}
}

// syncURLFromParams reconstructs the URL from the path and query inputs
func (m *RequestModel) syncURLFromParams() {
	// 1. Get base path from current URL mechanism involves splitting by ?
	rawURL := m.pathInput.Value()
	pathOnly := rawURL
	if idx := strings.Index(rawURL, "?"); idx != -1 {
		pathOnly = rawURL[:idx]
	}
	
	// We don't substitute path params back into the URL string KEY (:key),
	// we keep them as :key in the URL.
	// But we DO reconstruct the query string.
	
	if len(m.queryInputs) > 0 {
		vals := url.Values{}
		hasQuery := false
		for _, qi := range m.queryInputs {
			k := qi.key.Value()
			v := qi.value.Value()
			if k != "" {
				vals.Add(k, v)
				hasQuery = true
			}
		}
		
		if hasQuery {
			pathOnly += "?" + vals.Encode()
		}
	}
	
	// Update the URL input without triggering a parse loop
	// We need a flag or just be careful. 
	// The Update loop calls parseURLParams only on URL input change.
	// Here we are programmatically changing it.
	m.pathInput.SetValue(pathOnly)
}


func (m RequestModel) Update(msg tea.Msg) (RequestModel, tea.Cmd) {
	var cmds []tea.Cmd

	// 1. Handle Autocomplete
	if m.suggestions.visible {
		newSugg, cmd := m.suggestions.Update(msg)
		m.suggestions = newSugg
		cmds = append(cmds, cmd)
		return m, tea.Batch(cmds...) // block other input while suggestions open
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+a": // Add row
			if m.focusedSection == SectionHeaders {
				m.addRow(&m.headerInputs, "", "")
				m.focusedIndex = len(m.headerInputs) - 1
				m.isKeyFocused = true
			} else if m.focusedSection == SectionQueryParams {
				m.addRow(&m.queryInputs, "", "")
				m.focusedIndex = len(m.queryInputs) - 1
				m.isKeyFocused = true
			}
			m.updateFocus()
			return m, nil

		case "ctrl+d": // Delete row
			if m.focusedSection == SectionHeaders && len(m.headerInputs) > 0 {
				m.headerInputs = append(m.headerInputs[:m.focusedIndex], m.headerInputs[m.focusedIndex+1:]...)
				if m.focusedIndex >= len(m.headerInputs) {
					m.focusedIndex = max(0, len(m.headerInputs)-1)
				}
			} else if m.focusedSection == SectionQueryParams && len(m.queryInputs) > 0 {
				m.queryInputs = append(m.queryInputs[:m.focusedIndex], m.queryInputs[m.focusedIndex+1:]...)
				if m.focusedIndex >= len(m.queryInputs) {
					m.focusedIndex = max(0, len(m.queryInputs)-1)
				}
				// Sync back to URL
				m.syncURLFromParams()
			}
			m.updateFocus()
			return m, nil

		case "tab":
			m.nextField()
			return m, nil

		case "shift+tab":
			m.prevField()
			return m, nil

		case "up", "down", "left", "right":
			if m.focusedSection == SectionURL && m.focusedIndex == 0 {
				if msg.String() == "up" || msg.String() == "left" {
					m.methodIndex = (m.methodIndex - 1 + len(m.methods)) % len(m.methods)
				} else {
					m.methodIndex = (m.methodIndex + 1) % len(m.methods)
				}
				return m, nil
			}
		}

	case SuggestionSelectedMsg:
		// Insert the selected variable at the end (simplification)
		// Ideally insert at cursor, but we lack cursor access for now.
		// We know we triggered on "{{", so we append "var_name}}".
		// Actually we triggered on "{{", so we are right after it.
		// We just need to append "var_name}}".
		
		toAppend := msg.VarName + "}}"
		
		switch m.focusedSection {
		case SectionURL:
			m.pathInput.SetValue(m.pathInput.Value() + toAppend)
			m.parseURLParams()
		case SectionPathParams:
			if m.focusedIndex < len(m.pathParamsInputs) {
				m.pathParamsInputs[m.focusedIndex].value.SetValue(m.pathParamsInputs[m.focusedIndex].value.Value() + toAppend)
			}
		case SectionQueryParams:
			if m.focusedIndex < len(m.queryInputs) {
				if m.isKeyFocused {
					m.queryInputs[m.focusedIndex].key.SetValue(m.queryInputs[m.focusedIndex].key.Value() + toAppend)
				} else {
					m.queryInputs[m.focusedIndex].value.SetValue(m.queryInputs[m.focusedIndex].value.Value() + toAppend)
				}
				m.syncURLFromParams()
			}
		case SectionHeaders:
			if m.focusedIndex < len(m.headerInputs) {
				if m.isKeyFocused {
					m.headerInputs[m.focusedIndex].key.SetValue(m.headerInputs[m.focusedIndex].key.Value() + toAppend)
				} else {
					m.headerInputs[m.focusedIndex].value.SetValue(m.headerInputs[m.focusedIndex].value.Value() + toAppend)
				}
			}
		case SectionBody:
			m.bodyInput.SetValue(m.bodyInput.Value() + toAppend)
		}
		
		// Refocus input? They are already focused.
		// Maybe move cursor to end?
		m.moveCursorToEnd()
		return m, nil
	}

	// Update focused field
	var cmd tea.Cmd
	switch m.focusedSection {
	case SectionURL:
		if m.focusedIndex == 1 {
			oldVal := m.pathInput.Value()
			m.pathInput, cmd = m.pathInput.Update(msg)
			if m.pathInput.Value() != oldVal {
				m.parseURLParams()
				m.checkForTrigger(m.pathInput.Value())
			}
			cmds = append(cmds, cmd)
		}
	case SectionPathParams:
		if m.focusedIndex < len(m.pathParamsInputs) {
			if !m.isKeyFocused {
				inp := &m.pathParamsInputs[m.focusedIndex].value
				oldVal := inp.Value()
				*inp, cmd = inp.Update(msg)
				if inp.Value() != oldVal {
					m.checkForTrigger(inp.Value())
				}
				cmds = append(cmds, cmd)
			}
		}
	case SectionQueryParams:
		if m.focusedIndex < len(m.queryInputs) {
			oldKey := m.queryInputs[m.focusedIndex].key.Value()
			oldVal := m.queryInputs[m.focusedIndex].value.Value()
			
			if m.isKeyFocused {
				m.queryInputs[m.focusedIndex].key, cmd = m.queryInputs[m.focusedIndex].key.Update(msg)
			} else {
				m.queryInputs[m.focusedIndex].value, cmd = m.queryInputs[m.focusedIndex].value.Update(msg)
			}
			
			if m.queryInputs[m.focusedIndex].key.Value() != oldKey || m.queryInputs[m.focusedIndex].value.Value() != oldVal {
				m.syncURLFromParams()
				// Check trigger on focused simple inputs
				if m.isKeyFocused {
					m.checkForTrigger(m.queryInputs[m.focusedIndex].key.Value())
				} else {
					m.checkForTrigger(m.queryInputs[m.focusedIndex].value.Value())
				}
			}
			cmds = append(cmds, cmd)
		}
	case SectionHeaders:
		if m.focusedIndex < len(m.headerInputs) {
			oldKey := m.headerInputs[m.focusedIndex].key.Value()
			oldVal := m.headerInputs[m.focusedIndex].value.Value()

			if m.isKeyFocused {
				m.headerInputs[m.focusedIndex].key, cmd = m.headerInputs[m.focusedIndex].key.Update(msg)
			} else {
				m.headerInputs[m.focusedIndex].value, cmd = m.headerInputs[m.focusedIndex].value.Update(msg)
			}

			if m.headerInputs[m.focusedIndex].key.Value() != oldKey || m.headerInputs[m.focusedIndex].value.Value() != oldVal {
				if m.isKeyFocused {
					m.checkForTrigger(m.headerInputs[m.focusedIndex].key.Value())
				} else {
					m.checkForTrigger(m.headerInputs[m.focusedIndex].value.Value())
				}
			}
			cmds = append(cmds, cmd)
		}
	case SectionBody:
		oldVal := m.bodyInput.Value()
		m.bodyInput, cmd = m.bodyInput.Update(msg)
		if m.bodyInput.Value() != oldVal {
			m.checkForTrigger(m.bodyInput.Value())
		}
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m *RequestModel) checkForTrigger(text string) {
	// Simple trigger: if ends with {{
	if strings.HasSuffix(text, "{{") {
		m.suggestions.Show(m.vars, "", func(selected string) tea.Msg {
			// Callback to insert
			return SuggestionSelectedMsg{VarName: selected}
		}, nil)
	}
}

func (m *RequestModel) moveCursorToEnd() {
	// Best effort to move cursor to end
	switch m.focusedSection {
	case SectionURL:
		m.pathInput.CursorEnd()
	case SectionBody:
		m.bodyInput.CursorEnd()
	// Others...
	}
}


func (m *RequestModel) nextField() {
	switch m.focusedSection {
	case SectionURL:
		if m.focusedIndex == 0 { // at method
			m.focusedIndex = 1 // to url
		} else {
			if len(m.pathParamsInputs) > 0 {
				m.focusedSection = SectionPathParams
				m.focusedIndex = 0
				m.isKeyFocused = false // keys are derived
			} else {
				m.focusedSection = SectionQueryParams
				m.focusedIndex = 0
				m.isKeyFocused = true
			}
		}
	case SectionPathParams:
		m.focusedIndex++
		if m.focusedIndex >= len(m.pathParamsInputs) {
			m.focusedSection = SectionQueryParams
			m.focusedIndex = 0
			m.isKeyFocused = true
		}
	case SectionQueryParams:
		if m.isKeyFocused {
			m.isKeyFocused = false
		} else {
			m.focusedIndex++
			m.isKeyFocused = true
			if m.focusedIndex >= len(m.queryInputs) {
				m.focusedSection = SectionHeaders
				m.focusedIndex = 0
			}
		}
	case SectionHeaders:
		if m.isKeyFocused {
			m.isKeyFocused = false
		} else {
			m.focusedIndex++
			m.isKeyFocused = true
			if m.focusedIndex >= len(m.headerInputs) {
				m.focusedSection = SectionBody
				m.focusedIndex = 0
			}
		}
	case SectionBody:
		m.focusedSection = SectionURL
		m.focusedIndex = 0 // back to method
	}
	m.updateFocus()
}

func (m *RequestModel) prevField() {
	switch m.focusedSection {
	case SectionURL:
		if m.focusedIndex == 1 {
			m.focusedIndex = 0
		} else {
			m.focusedSection = SectionBody
			m.focusedIndex = 0
		}
	case SectionPathParams:
		if m.focusedIndex > 0 {
			m.focusedIndex--
		} else {
			m.focusedSection = SectionURL
			m.focusedIndex = 1
		}
	case SectionQueryParams:
		if !m.isKeyFocused {
			m.isKeyFocused = true
		} else {
			m.focusedIndex--
			m.isKeyFocused = false
			if m.focusedIndex < 0 {
				if len(m.pathParamsInputs) > 0 {
					m.focusedSection = SectionPathParams
					m.focusedIndex = len(m.pathParamsInputs) - 1
				} else {
					m.focusedSection = SectionURL
					m.focusedIndex = 1
				}
			}
		}
	case SectionHeaders:
		if !m.isKeyFocused {
			m.isKeyFocused = true
		} else {
			m.focusedIndex--
			m.isKeyFocused = false
			if m.focusedIndex < 0 {
				m.focusedSection = SectionQueryParams
				m.focusedIndex = max(0, len(m.queryInputs)-1)
			}
		}
	case SectionBody:
		m.focusedSection = SectionHeaders
		m.focusedIndex = max(0, len(m.headerInputs)-1)
		m.isKeyFocused = false
	}
	m.updateFocus()
}

func (m *RequestModel) updateFocus() {
	m.pathInput.Blur()
	m.bodyInput.Blur()
	
	for i := range m.pathParamsInputs {
		m.pathParamsInputs[i].value.Blur()
	}
	for i := range m.headerInputs {
		m.headerInputs[i].key.Blur()
		m.headerInputs[i].value.Blur()
	}
	for i := range m.queryInputs {
		m.queryInputs[i].key.Blur()
		m.queryInputs[i].value.Blur()
	}

	switch m.focusedSection {
	case SectionURL:
		if m.focusedIndex == 1 {
			m.pathInput.Focus()
		}
	case SectionPathParams:
		if m.focusedIndex < len(m.pathParamsInputs) {
			m.pathParamsInputs[m.focusedIndex].value.Focus()
		}
	case SectionQueryParams:
		if m.focusedIndex < len(m.queryInputs) {
			if m.isKeyFocused {
				m.queryInputs[m.focusedIndex].key.Focus()
			} else {
				m.queryInputs[m.focusedIndex].value.Focus()
			}
		}
	case SectionHeaders:
		if m.focusedIndex < len(m.headerInputs) {
			if m.isKeyFocused {
				m.headerInputs[m.focusedIndex].key.Focus()
			} else {
				m.headerInputs[m.focusedIndex].value.Focus()
			}
		}
	case SectionBody:
		m.bodyInput.Focus()
	}
}

// buildRequest constructs the request and possibly a targeted URL (with substitutions)
func (m *RequestModel) buildRequest() (storage.Request, string) {
	req := storage.Request{
		Name:    m.request.Name,
		Method:  m.methods[m.methodIndex],
		Headers: make(map[string]string),
		Body:    m.bodyInput.Value(),
	}

	// We use pathInput as source of truth for base URL structure
	// Queries are synced into pathInput so it contains them.
	// But we need to handle Path Params (:param) substitution for the ACTUAL execution URL.
	
	rawURL := m.pathInput.Value()
	
	// Apply path params substitutions
	targetedURL := rawURL
	for _, pp := range m.pathParamsInputs {
		key := ":" + pp.key.Value()
		val := pp.value.Value()
		if val != "" {
			targetedURL = strings.ReplaceAll(targetedURL, key, val)
		}
	}
	
	// Apply query params (already in URL mostly, but let's be safe)
	// Actually, if we use m.pathInput.Value(), it relies on syncURLFromParams working perfectly.
	// Let's assume it does.
	
	req.URL = rawURL // Saved request keeps the :params
	
	// Headers
	for _, hi := range m.headerInputs {
		if hi.key.Value() != "" {
			req.Headers[hi.key.Value()] = hi.value.Value()
		}
	}

	return req, targetedURL
}

func (m RequestModel) View() string {
	var sb strings.Builder

	// Base URL context
	if m.baseURL != "" {
		sb.WriteString(DimStyle.Render("Base URL: "+m.baseURL) + "\n\n")
	}

	// Method & URL
	method := m.methods[m.methodIndex]
	var methodView string
	if m.focusedSection == SectionURL && m.focusedIndex == 0 {
		methodView = MethodBadge(method) + " "
	} else {
		// Dims the badge slightly if not focused? Or keep it bright
		methodView = MethodBadge(method) + " " 
	}

	urlStr := m.pathInput.Value()
	// Highlight params in URL string for display would be cool
	// For now, just render input
	
	if m.preview {
		urlStr = m.highlightVars(urlStr)
	}

	urlView := m.pathInput.View()
	if m.focusedSection == SectionURL && m.focusedIndex == 1 {
		// Focused URL handled by textinput.View()
	} else if m.preview {
		urlView = urlStr // Already highlighted by highlightVars (which returns validated string with colors)
	} else {
		urlView = highlightPathParams(urlStr)
	}

	// top bar: METHOD URL
	sb.WriteString(lipgloss.JoinHorizontal(lipgloss.Center, methodView, urlView))
	sb.WriteString("\n\n")

	// Path Params
	if len(m.pathParamsInputs) > 0 {
		sb.WriteString(HeaderStyle.Render("PARAMS"))
		sb.WriteString("\n")
		sb.WriteString(m.renderKVSection(m.pathParamsInputs, SectionPathParams))
		sb.WriteString("\n")
	}

	// Headers
	sb.WriteString(HeaderStyle.Render("HEADERS"))
	sb.WriteString("\n")
	sb.WriteString(m.renderKVSection(m.headerInputs, SectionHeaders))
	sb.WriteString("\n")

	// Query Params
	sb.WriteString(HeaderStyle.Render("QUERY PARAMS"))
	sb.WriteString("\n")
	sb.WriteString(m.renderKVSection(m.queryInputs, SectionQueryParams))
	sb.WriteString("\n")

	// Body
	sb.WriteString(HeaderStyle.Render("BODY"))
	sb.WriteString("\n")
	
	bodyView := m.bodyInput.View()
	if m.preview {
		bodyView = m.highlightVars(m.bodyInput.Value())
	}
	sb.WriteString(bodyView)

	// Render suggestions overlay if visible
	if m.suggestions.visible {
		return lipgloss.JoinHorizontal(lipgloss.Top, sb.String(), m.suggestions.View())
	}

	return sb.String()
}

func (m RequestModel) renderKVSection(inputs []KVInput, section RequestSection) string {
	if len(inputs) == 0 {
		return DimStyle.Render("  (empty)") + "\n"
	}
	
	var sb strings.Builder
	for i, input := range inputs {
		keyView := input.key.View()
		valView := input.value.View()

		if m.preview {
			keyView = m.highlightVars(input.key.Value())
			valView = m.highlightVars(input.value.Value())
		}
		
		// Indicator for active row
		prefix := "  "
		if m.focusedSection == section && m.focusedIndex == i {
			prefix = "> "
		}

		row := lipgloss.JoinHorizontal(lipgloss.Center,
			SelectedStyle.Render(prefix),
			keyView,
			DimStyle.Render(" : "),
			valView,
		)
		sb.WriteString(row + "\n")
	}
	return sb.String()
}

func (m *RequestModel) SetVariables(vars map[string]string) {
	m.vars = vars
}
func (m RequestModel) highlightVars(text string) string {
	// Re-implement basic regex replacement with checks
	// varRegex is technically private in storage, need to redefine or expose?
	// It's defined in internal/storage/templating.go as var varRegex, but not exported.
	// We can redefine it here.
	
	// We iterate over MATCHES.
	// For every match {{key}}, we check if it exists in m.vars.
	// If yes -> Substitute and color GREEN/BOLD.
	// If no -> Leave as {{key}} and color RED/BOLD/WARNING.
	
	// Since we need to replace parts of string, we can use ReplaceAllStringFunc equivalent manually.
	
	// Simple approach: Split by {{ and }}
	
	// Let's use loop.
	result := text
	start := 0
	for {
		startIdx := strings.Index(result[start:], "{{")
		if startIdx == -1 {
			break
		}
		actualStart := start + startIdx
		
		endIdx := strings.Index(result[actualStart:], "}}")
		if endIdx == -1 {
			break
		}
		actualEnd := actualStart + endIdx + 2
		
		key := result[actualStart+2 : actualEnd-2]
		key = strings.TrimSpace(key)
		
		var replacement string
		if val, ok := m.vars[key]; ok {
			// Found
			replacement = lipgloss.NewStyle().Foreground(SecondaryColor).Bold(true).Render(val)
		} else {
			// Not Found
			replacement = lipgloss.NewStyle().Foreground(ErrorColor).Bold(true).Render("{{" + key + "}}")
		}
		
		// Replace
		result = result[:actualStart] + replacement + result[actualEnd:]
		
		// Update start to skip replaced part
		start = actualStart + len(replacement)
	}
	
	return result
}
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func highlightPathParams(urlStr string) string {
	parts := strings.Split(urlStr, "/")
	for i, part := range parts {
		if strings.HasPrefix(part, ":") && len(part) > 1 {
			parts[i] = ParamStyle.Render(part)
		}
	}
	return strings.Join(parts, "/")
}
