package provider

import "testing"

func TestEncodePath(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "blank",
			input:    "",
			expected: "",
		},
		{
			name:     "basic",
			input:    "foo.ts",
			expected: "foo.ts",
		},
		{
			name:     "two hierarchy without special characters",
			input:    "foo/bar.ts",
			expected: "foo/bar.ts",
		},
		{
			name:     "two hierarchy with a special character `:`",
			input:    "foo/x:y.ts",
			expected: "foo/x%3Ay.ts",
		},
		{
			name:     "two hierarchy with a special character `?`",
			input:    "foo/x?y.ts",
			expected: "foo/x%3Fy.ts",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := encodePath(tt.input)
			if got != tt.expected {
				t.Errorf("encodePath() = %v, want %v", got, tt.expected)
			}
		})
	}
}
