package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"
	"github.com/muesli/ansi"
)

// placeOverlay composites the overlay string onto the base string at [x, y] coordinates.
func placeOverlay(base, overlay string, x, y int) string {
	baseLines := strings.Split(base, "\n")
	overlayLines := strings.Split(overlay, "\n")

	for i, oLine := range overlayLines {
		targetY := y + i
		if targetY < 0 || targetY >= len(baseLines) {
			continue
		}

		bLine := baseLines[targetY]
		baseLines[targetY] = compositeLine(bLine, oLine, x)
	}

	return strings.Join(baseLines, "\n")
}

// compositeLine replaces a portion of the base line with the overlay line at visual offset x.
func compositeLine(base, overlay string, x int) string {
	if x < 0 {
		x = 0
	}

	overlayWidth := lipgloss.Width(overlay)
	
	// We need to slice the base string at visual offsets x and x + overlayWidth.
	left := truncateVisual(base, x)
	right := tailVisual(base, x+overlayWidth)
	
	return left + overlay + right
}

// truncateVisual returns the prefix of s with visual width w.
func truncateVisual(s string, w int) string {
	if w <= 0 {
		return ""
	}
	
	var (
		curWidth int
		result   strings.Builder
		inAnsi   bool
	)

	for _, r := range s {
		if r == ansi.Marker {
			inAnsi = true
			result.WriteRune(r)
			continue
		}
		if inAnsi {
			result.WriteRune(r)
			if ansi.IsTerminator(r) {
				inAnsi = false
			}
			continue
		}

		rw := runewidth.RuneWidth(r)
		if curWidth+rw > w {
			break
		}
		curWidth += rw
		result.WriteRune(r)
	}
	
	return result.String()
}

// tailVisual returns the suffix of s starting at visual width w.
func tailVisual(s string, w int) string {
	var (
		curWidth int
		inAnsi   bool
		found    bool
		result   strings.Builder
	)

	for _, r := range s {
		if r == ansi.Marker {
			inAnsi = true
			if found {
				result.WriteRune(r)
			}
			continue
		}
		if inAnsi {
			if found {
				result.WriteRune(r)
			}
			if ansi.IsTerminator(r) {
				inAnsi = false
			}
			continue
		}

		if found {
			result.WriteRune(r)
			continue
		}

		rw := runewidth.RuneWidth(r)
		if curWidth >= w {
			found = true
			result.WriteRune(r)
		}
		curWidth += rw
	}
	
	return result.String()
}
