package main

import "testing"

func TestIsPortableStatus(t *testing.T) {
	tests := []struct {
		status string
		want   bool
	}{
		{"ALLOCATED PORTABLE", true},
		{"ASSIGNED PORTABLE", true},
		{"ALLOCATED NON-PORTABLE", false},
		{"ASSIGNED NON-PORTABLE", false},
		{"", false},
	}
	for _, test := range tests {
		if got := isPortableStatus(test.status); got != test.want {
			t.Fatalf("isPortableStatus(%q) = %v, want %v", test.status, got, test.want)
		}
	}
}
