package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFetchCidOnlineByPlat(t *testing.T) {
	in := map[string]int64{
		"ugc://806677215/442758847/android":        3,
		"ugc://806677215/442758847/android_hd":     1,
		"ugc://806677215/442758847/ipad":           1,
		"ugc://806677215/442758847/iphone":         19,
		"ugc://806677215/443220416/android":        2657,
		"ugc://806677215/443220416/android_b":      4,
		"ugc://806677215/443220416/android_hd":     16,
		"ugc://806677215/443220416/android_i":      13,
		"ugc://806677215/443220416/android_tv_yst": 39,
		"video://806677215/442758847":              636,
		"video://806677215/443220416":              3968,
	}
	webCountMap := map[int64]int64{
		442758847: 636,
		443220416: 3968,
	}
	appCountMap := map[int64]int64{
		442758847: 24,
		443220416: 2729,
	}
	web, app := fetchCidOnlineByPlat(in)
	assert.Equal(t, len(web), 2)
	assert.Equal(t, len(app), 2)
	for cid, count := range web {
		assert.Equal(t, webCountMap[cid], count)
	}
	for cid, count := range app {
		assert.Equal(t, appCountMap[cid], count)
	}
}
