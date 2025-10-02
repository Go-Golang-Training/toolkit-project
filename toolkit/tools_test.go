package toolkit

import "testing"

func TestTools_RandomString(t *testing.T) {

	var tools Tools

	s := tools.RandomString(10)
	if len(s) != 10 {
		t.Errorf("Expected string length of 10, but got %d", len(s))
	}

	s2 := tools.RandomString(10)
	if s == s2 {
		t.Errorf("Expected different random strings, but got the same: %s", s)
	}
}
