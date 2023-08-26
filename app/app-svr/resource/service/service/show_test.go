package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMatchChannelWithHiddenMetaInfo(t *testing.T) {
	channels := []string{
		"xiaomi",
		"huawei",
		"oppo",
		"vivo",
		"onePlus",
	}
	channelMap := map[string]string{
		"xiaomi": "xiaomi",
		"iphone": "iphone",
	}
	channelFuzzy := []string{"%hua", "op%", "vi@"}
	result := []bool{
		true,
		true,
		true,
		false,
		false,
	}
	for k, v := range channels {
		hit := matchChannelWithHiddenMetaInfo(v, channelMap, channelFuzzy)
		assert.Equal(t, result[k], hit)
	}
}
