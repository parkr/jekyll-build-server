package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExecer_scrubLogMsg(t *testing.T) {
	e := &Execer{}
	input := "https://parkr:sekrit!!@github.com/parkr/jekyll-build-server.git"
	assert.Equal(t, e.scrubLogMsg(input, "sekrit!!"), "https://parkr:*******@github.com/parkr/jekyll-build-server.git")
}
