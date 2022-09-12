package main

import "testing"

func TestIsPrime(t *testing.T) {
	type test struct {
		Number  float64
		IsPrime bool
	}

	testCases := []test{
		{Number: 1, IsPrime: false},
		{Number: 2, IsPrime: true},
		{Number: 3, IsPrime: true},
		{Number: 4, IsPrime: false},
		{Number: 5, IsPrime: true},
		{Number: 6, IsPrime: false},
		{Number: 7, IsPrime: true},
		{Number: 11, IsPrime: true},
		{Number: 13, IsPrime: true},
		{Number: 20, IsPrime: false},
		{Number: 100, IsPrime: false},
		{Number: 101, IsPrime: true},

		{Number: -1, IsPrime: false},
		{Number: -100.23, IsPrime: false},
		{Number: -100, IsPrime: false},
		{Number: 0, IsPrime: false},
		{Number: 12.12, IsPrime: false},
		{Number: 11.000, IsPrime: true},
	}

	for _, tc := range testCases {
		res := IsPrime(tc.Number)
		if res != tc.IsPrime {
			t.Errorf("Number: %v, expected: %v, got: %v", tc.Number, tc.IsPrime, res)
		}
	}
}
