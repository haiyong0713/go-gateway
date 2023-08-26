package channel_v2

import (
	"context"
	"time"

	"go-common/component/metadata/auth"
	"go-common/component/metadata/device"
	"go-common/library/log"
	"go-gateway/app/app-svr/app-card/interface/model/card/ai"
	"go-gateway/app/app-svr/app-card/interface/model/card/threePointMeta"
	"go-gateway/app/app-svr/app-channel/interface/model/pedia"
	cardschema "go-gateway/app/app-svr/app-feed/interface-ng/card-schema"
	jsoncard "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json"
	feedcard "go-gateway/app/app-svr/app-feed/interface-ng/feed-card"

	baikegrpc "git.bilibili.co/bapis/bapis-go/community/interface/baike"
	channelmodel "git.bilibili.co/bapis/bapis-go/community/model/channel"
)

func (s *Service) BaikeNav(ctx context.Context, params *pedia.NavReq) (*pedia.NavResponse, error) {
	// 获取设备信息
	dev, _ := device.FromContext(ctx)
	reply, err := s.pediaDao.BaikeDetails(ctx, &baikegrpc.BaikeDetailReq{
		Bid: params.Bid,
		Meta: &channelmodel.MetaDataCtrl{
			Platform: dev.RawPlatform,
			From:     "/x/v2/channel/baike/nav",
		},
	})
	if err != nil {
		log.Error("BaikeNav params=%+v, err=%+v", params, err)
		return nil, err
	}
	return constructNavResponse(reply)
}

func constructNavResponse(detailRsp *baikegrpc.BaikeDetailRsp) (*pedia.NavResponse, error) {
	return &pedia.NavResponse{
		Version:    detailRsp.Version,
		Navigation: &pedia.Navigation{List: convertNodesToBaikeNavigation(detailRsp.Navigation)},
		BaikeTree: &pedia.BaikeTree{
			ContentTitle: "目录",
			Part:         convertNodesToBaikePart(detailRsp.Tree),
		},
		BaikeInfo: &pedia.BaikeInfo{
			BaikeName: detailRsp.Name,
			Desc:      detailRsp.Desc,
		},
	}, nil
}

func convertNodesToBaikePart(nodes []*baikegrpc.Node) []*pedia.NavPart {
	if len(nodes) == 0 {
		return nil
	}
	var navPart []*pedia.NavPart
	for k, v := range nodes {
		navPart = append(navPart, &pedia.NavPart{Position: k + 1, Nid: v.Nid, Title: v.Name, Part: convertNodesToBaikePart(v.Child)})
	}
	return navPart
}

func convertNodesToBaikeNavigation(nodes []*baikegrpc.Node) []*pedia.NavList {
	if len(nodes) == 0 {
		return nil
	}
	var navList []*pedia.NavList
	for _, v := range nodes {
		navList = append(navList, &pedia.NavList{Nid: v.Nid, Title: v.Name})
	}
	return navList
}

func (s *Service) BaikeFeed(ctx context.Context, params *pedia.FeedReq) (*pedia.FeedResponse, error) {
	// 获取设备信息
	dev, _ := device.FromContext(ctx)
	reply, err := s.pediaDao.BaikeFeed(ctx, &baikegrpc.BaikeFeedReq{
		Bid:      params.Bid,
		Nid:      params.Nid,
		Vertical: params.Vertical,
		Offset:   params.Offset,
		Version:  params.Version,
		Ps:       params.Ps,
		Meta: &channelmodel.MetaDataCtrl{
			Platform: dev.RawPlatform,
			From:     "/x/v2/channel/baike/feed",
		},
	})
	if err != nil {
		log.Error("BaikeFeed params=%+v, err=%+v", params, err)
		return nil, err
	}
	// load av inline
	loader := NewInlineCardFanoutLoader{Service: s}
	loader.setGeneralParamFromCtx(ctx)
	for _, v := range reply.List {
		if v.Res != nil {
			if v.Res.Type == 0 {
				loader.Archive.Aids = append(loader.Archive.Aids, v.Res.Rid)
			}
		}
	}
	fanout, err := loader.doChannelInlineCardFanoutLoad(ctx)
	if err != nil {
		log.Error("No baike card output err=%+v", err)
		return nil, err
	}
	feedCardCtx := fakeBuilderContext(ctx, fanout.Account.IsAttention)
	return constructFeedResponse(feedCardCtx, reply, fanout)
}

func constructFeedResponse(feedCardCtx cardschema.FeedContext, feedRsp *baikegrpc.BaikeFeedRsp, fanout *feedcard.FanoutResult) (*pedia.FeedResponse, error) {
	return &pedia.FeedResponse{
		Items:      convertBaikeFeedToItems(feedCardCtx, feedRsp.List, fanout),
		UpOffset:   feedRsp.UpOffset,
		DownOffset: feedRsp.DownOffset,
		UpMore:     feedRsp.UpMore,
		DownMore:   feedRsp.DownMore,
	}, nil
}

func convertBaikeFeedToItems(feedCardCtx cardschema.FeedContext, feedItems []*baikegrpc.FeedItem, fanout *feedcard.FanoutResult) []*pedia.FeedItem {
	if len(feedItems) == 0 {
		return nil
	}
	var res []*pedia.FeedItem
	for _, v := range feedItems {
		if v.Res == nil {
			continue
		}
		contentId := resolveBaikeItemContentId(v.FirstNid, v.SecondNid)
		if v.FirstNode != nil {
			if card := constructBaikeTitle1Card(v.FirstNode, contentId, v.FirstNid, v.SecondNid); card != nil {
				res = append(res, card)
			}
		}
		if v.SecondNode != nil {
			if card := constructBaikeTitle2Card(v.SecondNode, contentId, v.FirstNid, v.SecondNid); card != nil {
				res = append(res, card)
			}
		}
		if card := constructBaikeAvInlineCard(feedCardCtx, v.Res, fanout, v.NavNid, contentId, v.FirstNid, v.SecondNid); card != nil {
			res = append(res, card)
		}
	}
	return res
}

func resolveBaikeItemContentId(firstNid, secondNid int64) int64 {
	if secondNid > 0 {
		return secondNid
	}
	return firstNid
}

func constructBaikeAvInlineCard(feedCardCtx cardschema.FeedContext, resource *baikegrpc.Resource, fanout *feedcard.FanoutResult, nid, contentNid, firstNid, secondNid int64) *pedia.FeedItem {
	if resource.Type != 0 || fanout == nil {
		return nil
	}
	fakeRcmd := &ai.Item{ID: resource.Rid}
	card, err := feedcard.BuildLargeCoverSingleV9(feedCardCtx, resource.Rid, fakeRcmd, fanout)
	if err != nil {
		log.Error("BuildLargeCoverV9FromArchive err=%+v", err)
		return nil
	}
	v, ok := card.(*jsoncard.LargeCoverInline)
	if !ok {
		log.Error("unExpected BuildLargeCoverV9FromArchive card output card=%+v", card)
		return nil
	}
	if v.ThreePointMeta != nil && len(v.ThreePointMeta.FunctionalButtons) > 0 {
		v.ThreePointMeta.ShareOrigin = ""
		v.ThreePointMeta.ShareId = "traffic.new-channel-detail-baike.inline.three-point.click"
		v.ThreePointMeta.FunctionalButtons = removeInlineThreePointMetaDislike(v.ThreePointMeta.FunctionalButtons)
	}
	return &pedia.FeedItem{
		CardType:         "baike_large_cover_single_v9",
		NavNid:           nid,
		ContentNid:       contentNid,
		FirstNid:         firstNid,
		SecondNid:        secondNid,
		LargeCoverInline: v,
		// 注意不要和LargeCoverInline中的字段冲突
		Desc: v.Desc,
	}
}

func removeInlineThreePointMetaDislike(in []*threePointMeta.FunctionalButton) []*threePointMeta.FunctionalButton {
	const (
		_typeNotInterested = 1
	)
	out := make([]*threePointMeta.FunctionalButton, 0, len(in))
	for _, v := range in {
		if v.Type == _typeNotInterested {
			continue
		}
		out = append(out, v)
	}
	return out
}

func constructBaikeTitle2Card(node *baikegrpc.Node, contentId, firstNid, secondNid int64) *pedia.FeedItem {
	return &pedia.FeedItem{
		CardType:   "baike_title_2",
		NavNid:     node.Nid,
		BaikeTitle: node.Name,
		ContentNid: contentId,
		FirstNid:   firstNid,
		SecondNid:  secondNid,
		Desc:       node.Desc,
		Image:      node.Image,
	}
}

func constructBaikeTitle1Card(node *baikegrpc.Node, contentId, firstNid, secondNid int64) *pedia.FeedItem {
	return &pedia.FeedItem{
		CardType:   "baike_title_1",
		NavNid:     node.Nid,
		BaikeTitle: node.Name,
		ContentNid: contentId,
		FirstNid:   firstNid,
		SecondNid:  secondNid,
		Desc:       node.Desc,
		Image:      node.Image,
	}
}

func fakeBuilderContext(ctx context.Context, follow map[int64]int8) cardschema.FeedContext {
	authn, _ := auth.FromContext(ctx)
	userSession := feedcard.NewUserSession(authn.Mid, follow, &feedcard.IndexParam{})
	dev, _ := device.FromContext(ctx)
	fCtx := feedcard.NewFeedContext(userSession, feedcard.NewCtxDevice(&dev), time.Now())
	return fCtx
}
