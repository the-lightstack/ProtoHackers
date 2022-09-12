package main

import (
	"log"
	"reflect"
	"testing"
)

func TestParseJsonToFields(t *testing.T) {
	type test struct {
		input string
		want  map[string]Value
		fail  bool
	}

	tests := []test{
		{input: ``, fail: true},
		{input: `"`, fail: true},
		{input: `{"}`, fail: true},
		{input: `{::}`, fail: true},
		{input: `{""}`, fail: true},
		{input: `{"""}`, fail: true},
		{input: `{"Hey"`, fail: true},
		{input: `{John,`, fail: true},
		{input: `{"Hello"}`, fail: true},
		{input: `{Hello:200}`, fail: true},

		{input: `{"Hello":"How":"Are you"}`, fail: true},

		{input: `{"Hello":"John"}`, want: map[string]Value{"Hello": {Val: "John", IsString: true}}, fail: false},
		{input: `{"number":213}`, want: map[string]Value{"number": {Val: "213", IsString: false}}, fail: false},
		{input: `{"number":213.12}`, want: map[string]Value{"number": {Val: "213.12", IsString: false}}, fail: false},
		{input: `{"method":"isPrime","number":400}`, want: map[string]Value{"method": {Val: "isPrime", IsString: true}, "number": {Val: "400", IsString: false}}, fail: false},
		{input: `{"method":"isPrime","number":400.232}`, want: map[string]Value{"method": {Val: "isPrime", IsString: true}, "number": {Val: "400.232", IsString: false}}, fail: false},

		{input: `{"Hello":"John","Cool":"Right?"}`, want: map[string]Value{"Hello": {Val: "John", IsString: true}, "Cool": {Val: "Right?", IsString: true}}, fail: false},
		{input: `{"method":"isPrime","number":871218}`, want: map[string]Value{"method": {Val: "isPrime", IsString: true}, "number": {Val: "871218", IsString: false}}, fail: false},

		{input: `{"method":"isPrime","number":"400"}`, want: map[string]Value{"method": {Val: "isPrime", IsString: true}, "number": {Val: "400", IsString: true}}, fail: false},
		{input: `{"method":"isPrime","other":{"n":200},"number":100}`,
			want: map[string]Value{"method": {Val: "isPrime", IsString: true}, "other": {Val: `{"n":200}`, IsString: false}, "number": {Val: "100", IsString: false}}, fail: false},

		{input: `{"method":"isPrime","not\"method":"isntPrime","number":6534740}`,
			want: map[string]Value{"method": {Val: "isPrime", IsString: true}, `not"method`: {Val: `isntPrime`, IsString: true}, "number": {Val: "6534740", IsString: false}}, fail: false},
	}

	for _, tc := range tests {
		log.Println("Input", tc.input)
		result, err := ParseJsonToFields(tc.input)

		if tc.fail {
			// Check if err is also an error
			if err == nil {
				t.Errorf("Expected errror: %v", tc.input)
			}
		} else {
			if err != nil {
				t.Errorf("Error although should be fine: %v, error: %v", tc.input, err)
				continue
			}

			if !reflect.DeepEqual(result, tc.want) {
				t.Errorf("expected: %v, got: %v", tc.want, result)
			}
		}

	}
}

func TestFieldsToValidJsonRequest(t *testing.T) {
	type test struct {
		Input map[string]Value
		Want  JsonRequest
	}
	tests := []test{
		{Input: map[string]Value{"method": {Val: "isPrime", IsString: true}, "number": {Val: "100", IsString: false}},
			Want: JsonRequest{Malformed: false, Method: "isPrime", Number: 100.0}},
		{Input: map[string]Value{"method": {Val: "isPrime", IsString: true}, "number": {Val: "-200", IsString: false}},
			Want: JsonRequest{Malformed: false, Method: "isPrime", Number: -200.0}},
		{Input: map[string]Value{"method": {Val: "isPrime", IsString: true}, "number": {Val: "75642.232", IsString: false}},
			Want: JsonRequest{Malformed: false, Method: "isPrime", Number: 75642.232}},
		{Input: map[string]Value{"method": {Val: "isPrime", IsString: true}, "confuse": {Val: "you", IsString: false}, "number": {Val: "1337", IsString: false}},
			Want: JsonRequest{Malformed: false, Method: "isPrime", Number: 1337}},
		{Input: map[string]Value{"method": {Val: "weirdCrap", IsString: true}, "number": {Val: "120", IsString: false}},
			Want: JsonRequest{Malformed: true}},
		{Input: map[string]Value{"number": {Val: "120", IsString: false}},
			Want: JsonRequest{Malformed: true}},
		{Input: map[string]Value{},
			Want: JsonRequest{Malformed: true}},
		{Input: map[string]Value{"NotNumber": {Val: "hello", IsString: true}},
			Want: JsonRequest{Malformed: true}},
		{Input: map[string]Value{"method": {Val: "isPrime", IsString: true}, "number": {Val: "22dhb9", IsString: false}},
			Want: JsonRequest{Malformed: true}},
		{Input: map[string]Value{"method": {Val: "isPrime", IsString: true}, "number": {Val: "10", IsString: true}},
			Want: JsonRequest{Malformed: true}},
	}

	for _, tc := range tests {
		result := FieldsToValidJsonRequest(tc.Input)
		if result.Malformed != tc.Want.Malformed {
			t.Errorf("Expected Malformed: %v", tc.Input)
			continue
		}

		if !reflect.DeepEqual(result, tc.Want) {
			t.Errorf("expected: %v, got: %v", tc.Want, result)
		}
	}
}
