package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestVersionComparisonWithMajorVersion(t *testing.T) {
	assert.True(t, isLess("1.2", "2.1"))
	assert.False(t, isLess("2.1", "1.2"))
}

func TestVersionComparisonWithMinorVersion(t *testing.T) {
	assert.True(t, isLess("1.2", "1.3"))
	assert.False(t, isLess("1.3", "1.2"))
}

func TestVersionComparisonWithTwoDigitVersion(t *testing.T) {
	assert.True(t, isLess("1.2", "1.10"))
	assert.False(t, isLess("1.10", "1.2"))
}

func TestVersionComparisonWithDifferentVersionParts(t *testing.T) {
	assert.True(t, isLess("1", "1.1"))
	assert.False(t, isLess("1.1", "1"))
}

func TestVersionComparisonWithLettersDoesNotFail(t *testing.T) {
	isLess("1.a", "1.1")
}
