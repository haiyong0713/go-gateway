package model

import (
	api "go-gateway/app/app-svr/app-player/interface/api/playurl"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestChooseFnVideo(t *testing.T) {
	var (
		baseurl7  = "7baseurl"
		baseurl12 = "12baseurl"
		baseurl13 = "13baseurl"
	)
	rawDashVideos := map[uint32][]*api.DashVideo{
		125: {
			{Codecid: 7, BaseUrl: baseurl7},
			{Codecid: 12, BaseUrl: baseurl12},
			{Codecid: 13, BaseUrl: baseurl13},
		},
		120: {
			{Codecid: 7, BaseUrl: baseurl7},
			{Codecid: 12, BaseUrl: baseurl12},
			{Codecid: 13, BaseUrl: baseurl13},
		},
		112: {
			{Codecid: 7, BaseUrl: baseurl7},
			{Codecid: 12, BaseUrl: baseurl12},
			{Codecid: 13, BaseUrl: baseurl13},
		},
	}
	fnVideo := chooseFnVideo(7, rawDashVideos)
	assert.Equal(t, 3, len(fnVideo))
	for _, v := range fnVideo {
		assert.Equal(t, baseurl7, v.BaseUrl)
	}
	fnVideo1 := chooseFnVideo(12, rawDashVideos)
	assert.Equal(t, 3, len(fnVideo1))
	for _, v := range fnVideo1 {
		assert.Equal(t, baseurl12, v.BaseUrl)
	}
	fnVideo2 := chooseFnVideo(13, rawDashVideos)
	assert.Equal(t, 3, len(fnVideo2))
	for _, v := range fnVideo2 {
		assert.Equal(t, baseurl13, v.BaseUrl)
	}
}
