package weir

import "testing"

func TestIsPowerOfTwo(t *testing.T) {
	testCases := []struct {
		name    string
		payload int
		wants   bool
	}{
		{
			name:    "zero is not a power of two",
			payload: 0,
			wants:   false,
		},
		{
			name:    "negative number is not a power of two",
			payload: -4,
			wants:   false,
		},
		{
			name:    "1 is a power of 2",
			payload: 1,
			wants:   true,
		},
		{
			name:    "16 is a power of 2",
			payload: 16,
			wants:   true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			observed := isPowerOfTwo(tc.payload)

			if observed != tc.wants {
				t.Errorf("isPowerOfTwo(%d) = %v; want %v", tc.payload, observed, tc.wants)
			}
		})
	}
}
