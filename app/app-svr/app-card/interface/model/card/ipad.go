package card

import (
	"go-common/library/log"

	"go-gateway/app/app-svr/app-card/interface/model"
	"go-gateway/app/app-svr/app-card/interface/model/card/ai"
	"go-gateway/app/app-svr/app-card/interface/model/stat"

	accountgrpc "git.bilibili.co/bapis/bapis-go/account/service"
	relationgrpc "git.bilibili.co/bapis/bapis-go/account/service/relation"
	taggrpc "git.bilibili.co/bapis/bapis-go/community/interface/tag"
)

func ipadHandle(cardGoto model.CardGt, cardType model.CardType, rcmd *ai.Item, tagm map[int64]*taggrpc.Tag, isAttenm, hasLike map[int64]int8, statm map[int64]*relationgrpc.StatReply, cardm map[int64]*accountgrpc.Card, authorRelations map[int64]*relationgrpc.InterrelationReply) (hander Handler) {
	base := &Base{CardGoto: cardGoto, Rcmd: rcmd, Tagm: tagm, IsAttenm: isAttenm, HasLike: hasLike, Statm: statm, Cardm: cardm, Columnm: model.ColumnSvrSingle, AuthorRelations: authorRelations}
	if rcmd != nil {
		base.fillRcmdMeta(rcmd)
	}
	switch cardType {
	case model.BannerV6:
		base.CardType = model.BannerV6
		hander = &Banner{Base: base}
	case model.SmallCoverV5: // 针对热门ipad特殊处理
		base.CardType = model.SmallCoverV5
		hander = &SmallCoverV5{Base: base}
	case model.BannerIPadV8: // 应客户端要求，16:9 ipad仅更改card_type,不变更banner结构
		base.CardType = model.BannerIPadV8
		hander = &Banner{Base: base}
	case model.SmallCoverV9:
		base.CardType = model.SmallCoverV9
		base.CardLen = 1
		hander = &SmallCoverV9{Base: base}
	default:
		switch cardGoto {
		case model.CardGotoAv, model.CardGotoBangumi, model.CardGotoLive, model.CardGotoPGC, model.CardGotoSpecialS:
			base.CardType = model.LargeCoverV1
			base.CardLen = 1
			hander = &LargeCoverV1{Base: base}
		case model.CardGotoBangumiRcmd:
			base.CardType = model.SmallCoverV1
			hander = &SmallCoverV1{Base: base}
		case model.CardGotoBanner:
			base.CardType = model.BannerV3
			hander = &Banner{Base: base}
		case model.CardGotoAdAv:
			base.CardType = model.CmV1
			base.CardLen = 1
			hander = &LargeCoverV1{Base: base}
		case model.CardGotoAdWebS:
			base.CardType = model.CmV1
			base.CardLen = 1
			hander = &SmallCoverV1{Base: base}
		case model.CardGotoSearchUpper:
			base.CardType = model.ThreeItemAll
			hander = &ThreeItemAll{Base: base}
		case model.CardGotoTunnel:
			base.CardType = model.SmallCoverV1
			hander = &SmallCoverV1{Base: base}
		case model.CardGotoArticleS:
			base.CardType = model.LargeCoverV1
			hander = &LargeCoverV1{Base: base}
		default:
			log.Error("Fail to build handler, rowType=%s cardType=%s cardGoto=%s ai={%+v}",
				stat.RowTypeIPad, string(cardType), string(cardGoto), rcmd)
		}
	}
	stat.MetricAppCardTotal.Inc(stat.RowTypeIPad, string(base.CardType), string(cardGoto))
	return
}
