package client

import (
	"testing"
)

func TestMarshalJSONNoEscape(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		contains string
	}{
		{
			name:     "preserves >= operator",
			input:    map[string]string{"query": "value >= 100"},
			contains: `"query":"value >= 100"`,
		},
		{
			name:     "preserves < operator",
			input:    map[string]string{"query": "value < 50"},
			contains: `"query":"value < 50"`,
		},
		{
			name:     "preserves & character",
			input:    map[string]string{"query": "a & b"},
			contains: `"query":"a & b"`,
		},
		{
			name:     "preserves all HTML-sensitive chars together",
			input:    map[string]string{"filter": "count >= 1 && value < 10"},
			contains: `"filter":"count >= 1 && value < 10"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := marshalJSONNoEscape(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got := string(result); !containsSubstring(got, tt.contains) {
				t.Errorf("expected output to contain %q, got %q", tt.contains, got)
			}
		})
	}
}

func containsSubstring(s, substr string) bool {
	return len(s) >= len(substr) && searchSubstring(s, substr)
}

func searchSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
