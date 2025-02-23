package shared

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFilter_Slice(t *testing.T) {
	tests := []struct {
		name      string
		input     []int
		predicate func(int) bool
		selector  func(n int) int
		expected  []int
	}{
		{
			name:  "filter even numbers",
			input: []int{1, 2, 3, 4, 5, 6},
			predicate: func(n int) bool {
				return n%2 == 0
			},
			selector: func(n int) int {
				return n
			},
			expected: []int{2, 4, 6},
		},
		{
			name:  "filter odd numbers",
			input: []int{1, 2, 3, 4, 5, 6},
			predicate: func(n int) bool {
				return n%2 != 0
			},
			selector: func(n int) int {
				return n
			},
			expected: []int{1, 3, 5},
		},
		{
			name:  "filter greater than 3",
			input: []int{1, 2, 3, 4, 5, 6},
			predicate: func(n int) bool {
				return n > 3
			},
			selector: func(n int) int {
				return n
			},
			expected: []int{4, 5, 6},
		},
		{
			name:  "filter less than 0",
			input: []int{-3, -2, -1, 0, 1, 2, 3},
			predicate: func(n int) bool {
				return n < 0
			},
			selector: func(n int) int {
				return n
			},
			expected: []int{-3, -2, -1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Filter(tt.input, tt.predicate, tt.selector)
			assert.Equal(t, tt.expected, result)
		})
	}
}
