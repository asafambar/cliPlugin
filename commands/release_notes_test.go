package commands

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSimpleHello(t *testing.T) {
	conf := &releaseNotesConfiguration{
		addressee: "World",
		repeat:    1,
	}
	assert.Equal(t, doGetReleaseNotes(conf), "Hello World!")
}

func TestComplexHello(t *testing.T) {
	conf := &releaseNotesConfiguration{
		addressee: "World",
		repeat:    3,
		shout:     true,
		prefix:    "test: ",
	}
	assert.Equal(t, doGetReleaseNotes(conf), "TEST: HELLO WORLD!\nTEST: HELLO WORLD!\nTEST: HELLO WORLD!")
}
