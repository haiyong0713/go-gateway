package card

import (
	"go-gateway/app/app-svr/app-card/interface/model/card/ai"
	"testing"

	"github.com/stretchr/testify/assert"
	"go-gateway/app/app-svr/app-card/interface/model"
)

func TestThreePointFrom(t *testing.T) {
	var (
		card1 = &SmallCoverV2{
			Base: &Base{
				CardType: model.SmallCoverV2,
				CardGoto: model.CardGotoAv,
				Goto:     model.GotoAv,
			},
		}
		card2 = &LargeCoverInline{
			Base: &Base{
				CardType: model.LargeCoverV9,
				CardGoto: model.CardGotoInlineAv,
				Goto:     model.GotoAv,
			},
		}
		card3 = &SmallCoverV9{
			Base: &Base{
				CardType: model.SmallCoverV9,
				CardGoto: model.CardGotoLive,
				Goto:     model.GotoLive,
			},
		}
		dislikeSubtitle = "(选择后将减少相似内容推荐)"
		dislikeToast    = "将减少相似内容推荐"
	)
	card1.Base.ThreePointFrom("iphone", 64900000, 1, nil, 2, 0)
	assert.Equal(t, dislikeSubtitle, card1.ThreePointV2[2].Subtitle)
	assert.Equal(t, dislikeToast, card1.ThreePointV2[2].Reasons[0].Toast)

	card1.ThreePointV2 = nil
	card1.Base.ThreePointFrom("iphone", 64900000, 1, nil, 2, 1)
	assert.Equal(t, "", card1.ThreePointV2[2].Subtitle)
	assert.Equal(t, "", card1.ThreePointV2[2].Reasons[0].Toast)

	card2.Base.ThreePointFrom("iphone", 64900000, 1, nil, 2, 0)
	assert.Equal(t, dislikeSubtitle, card2.ThreePointV2[2].Subtitle)
	assert.Equal(t, dislikeToast, card2.ThreePointV2[2].Reasons[0].Toast)

	card2.ThreePointV2 = nil
	card2.Base.ThreePointFrom("iphone", 64900000, 1, nil, 2, 1)
	assert.Equal(t, "", card2.ThreePointV2[2].Subtitle)
	assert.Equal(t, "", card2.ThreePointV2[2].Reasons[0].Toast)

	card3.Base.ThreePointFrom("android", 6490000, 1, nil, 2, 0)
	assert.Equal(t, dislikeSubtitle, card3.ThreePointV2[0].Subtitle)
	assert.Equal(t, dislikeToast, card3.ThreePointV2[0].Reasons[0].Toast)

	card3.ThreePointV2 = nil
	card3.Base.ThreePointFrom("android", 6490000, 1, nil, 2, 1)
	assert.Equal(t, "", card3.ThreePointV2[0].Subtitle)
	assert.Equal(t, "", card3.ThreePointV2[0].Reasons[0].Toast)
}

func TestCanEnableAvDislikeInfo(t *testing.T) {
	item1 := &ai.Item{AvDislikeInfo: 1}
	assert.Equal(t, CanEnableAvDislikeInfo(item1), true)

	item2 := &ai.Item{}
	assert.Equal(t, CanEnableAvDislikeInfo(item2), false)
}

func TestCheckMidMaxInt32(t *testing.T) {
	assert.Equal(t, true, CheckMidMaxInt32(2147483648))
	assert.Equal(t, false, CheckMidMaxInt32(2147483647))
	assert.Equal(t, false, CheckMidMaxInt32(1))
}
