package jsoncommon

import (
	"testing"

	jsoncard "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json"

	"github.com/stretchr/testify/assert"
)

func TestConstructArchiveDislikeReasons(t *testing.T) {
	arg := &jsoncard.Args{
		UpName: "测试up主",
		Rname:  "测试分区名",
		Tname:  "测试频道名",
	}
	dislikeReason1 := constructArchiveDislikeReasons(arg, 0)
	assert.Equal(t, len(dislikeReason1), 4)
	dislikeReason2 := constructArchiveDislikeReasons(arg, 1)
	assert.Equal(t, len(dislikeReason2), 6)
	assert.Equal(t, dislikeReason2[3].ID, int64(11))
	assert.Equal(t, dislikeReason2[3].Name, "此类内容过多")
	assert.Equal(t, dislikeReason2[4].ID, int64(12))
	assert.Equal(t, dislikeReason2[4].Name, "重复推荐")
}
