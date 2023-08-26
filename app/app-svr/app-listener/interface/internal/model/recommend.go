package model

import (
	"fmt"
	"strconv"

	"go-common/component/metadata/device"
	"go-common/library/log"
	v1 "go-gateway/app/app-svr/app-listener/interface/api/v1"
	"go-gateway/app/app-svr/app-listener/interface/conf"

	listenerSvc "git.bilibili.co/bapis/bapis-go/dynamic/service/listener"
)

type RcmdTopCard struct {
	Card     *listenerSvc.HeadCard
	PickInfo *SingleCollection
	Detail   *v1.DetailItem
}

const (
	_                    = iota
	TpcdChannelHistory   // 历史召回
	TpcdChannelFavFolder // 收藏夹召回
	TpcdChannelUpRecall  // up主召回
	TpcdChannelPickToday // 运营精选召回
	TpcdChannelFavRecent // 最近收藏

	// 卡片外露类型稿件
	TpcdTypeArchive = 1
	// 卡片外露类型up主
	TpcdTypeUp = 2

	TpcdClickNoOp       = 1 // 点击无动作
	TpcdClickAutoTarget = 2 // 点击起播目标稿件
	TpcdClickAutoFirst  = 3 // 点击起播列表第一个
)

var (
	channel2CardType = map[int64]v1.TopCardType{
		TpcdChannelHistory:   v1.TopCardType_LISTEN_HISTORY,
		TpcdChannelFavFolder: v1.TopCardType_FAVORITE_FOLDER,
		TpcdChannelFavRecent: v1.TopCardType_FAVORITE_FOLDER,
		TpcdChannelUpRecall:  v1.TopCardType_UP_RECALL,
		TpcdChannelPickToday: v1.TopCardType_PICK_TODAY,
	}
	click2PlayStrategy = map[int64]v1.TopCard_PlayStrategy{
		TpcdClickNoOp:       v1.TopCard_NO_INTERRUPT,
		TpcdClickAutoTarget: v1.TopCard_PLAY_TARGET,
		TpcdClickAutoFirst:  v1.TopCard_PLAY_FIRST,
	}
)

func (c RcmdTopCard) PlayImmediately() bool {
	return c.Card.GetClick() != TpcdClickNoOp
}

func (c RcmdTopCard) ToV1TopCard(currentMid int64, pos int64, dev *device.Device) (ret *v1.TopCard, dt *v1.DetailItem) {
	if c.Card == nil {
		return nil, nil
	}
	defer func() {
		if r := recover(); r != nil {
			log.Error("error converting RcmdTopCard to v1.TopCard: %v", r)
			ret = nil
		}
	}()

	channelTyp := c.Card.GetChannel().GetChannel()
	cardTyp, ok := channel2CardType[channelTyp]
	if !ok {
		return nil, nil
	}
	ret = &v1.TopCard{
		CardType:  cardTyp,
		PlayStyle: click2PlayStrategy[c.Card.GetClick()],
		Pos:       pos,
		TitleIcon: c.Card.GetChannel().GetIcon(),
	}
	switch channelTyp {
	case TpcdChannelHistory:
		dt = c.Detail
		dt.Item.SetEventTracking(v1.OpHistory)
		ret.Title = conf.C.Res.Text.TpcdHistory
		ret.Card = &v1.TopCard_ListenHistory{
			ListenHistory: &v1.TpcdHistory{
				Item: dt,
				Text: c.Card.GetArchive().GetTitle(),
				Pic:  c.Card.GetArchive().GetCover(),
			},
		}
		// 允许显示icon的时候屏蔽历史卡片的内部text
		if conf.C.Feature.RcmdHeadCardIconShow.Enabled(dev) {
			ret.Card.(*v1.TopCard_ListenHistory).ListenHistory.Text = ""
		}

	case TpcdChannelFavFolder, TpcdChannelFavRecent:
		ret.Title = conf.C.Res.Text.TpcdFavFolder
		if channelTyp == TpcdChannelFavRecent {
			ret.Title = conf.C.Res.Text.TpcdFavRecent
		}
		//nolint:gomnd
		favMlid := c.Card.GetChannel().GetBizId()*100 + currentMid%100
		dt = c.Detail
		dt.Item.SetEventTracking(v1.OpFavorite, func(et *v1.EventTracking) {
			et.TrackId = strconv.FormatInt(favMlid, 10)
			et.Batch = strconv.FormatInt(c.Card.GetChannel().GetBizType(), 10)
		})
		ret.Card = &v1.TopCard_FavFolder{
			FavFolder: &v1.TpcdFavFolder{
				Item:       dt,
				Text:       c.Card.GetArchive().GetTitle(),
				Pic:        c.Card.GetArchive().GetCover(),
				Fid:        favMlid,
				FolderType: int32(c.Card.GetChannel().GetBizType()),
			},
		}

	case TpcdChannelUpRecall:
		ret.Title = conf.C.Res.Text.TpcdUpRecall
		dt = c.Detail
		if dt != nil {
			dt.Item.SetEventTracking(v1.OpMediaList, func(et *v1.EventTracking) {
				et.Batch = strconv.FormatInt(MediaListTypeSpace, 10)
				et.TrackId = strconv.FormatInt(c.Card.GetUpper().GetMid(), 10)
			})
		}
		ret.Card = &v1.TopCard_UpRecall{
			UpRecall: &v1.TpcdUpRecall{
				UpMid:          c.Card.GetUpper().GetMid(),
				Text:           c.Card.GetUpper().GetName(),
				Avatar:         c.Card.GetUpper().GetCover(),
				MedialistType:  MediaListTypeSpace,
				MedialistBizId: c.Card.GetUpper().GetMid(),
				Item:           dt,
			},
		}

	case TpcdChannelPickToday:
		ret.Title = conf.C.Res.Text.TpcdPickToday
		dt = c.Detail
		dt.Item.SetEventTracking(v1.OpFinding, func(et *v1.EventTracking) {
			et.TrackId = strconv.FormatInt(c.Card.GetChannel().GetBizId(), 10)
			et.Batch = strconv.FormatInt(c.Card.GetChannel().GetBizType(), 10)
		})
		ret.Card = &v1.TopCard_PickToday{
			PickToday: &v1.TpcdPickToday{
				Item:       dt,
				Text:       c.PickInfo.Collection.GetTitle(),
				Pic:        c.Card.GetArchive().GetCover(),
				PickCardId: c.Card.GetChannel().GetBizId(),
				PickId:     c.Card.GetChannel().GetBizType(),
			},
		}

	default:
		panic(fmt.Sprintf("impossible case: unhandled card type %v", cardTyp))
	}
	return
}

func (c RcmdTopCard) ToV1PlayItem() *v1.PlayItem {
	if c.Card.GetArchive() == nil {
		panic(fmt.Sprintf("unexpected nil archive for RcmdTopCard(%+v)", c))
	}
	arc := c.Card.GetArchive()
	ret := &v1.PlayItem{
		ItemType: int32(arc.GetType()),
		Oid:      arc.GetAid(),
		Et: &v1.EventTracking{
			EntityType: playType2EntityType[int32(arc.GetType())],
			EntityId:   strconv.FormatInt(arc.GetAid(), 10),
		},
	}
	return ret
}
