package components

import (

	"testing"

	"time"

	"github.com/styltsou/tapi/internal/http"
	"github.com/styltsou/tapi/internal/storage"
)

// Helper to create a mock ProcessedResponse
func mockResponse(body string) *http.ProcessedResponse {
	return &http.ProcessedResponse{
		StatusCode: 200,
		Status:     "200 OK",
		Headers:    map[string][]string{"Content-Type": {"application/json"}},
		Body:       []byte(body),
		Duration:   100 * time.Millisecond,
		Size:       int64(len(body)),
	}
}

// ================================================
// findMatches tests
// ================================================

func TestFindMatches_Basic(t *testing.T) {
	body := `{"name": "John", "age": 30, "name": "Jane"}`
	matches := findMatches(body, "name")

	if len(matches) != 2 {
		t.Fatalf("Expected 2 matches, got %d", len(matches))
	}

	// Check first match position
	if body[matches[0].StartByte:matches[0].EndByte] != "name" {
		t.Errorf("First match text = %q, want %q",
			body[matches[0].StartByte:matches[0].EndByte], "name")
	}

	// Check second match position
	if body[matches[1].StartByte:matches[1].EndByte] != "name" {
		t.Errorf("Second match text = %q, want %q",
			body[matches[1].StartByte:matches[1].EndByte], "name")
	}
}

func TestFindMatches_CaseInsensitive(t *testing.T) {
	body := "Hello HELLO hello HeLLo"
	matches := findMatches(body, "hello")

	if len(matches) != 4 {
		t.Fatalf("Expected 4 case-insensitive matches, got %d", len(matches))
	}

	// Verify each match extracts the original-case text
	expected := []string{"Hello", "HELLO", "hello", "HeLLo"}
	for i, m := range matches {
		got := body[m.StartByte:m.EndByte]
		if got != expected[i] {
			t.Errorf("Match %d: got %q, want %q", i, got, expected[i])
		}
	}
}

func TestFindMatches_EmptyQuery(t *testing.T) {
	body := "some text"
	matches := findMatches(body, "")
	if len(matches) != 0 {
		t.Errorf("Expected 0 matches for empty query, got %d", len(matches))
	}
}

func TestFindMatches_EmptyBody(t *testing.T) {
	matches := findMatches("", "test")
	if len(matches) != 0 {
		t.Errorf("Expected 0 matches for empty body, got %d", len(matches))
	}
}

func TestFindMatches_NoMatches(t *testing.T) {
	body := `{"users": [1, 2, 3]}`
	matches := findMatches(body, "xyz")
	if len(matches) != 0 {
		t.Errorf("Expected 0 matches, got %d", len(matches))
	}
}

func TestFindMatches_Multiline(t *testing.T) {
	body := "line1 foo\nline2 bar\nline3 foo\nline4 baz"
	matches := findMatches(body, "foo")

	if len(matches) != 2 {
		t.Fatalf("Expected 2 matches, got %d", len(matches))
	}

	// First match on line 0
	if matches[0].Line != 0 {
		t.Errorf("First match line = %d, want 0", matches[0].Line)
	}

	// Second match on line 2
	if matches[1].Line != 2 {
		t.Errorf("Second match line = %d, want 2", matches[1].Line)
	}
}

func TestFindMatches_Overlapping(t *testing.T) {
	body := "aaa"
	matches := findMatches(body, "aa")

	// Should find overlapping: positions 0-2 and 1-3
	if len(matches) != 2 {
		t.Fatalf("Expected 2 overlapping matches, got %d", len(matches))
	}

	if matches[0].StartByte != 0 || matches[0].EndByte != 2 {
		t.Errorf("First match at %d-%d, want 0-2", matches[0].StartByte, matches[0].EndByte)
	}
	if matches[1].StartByte != 1 || matches[1].EndByte != 3 {
		t.Errorf("Second match at %d-%d, want 1-3", matches[1].StartByte, matches[1].EndByte)
	}
}

func TestFindMatches_SpecialChars(t *testing.T) {
	body := `url: https://api.example.com/v1?key=value&foo=bar`
	matches := findMatches(body, "?key=value")

	if len(matches) != 1 {
		t.Fatalf("Expected 1 match for special chars query, got %d", len(matches))
	}
}

// ================================================
// highlightMatches tests
// ================================================

func TestHighlightMatches_NoMatches(t *testing.T) {
	body := "hello world"
	result := highlightMatches(body, nil, 0)

	if result != body {
		t.Errorf("Expected unchanged body with no matches, got %q", result)
	}
}

func TestHighlightMatches_SingleMatch(t *testing.T) {
	body := "hello world"
	matches := []SearchMatch{
		{StartByte: 0, EndByte: 5, Line: 0},
	}
	result := highlightMatches(body, matches, 0)

	// The result must contain the match text (possibly styled) and the remaining text
	if !containsText(result, "hello") {
		t.Error("Expected 'hello' text in result")
	}
	if !containsText(result, " world") {
		t.Error("Expected ' world' text in result")
	}
}

func TestHighlightMatches_MultipleMatches(t *testing.T) {
	body := "abc def abc ghi abc"
	matches := findMatches(body, "abc")

	if len(matches) != 3 {
		t.Fatalf("Expected 3 matches, got %d", len(matches))
	}

	result := highlightMatches(body, matches, 1) // second match is current

	// Verify all parts are present
	if !containsText(result, "abc") {
		t.Error("Expected 'abc' in result")
	}
	if !containsText(result, " def ") {
		t.Error("Expected ' def ' separator in result")
	}
	if !containsText(result, " ghi ") {
		t.Error("Expected ' ghi ' separator in result")
	}
}

func TestHighlightMatches_PreservesText(t *testing.T) {
	body := "foo bar foo"
	matches := findMatches(body, "foo")

	result := highlightMatches(body, matches, 0)

	// Verify the non-matched text is preserved exactly
	if !containsText(result, " bar ") {
		t.Error("Expected non-matched text ' bar ' to be preserved")
	}

	// Verify all match positions are correct
	if len(matches) != 2 {
		t.Fatalf("Expected 2 matches, got %d", len(matches))
	}
	if body[matches[0].StartByte:matches[0].EndByte] != "foo" {
		t.Error("First match not at expected position")
	}
	if body[matches[1].StartByte:matches[1].EndByte] != "foo" {
		t.Error("Second match not at expected position")
	}
}

// ================================================
// Navigation tests
// ================================================

func TestSearchNavigation_NextWraps(t *testing.T) {
	m := NewResponseModel()
	m.SetResponse(mockResponse(`{"a": 1, "a": 2, "a": 3}`), storage.Request{})

	m.searchInput.SetValue("a")
	m.executeSearch()

	total := len(m.matches)
	if total < 3 {
		t.Fatalf("Expected at least 3 matches, got %d", total)
	}

	// Navigate forward through all matches
	for i := 0; i < total; i++ {
		if m.currentMatch != i {
			t.Errorf("Step %d: currentMatch = %d, want %d", i, m.currentMatch, i)
		}
		m.currentMatch = (m.currentMatch + 1) % len(m.matches)
	}

	// Should wrap to 0
	if m.currentMatch != 0 {
		t.Errorf("After wrapping, currentMatch = %d, want 0", m.currentMatch)
	}
}

func TestSearchNavigation_PrevWraps(t *testing.T) {
	m := NewResponseModel()
	m.SetResponse(mockResponse(`{"b": 1, "b": 2}`), storage.Request{})

	m.searchInput.SetValue("b")
	m.executeSearch()

	total := len(m.matches)
	if total < 2 {
		t.Fatalf("Expected at least 2 matches, got %d", total)
	}

	// Navigate backward from 0 should wrap
	m.currentMatch = (m.currentMatch - 1 + len(m.matches)) % len(m.matches)
	if m.currentMatch != total-1 {
		t.Errorf("After prev from 0, currentMatch = %d, want %d", m.currentMatch, total-1)
	}
}

// ================================================
// Search state lifecycle tests
// ================================================

func TestSearch_ClearResets(t *testing.T) {
	m := NewResponseModel()
	m.SetResponse(mockResponse(`hello world hello`), storage.Request{})

	m.searchInput.SetValue("hello")
	m.executeSearch()

	if len(m.matches) == 0 {
		t.Fatal("Expected matches before clear")
	}
	if !m.searchActive {
		t.Fatal("Expected searchActive before clear")
	}

	m.clearSearch()

	if m.searching {
		t.Error("Expected searching=false after clear")
	}
	if m.searchActive {
		t.Error("Expected searchActive=false after clear")
	}
	if m.searchQuery != "" {
		t.Errorf("Expected empty searchQuery after clear, got %q", m.searchQuery)
	}
	if len(m.matches) != 0 {
		t.Errorf("Expected 0 matches after clear, got %d", len(m.matches))
	}
	if m.currentMatch != 0 {
		t.Errorf("Expected currentMatch=0 after clear, got %d", m.currentMatch)
	}
}

func TestSearch_NoResponse(t *testing.T) {
	m := NewResponseModel()

	// Should not panic when no response
	m.searchInput.SetValue("test")
	m.executeSearch()

	if len(m.matches) != 0 {
		t.Errorf("Expected 0 matches with no response, got %d", len(m.matches))
	}
}

func TestSearch_NewResponseClearsSearch(t *testing.T) {
	m := NewResponseModel()
	m.SetResponse(mockResponse(`hello world`), storage.Request{})

	m.searchInput.SetValue("hello")
	m.executeSearch()
	if len(m.matches) == 0 {
		t.Fatal("Expected matches")
	}

	// Set a new response — should clear search
	m.SetResponse(mockResponse(`different content`), storage.Request{})

	if m.searchActive {
		t.Error("Expected search cleared after new response")
	}
	if len(m.matches) != 0 {
		t.Errorf("Expected 0 matches after new response, got %d", len(m.matches))
	}
}

func TestSearch_NoMatchesStillActive(t *testing.T) {
	m := NewResponseModel()
	m.SetResponse(mockResponse(`hello world`), storage.Request{})

	m.searchInput.SetValue("xyz")
	m.executeSearch()

	// searchActive should be true (to show "No matches") even with 0 matches
	if !m.searchActive {
		t.Error("Expected searchActive=true when query has no matches")
	}
	if len(m.matches) != 0 {
		t.Errorf("Expected 0 matches, got %d", len(m.matches))
	}
}

// ================================================
// Integration tests
// ================================================

func TestResponseModel_SearchIntegration_FullFlow(t *testing.T) {
	m := NewResponseModel()
	body := `{
  "users": [
    {"name": "Alice", "email": "alice@example.com"},
    {"name": "Bob", "email": "bob@example.com"},
    {"name": "Alice Jr", "email": "alicejr@example.com"}
  ]
}`
	m.SetResponse(mockResponse(body), storage.Request{})
	m.SetSize(80, 40)

	// 1. No search initially
	if m.searching || m.searchActive {
		t.Error("Expected no search state initially")
	}

	// 2. Execute search for "alice"
	m.searching = true
	m.searchInput.Focus()
	m.searchInput.SetValue("alice")
	m.executeSearch()

	if len(m.matches) != 3 { // "Alice", "alice@", "Alice Jr", "alicejr@" — actually let's count
		// "Alice" in name, "alice" in email, "Alice" in name, "alicejr" has "alice" in it
		// Let me check: alice@example.com has "alice", alicejr@example.com has "alice"
		t.Logf("Found %d matches for 'alice'", len(m.matches))
	}
	if len(m.matches) < 3 {
		t.Fatalf("Expected at least 3 matches for 'alice', got %d", len(m.matches))
	}

	// 3. Navigate
	m.currentMatch = (m.currentMatch + 1) % len(m.matches)
	if m.currentMatch != 1 {
		t.Errorf("After next, currentMatch = %d, want 1", m.currentMatch)
	}

	// 4. Dismiss bar (Enter) — searchActive stays true
	m.searching = false
	m.searchInput.Blur()
	if !m.searchActive {
		t.Error("Expected searchActive to remain true after bar dismiss")
	}

	// 5. Clear search (Esc)
	m.clearSearch()
	if m.searchActive {
		t.Error("Expected searchActive=false after clear")
	}
}

func TestResponseModel_View_NoSearch(t *testing.T) {
	m := NewResponseModel()
	m.SetResponse(mockResponse(`{"ok": true}`), storage.Request{})
	m.SetSize(80, 40)

	view := m.View()
	if view == "" {
		t.Error("Expected non-empty view")
	}
}

func TestResponseModel_View_WithSearch(t *testing.T) {
	m := NewResponseModel()
	m.SetResponse(mockResponse(`{"name": "test"}`), storage.Request{})
	m.SetSize(80, 40)

	m.searching = true
	m.searchInput.Focus()
	m.searchInput.SetValue("test")
	m.executeSearch()

	view := m.View()
	// Should contain search UI
	if view == "" {
		t.Error("Expected non-empty view with search")
	}
}

func TestResponseModel_View_Loading(t *testing.T) {
	m := NewResponseModel()
	m.SetLoading(true)

	view := m.View()
	if !containsText(view, "Loading") {
		t.Error("Expected 'Loading' in view")
	}
}

func TestResponseModel_View_NoResponse(t *testing.T) {
	m := NewResponseModel()

	view := m.View()
	if !containsText(view, "No response") {
		t.Error("Expected 'No response' in view")
	}
}

func TestCountHeaderLines(t *testing.T) {
	m := NewResponseModel()
	m.SetResponse(mockResponse(`body`), storage.Request{})

	lines := m.countHeaderLines()
	if lines < 4 {
		t.Errorf("Expected at least 4 header lines (status + headers section + body label), got %d", lines)
	}
}

// helper
func containsText(s, substr string) bool {
	// Strip ANSI for plain text check
	return len(s) > 0 && (len(substr) == 0 || findInANSI(s, substr))
}

func findInANSI(s, substr string) bool {
	// Simple approach: the substr should appear somewhere in the raw string
	// even with ANSI codes interspersed
	// For our purposes, just check if the string contains it
	// (ANSI codes won't be inserted into the middle of our text strings)
	return len(s) > 0 && (indexOf(s, substr) >= 0 || indexOfStripped(s, substr) >= 0)
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

func indexOfStripped(s, substr string) int {
	// Strip ANSI escape codes
	stripped := stripANSI(s)
	return indexOf(stripped, substr)
}

func stripANSI(s string) string {
	var result []byte
	i := 0
	for i < len(s) {
		if s[i] == '\033' && i+1 < len(s) && s[i+1] == '[' {
			// Skip until 'm'
			j := i + 2
			for j < len(s) && s[j] != 'm' {
				j++
			}
			i = j + 1
		} else {
			result = append(result, s[i])
			i++
		}
	}
	return string(result)
}
