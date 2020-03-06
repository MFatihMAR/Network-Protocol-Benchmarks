package condition

import (
	"fmt"
	"testing"
)

func assert(t *testing.T, cond bool, msg string) {
	if cond {
		t.Fatal(msg)
	}
}

func assertf(t *testing.T, cond bool, msg string, args ...interface{}) {
	assert(t, cond, fmt.Sprintf(msg, args...))
}
