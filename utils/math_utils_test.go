package utils_test

import (
	"backend-go/utils"
	"testing"
)

func TestMin(t *testing.T) {
	type testCase struct {
		a, b     int
		expected int
		name     string
	}

	testCases := []testCase{
		{a: 2, b: 5, expected: 2, name: "FirstLessThanSecond"},
		{a: 10, b: 3, expected: 3, name: "SecondLessThanFirst"},
		{a: 7, b: 7, expected: 7, name: "BothEqual"},
		{a: -4, b: -2, expected: -4, name: "NegativeNumbers"},
		{a: -10, b: 5, expected: -10, name: "MixedSignNumbers"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := utils.Min(tc.a, tc.b)
			if result != tc.expected {
				t.Errorf("Min(%d, %d): expected %d, got %d", tc.a, tc.b, tc.expected, result)
			}
		})
	}
}
