package processor

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseDu(t *testing.T) {
	out := "4096\t/tmp\n123456789\t/home/rootless/.local/share/containers/storage\ngarbage\n"
	assert.Equal(t, map[string]int64{
		"/tmp": 4096,
		"/home/rootless/.local/share/containers/storage": 123456789,
	}, parseDu(out))
	assert.Empty(t, parseDu(""))
}

func TestCountLines(t *testing.T) {
	assert.Equal(t, 0, countLines(""))
	assert.Equal(t, 1, countLines("abc"))
	assert.Equal(t, 3, countLines("a\nb\nc"))
}
