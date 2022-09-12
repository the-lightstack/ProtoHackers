package main

import (
	"errors"
	"log"
	"strconv"
)

type JsonRequest struct {
	Malformed bool
	Method    string
	Number    float64
}

type Value struct {
	Val      string
	IsString bool
}

func ParseJsonToFields(s string) (map[string]Value, error) {
	fields := make(map[string]Value)

	// String must be longer than 0
	if len(s) < 2 {
		return nil, errors.New("string must be longer than 0 char")
	}

	// First char must be a curly brace
	if s[0] != '{' || s[len(s)-1] != '}' {
		return nil, errors.New("json must start with '{' and end with '}'")
	}

	currentName := ""
	currentValue := ""
	isInString := false
	isInBrackets := false
	targetIsName := true
	escapeNextChar := false
	lastFieldWasString := false

	// We will split it on every comma that is not in a string
	for _, c := range s[1:] {
		if c == '"' && !isInBrackets && !escapeNextChar {
			isInString = !isInString
			lastFieldWasString = true
			continue
		}

		// One key-value pair is over
		if c == ',' && !isInString && !isInBrackets {
			// fields[currentName] = currentValue
			fields[currentName] = Value{Val: currentValue, IsString: lastFieldWasString}

			currentName = ""
			currentValue = ""
			targetIsName = !targetIsName

			continue
		}
		if c == '{' {
			isInBrackets = true
		}

		if c == '\\' {
			log.Println("Set escapeNextChar true")
			escapeNextChar = true
			continue
		}

		if c == '}' && !isInString && !isInBrackets {
			fields[currentName] = Value{Val: currentValue, IsString: lastFieldWasString}
			break
		}

		if c == '}' {
			isInBrackets = false
		}

		// One field (name) is done
		if c == ':' && !isInString && !isInBrackets {
			lastFieldWasString = false
			if !targetIsName {
				// Would be something like this: {"a":"b":"c"}
				return nil, errors.New("one key can only have one value")
			}
			targetIsName = !targetIsName
			continue
		}

		// Key must always be a string
		if targetIsName && !isInString {
			return nil, errors.New("data not in string")
		}
		// Copy bytes into either Name or Value
		if targetIsName {
			currentName += string(rune(c))
		} else {
			currentValue += string(rune(c))
		}
		escapeNextChar = false
	}

	if isInString {
		return nil, errors.New("unclosed string")
	}

	if targetIsName {
		return nil, errors.New("focus should be on second element (don't end with comma)")
	}

	return fields, nil
}

func FieldsToValidJsonRequest(fields map[string]Value) JsonRequest {
	method, methodFieldExists := fields["method"]
	number, numberFieldExists := fields["number"]

	if !methodFieldExists || !numberFieldExists {
		return JsonRequest{Malformed: true}
	}

	if method.Val != "isPrime" {
		return JsonRequest{Malformed: true}
	}

	if number.IsString {
		return JsonRequest{Malformed: true}
	}
	numericNumber, err := strconv.ParseFloat(number.Val, 64)
	if err != nil {
		return JsonRequest{Malformed: true}
	}

	return JsonRequest{Number: numericNumber, Method: method.Val, Malformed: false}
}
