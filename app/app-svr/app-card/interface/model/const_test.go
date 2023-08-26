package model

import (
	"testing"

	"go-gateway/app/app-svr/app-card/interface/model/card/ai"

	"github.com/stretchr/testify/assert"
)

func TestURLTrackIDHandler(t *testing.T) {
	rcmd1 := &ai.Item{TrackID: "test"}
	rcmd2 := &ai.SubItems{TrackID: "test"}
	uri := "http://www.bilibili.com"
	assert.Equal(t, URLTrackIDHandler(rcmd1)(uri), "http://www.bilibili.com?trackid=test")
	assert.Equal(t, URLTrackIDHandler(rcmd2)(uri), "http://www.bilibili.com?trackid=test")
	assert.Equal(t, URLTrackIDHandler(nil)(uri), "http://www.bilibili.com")
}
