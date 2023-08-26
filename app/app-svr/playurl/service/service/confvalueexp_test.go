package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTimePeriods(t *testing.T) {
	durations := batchGetDurationBetweenExpAndNow(phoneExpTime, padExpTime)
	periods, _ := buildPeriods(durations)
	assert.Equal(t, "0-2245,0-13", periods)
}
