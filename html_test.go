package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuilderIconForBuildHappy(t *testing.T) {
	b := &Build{Success: true}
	icon := builderIconForBuild(b)
	assert.Equal(t, happyBuilder, icon, "should be the happy builder")
}

func TestBuilderIconForBuildWorking(t *testing.T) {
	b := &Build{Success: false}
	icon := builderIconForBuild(b)
	assert.Equal(t, workingBuilder, icon, "should be the working builder")
}

func TestBuilderIconForBuildSad(t *testing.T) {
	b := &Build{Success: false, CompletedAt: "some value"}
	icon := builderIconForBuild(b)
	assert.Equal(t, sadBuilder, icon, "should be the sad builder")
}

func TestGithubRevisionLink(t *testing.T) {
	b := &Build{Id: "mattr-/mattr-.github.com-1234567"}
	link := githubRevisionLink(b)
	assert.Equal(t, "<a href='https://github.com/mattr-/mattr-.github.com/commit/1234567'>1234567</a>", link)
}
