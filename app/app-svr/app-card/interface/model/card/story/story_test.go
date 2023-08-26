package story

import (
	accountgrpc "git.bilibili.co/bapis/bapis-go/account/service"
	xroom "git.bilibili.co/bapis/bapis-go/live/xroom"
	"testing"

	"go-gateway/app/app-svr/app-card/interface/model/card/ai"
	arcgrpc "go-gateway/app/app-svr/archive/service/api"

	channelgrpc "git.bilibili.co/bapis/bapis-go/community/interface/channel"
	"github.com/stretchr/testify/assert"
)

func TestDislikeV3From(t *testing.T) {
	var (
		arc = &arcgrpc.ArcPlayer{
			Arc: &arcgrpc.Arc{
				Aid:    123,
				TypeID: 1,
				Author: arcgrpc.Author{
					Mid:  15555180,
					Name: "格格格",
				},
			},
		}
		ts = []*channelgrpc.Channel{
			{
				ID:   1,
				Name: "测试",
			},
		}
	)
	rcmd1 := &ai.SubItems{}
	rcmd1.SetDisableRcmd(1)
	dislike1 := dislikeV3From(arc, ts, rcmd1, nil)
	assert.Equal(t, "将在开启个性化推荐后生效", dislike1.DislikeItems[0].Toast)
	assert.Equal(t, "将在开启个性化推荐后生效", dislike1.DislikeItems[2].Toast)

	rcmd2 := &ai.SubItems{}
	dislike2 := dislikeV3From(arc, ts, rcmd2, nil)
	assert.Equal(t, "操作成功，将减少此类内容推荐", dislike2.DislikeItems[0].Toast)
	assert.Equal(t, "操作成功，将减少竖屏模式推荐", dislike2.DislikeItems[2].Toast)

	rcmd3 := &ai.SubItems{}
	dislike3 := dislikeV3From(arc, ts, rcmd3, nil)
	rcmd1.SetDisableRcmd(2)
	assert.Equal(t, "操作成功，将减少此类内容推荐", dislike3.DislikeItems[0].Toast)
	assert.Equal(t, "操作成功，将减少竖屏模式推荐", dislike3.DislikeItems[2].Toast)
}

func TestDislikeToast(t *testing.T) {
	rcmd1 := &ai.SubItems{}
	rcmd2 := &ai.SubItems{}
	rcmd1.SetDisableRcmd(1)

	assert.Equal(t, dislikeToast(rcmd1), "")
	assert.Equal(t, dislikeToast(rcmd2), "操作成功，将减少此类内容推荐")
}

func TestDislikeStoryToast(t *testing.T) {
	rcmd1 := &ai.SubItems{}
	rcmd2 := &ai.SubItems{}
	rcmd1.SetDisableRcmd(1)

	assert.Equal(t, dislikeToast(rcmd1), "")
	assert.Equal(t, dislikeToast(rcmd2), "操作成功，将减少竖屏模式推荐")
}

func TestDislikeV3FromLive(t *testing.T) {
	var (
		room = &xroom.EntryRoomInfoResp_EntryList{
			Uid:      15555180,
			AreaId:   39,
			AreaName: "分区",
		}
		cardm = map[int64]*accountgrpc.Card{
			15555180: {Name: "我是测试"},
		}
	)
	rcmd1 := &ai.SubItems{}
	rcmd1.SetDisableRcmd(1)
	dislike1 := dislikeV3FromLive(room, cardm, rcmd1)
	assert.Equal(t, "", dislike1.DislikeItems[0].Toast)
	assert.Equal(t, "不看UP主", dislike1.DislikeItems[0].Title)
	assert.Equal(t, "不看分区", dislike1.DislikeItems[1].Title)
	assert.Equal(t, "我不想看", dislike1.DislikeItems[2].Title)

	rcmd2 := &ai.SubItems{}
	dislike2 := dislikeV3FromLive(room, cardm, rcmd2)
	assert.Equal(t, "操作成功，将减少此类内容推荐", dislike2.DislikeItems[0].Toast)
	assert.Equal(t, "不看UP主", dislike2.DislikeItems[0].Title)
	assert.Equal(t, "不看分区", dislike2.DislikeItems[1].Title)
	assert.Equal(t, "我不想看", dislike2.DislikeItems[2].Title)

	rcmd3 := &ai.SubItems{}
	dislike3 := dislikeV3FromLive(room, cardm, rcmd3)
	rcmd1.SetDisableRcmd(2)
	assert.Equal(t, "操作成功，将减少此类内容推荐", dislike3.DislikeItems[0].Toast)
}
