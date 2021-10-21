package errx

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetErrorStack(t *testing.T) {
	e := GetErrorStack()
	s := e.Stack()
	assert.Regexp(t, `(?s)--- goroutine.+GetErrorStack\(`, s)
}
