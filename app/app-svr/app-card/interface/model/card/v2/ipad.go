package v2

import (
	"go-gateway/app/app-svr/app-card/interface/model"
	cardm "go-gateway/app/app-svr/app-card/interface/model/card"
	"go-gateway/app/app-svr/app-card/interface/model/card/ai"
	api "go-gateway/app/app-svr/app-card/interface/model/card/proto"

	accountgrpc "git.bilibili.co/bapis/bapis-go/account/service"
	relationgrpc "git.bilibili.co/bapis/bapis-go/account/service/relation"
)

func ipadHandle(cardGoto model.CardGt, cardType model.CardType, rcmd *ai.Item, tagm map[int64]*cardm.Tag, isAttenm, hasLike map[int64]int8, statm map[int64]*relationgrpc.StatReply,
	cardm map[int64]*accountgrpc.Card, authorRelations map[int64]*relationgrpc.InterrelationReply) (hander Handler) {
	base := &api.Base{CardType: cardType, CardGoto: cardGoto}
	card := &Card{
		Base:       base,
		CardCommon: &CardCommon{Rcmd: rcmd, Tagm: tagm, IsAttenm: isAttenm, HasLike: hasLike, Statm: statm, Cardm: cardm, Columnm: model.ColumnSvrSingle, AuthorRelations: authorRelations},
	}
	switch cardType {
	case model.SmallCoverV5:
		hander = &SmallCoverV5{Card: card, Item: &api.SmallCoverV5{Base: base}}
	// case model.PopularTopEntrance:
	// 	hander = &PopularTopEntrance{Card: card, Item: &api.PopularTopEntrance{Base: base}}
	default:
		//nolint:exhaustive
		switch cardGoto {
		case model.CardGotoAv, model.CardGotoBangumi, model.CardGotoLive, model.CardGotoPGC:
			base.CardType = model.LargeCoverV1
			hander = &LargeCoverV1{Card: card, Item: &api.LargeCoverV1{Base: base}}
		}
	}
	return
}
