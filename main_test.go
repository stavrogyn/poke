package main

import (
	"reflect"
	"testing"
)

func TestCleanInput(t *testing.T) {
	tests := []struct {
		input    string
		expected []string
	}{
		{input: "  hello  world  ", expected: []string{"hello", "world"}},
	}
	for _, test := range tests {
		actual := CleanInput(test.input)
		if !reflect.DeepEqual(actual, test.expected) {
			t.Errorf("cleanInput(%q) = %v; want %v", test.input, actual, test.expected)
		}
	}
}
