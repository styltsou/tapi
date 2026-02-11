package ui

import (
	"testing"
)

func TestViewState_String(t *testing.T) {
	tests := []struct {
		state    ViewState
		expected string
	}{
		{ViewCollectionList, "Collection List"},
		{ViewRequestBuilder, "Request Builder"},
		{ViewResponse, "Response"},
		{ViewEnvironments, "Environments"},
		{ViewHelp, "Help"},
		{ViewState(99), "Unknown"},
	}

	for _, tt := range tests {
		if actual := tt.state.String(); actual != tt.expected {
			t.Errorf("ViewState(%d).String() = %q, want %q", tt.state, actual, tt.expected)
		}
	}
}
