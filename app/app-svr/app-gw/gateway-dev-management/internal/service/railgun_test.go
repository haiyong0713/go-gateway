package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_GetStartAndEndTime(t *testing.T) {
	start, end := GetStartAndEndTime()
	assert.Equal(t, "2021-11-01", start)
	assert.Equal(t, "2021-11-07", end)
}
