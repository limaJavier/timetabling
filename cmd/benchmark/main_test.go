package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseDuration(t *testing.T) {
	assert.Equal(t, int64(60*1000+1000+120), parseDuration("00:01:01.12"))
	assert.Equal(t, int64(60*60*1000+60*1000+1000+120), parseDuration("01:01:01.12"))
	assert.Equal(t, int64(60*1000+1000+120), parseDuration("1:01.12"))
	assert.Equal(t, int64(120), parseDuration("0:00.12"))
	assert.Equal(t, int64(120), parseDuration("00:00:00.12"))
}
