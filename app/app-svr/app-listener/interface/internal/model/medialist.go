package model

import (
	"strconv"

	"go-common/library/log"
	xtime "go-common/library/time"
	v1 "go-gateway/app/app-svr/app-listener/interface/api/v1"
	"go-gateway/app/app-svr/app-listener/interface/conf"
)

const (
	// 空间投稿
	MediaListTypeSpace = 1
	// 稍候再看
	MediaListTypeLater = 2
	// 收藏夹
	MediaListTypeFav = 3
	// 每周必看
	MediaListTypeWeeklyRank = 4
	// 系列连播
	MediaListTypeSeries = 5
	// 合集
	MediaListTypUGCSeason = 8
)

type MediaListData struct {
	HasMore   bool            `json:"has_more"`
	Total     int             `json:"total_count"`
	MediaList []MediaListItem `json:"media_list"`
}

// 目前都是稿件内容
type MediaListItem struct {
	Bvid string `json:"bv_id"`
	Stat struct {
		Coin    int `json:"coin"`
		Fav     int `json:"collect"`
		Danmaku int `json:"danmaku"`
		View    int `json:"play"`
		Reply   int `json:"reply"`
		Share   int `json:"share"`
		Thumb   int `json:"thumb_up"`
	} `json:"cnt_info"`
	UserCoinStat struct {
		CoinAdded int `json:"coin_number"`
		MaxCoin   int `json:"max_num"`
	} `json:"coin"`
	Cover    string     `json:"cover"`
	Duration int64      `json:"duration"` // seconds
	Avid     int64      `json:"id"`       // avid/epid/songid
	Index    int        `json:"index"`    // start from 0
	Intro    string     `json:"intro"`
	Pages    int        `json:"page"` // 分p数
	PubTime  xtime.Time `json:"pubtime"`
	Tid      int        `json:"tid"` //分区id
	Title    string     `json:"title"`
	Type     int32      `json:"type"` // 元素类型 应该和收藏夹类型一致
	UpInfo   struct {
		Avatar     string `json:"face"`
		Mid        int64  `json:"mid"`
		Name       string `json:"name"`
		IsFollowed int    `json:"followed"`
	} `json:"upper"`
}

func (mli MediaListItem) ToV1PlayItem() *v1.PlayItem {
	if mli.Avid <= 0 {
		log.Warn("unexpected invalid MediaListItem(%+v)", mli)
		return nil
	}
	if Fav2Play[mli.Type] == PlayItemUnknown {
		log.Warn("unknown type(%d) for MediaListItem(%+v)", mli.Type, mli)
		return nil
	}
	return &v1.PlayItem{
		ItemType: Fav2Play[mli.Type],
		Oid:      mli.Avid,
		Et: &v1.EventTracking{
			EntityType: playType2EntityType[Fav2Play[mli.Type]],
			EntityId:   strconv.FormatInt(mli.Avid, 10),
		},
	}
}

func (mli MediaListItem) ToV1MedialistItem() *v1.MedialistItem {
	if mli.Avid <= 0 {
		log.Warn("unexpected invalid MediaListItem(%+v)", mli)
		return nil
	}
	if Fav2Play[mli.Type] == PlayItemUnknown {
		log.Warn("unknown type(%d) for MediaListItem(%+v)", mli.Type, mli)
		return nil
	}
	ret := &v1.MedialistItem{
		Item: &v1.PlayItem{
			ItemType: Fav2Play[mli.Type],
			Oid:      mli.Avid,
			Et: &v1.EventTracking{
				EntityType: playType2EntityType[Fav2Play[mli.Type]],
				EntityId:   strconv.FormatInt(mli.Avid, 10),
			},
		},
		Title:     mli.Title,
		Cover:     mli.Cover,
		Duration:  mli.Duration,
		Parts:     int32(mli.Pages),
		UpMid:     mli.UpInfo.Mid,
		UpName:    mli.UpInfo.Name,
		StatView:  int64(mli.Stat.View),
		StatReply: int64(mli.Stat.Reply),
	}
	if ret.Item.ItemType == PlayItemOGV {
		ret.State = PlayableNO
		ret.Message = conf.C.Res.Text.MsgUnsupported
	}
	return ret
}
