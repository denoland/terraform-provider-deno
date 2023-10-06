package provider

import "testing"

func TestCalculateGitSha1(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected string
	}{
		{
			name:  "blank",
			input: []byte(""),
			// echo -n "" | git hash-object --stdin
			expected: "e69de29bb2d1d6434b8b29ae775ad8c2e48c5391",
		},
		{
			name:  "short string",
			input: []byte("hey"),
			// echo -n "hey" | git hash-object -t blob --stdin
			expected: "2b31011cf9de6c82d52dc386cd7d1a9be83188c1",
		},
		{
			name:  "emoji (non-ascii)",
			input: []byte("😀"),
			// echo -n "😀" | git hash-object -t blob --stdin
			expected: "3995456fa473a609f2b0ff67e73cfc35d031eb5d",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calculateGitSha1(tt.input)
			if got != tt.expected {
				t.Errorf("calculateGitSha1() = %v, want %v", got, tt.expected)
			}
		})
	}
}
