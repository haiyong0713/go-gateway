package ai

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTrackId(t *testing.T) {
	trackID := "test"
	a1 := &Item{TrackID: trackID}
	a2 := &SubItems{TrackID: trackID}

	assert.Equal(t, a1.TrackId(), trackID)
	assert.Equal(t, a2.TrackId(), trackID)
	assert.Equal(t, a1.TrackID, trackID)
	assert.Equal(t, a2.TrackID, trackID)
	a1 = nil
	assert.Equal(t, a1.TrackId(), "")
}
