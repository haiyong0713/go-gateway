package card

import (
	"fmt"
	"strconv"

	"go-common/library/log"
	"go-gateway/app/app-svr/app-car/interface/model"
	arcgrpc "go-gateway/app/app-svr/archive/service/api"

	episodegrpc "git.bilibili.co/bapis/bapis-go/pgc/service/season/episode"
	seasongrpc "git.bilibili.co/bapis/bapis-go/pgc/service/season/season"
)

const (
	_ugcType = "ugc"
	_pgcType = "pgc"
)

type MediaItem struct {
	ID       string      `json:"id"`
	Type     string      `json:"type"`
	Cover    string      `json:"cover"`
	Title    string      `json:"title"`
	Desc     string      `json:"desc"`
	URL      string      `json:"url"`
	Duration int64       `json:"duration"`
	Owner    *MediaOwner `json:"owner"`
	PubDate  int64       `json:"pubdate"`
	Stat     struct {
		View    int `json:"view"`
		Danmaku int `json:"danmaku"`
	} `json:"stat"`
}

type MediaOwner struct {
	Name string `json:"name"`
	Face string `json:"face"`
	URL  string `json:"url"`
}

// FromItem 目前只有/media下面的三个接口在使用
// todo 需要对小鹏渠道过来的增加参数
func (i *MediaItem) FromItem(id int64, gt string, main interface{}, materials *Materials) bool {
	const (
		_avkey = "av%d"
		_sskey = "ss%d"
	)
	switch main := main.(type) {
	case map[int64]*arcgrpc.Arc:
		am := main
		a, ok := am[id]
		if !ok {
			return false
		}
		if a.AttrVal(arcgrpc.AttrBitSteinsGate) == arcgrpc.AttrYes {
			return false
		}
		i.ID = fmt.Sprintf(_avkey, id)
		i.Type = _ugcType
		i.Cover = a.Pic
		i.Title = a.Title
		i.Desc = a.Desc
		i.URL = model.FillURI(gt, model.PlatCar, 0, strconv.FormatInt(id, 10), model.ParamHandler(materials.Prune, a.FirstCid, 0, model.EntranceCommonSearch, "", i.Title))
		i.URL = fmt.Sprintf("%s&from=beauty_space&resource=card", i.URL)
		i.Duration = a.Duration
		i.Owner = &MediaOwner{
			Name: a.Author.Name,
			Face: a.Author.Face,
			URL:  model.FillURI(model.GotoSpace, model.PlatCar, 0, strconv.FormatInt(a.Author.Mid, 10), nil),
		}
		i.PubDate = a.PubDate.Time().Unix()
		i.Stat.View = int(a.Stat.View)
		i.Stat.Danmaku = int(a.Stat.Danmaku)
	case map[int32]*episodegrpc.EpisodeCardsProto:
		sm := main
		s, ok := sm[int32(id)]
		if !ok {
			return false
		}
		i.ID = fmt.Sprintf(_sskey, s.Season.SeasonId)
		i.Type = _pgcType
		i.Cover = s.Cover
		i.Title = s.Season.Title
		i.URL = model.FillURI(gt, model.PlatCar, 0, strconv.Itoa(int(s.Season.SeasonId)), model.ParamHandler(materials.Prune, int64(s.EpisodeId), 0, model.EntranceCommonSearch, "", i.Title))
		i.URL = fmt.Sprintf("%s&from=beauty_space&resource=card", i.URL)
	case map[int32]*seasongrpc.CardInfoProto:
		sm := main
		s, ok := sm[int32(id)]
		if !ok {
			return false
		}
		i.ID = fmt.Sprintf(_sskey, s.SeasonId)
		i.Type = _pgcType
		i.Cover = s.Cover
		i.Title = s.Title
		i.URL = model.FillURI(gt, model.PlatCar, 0, strconv.Itoa(int(s.SeasonId)), model.ParamHandler(materials.Prune, int64(s.FirstEp), 0, model.EntranceCommonSearch, "", i.Title))
		i.URL = fmt.Sprintf("%s&from=beauty_space&resource=card", i.URL)
	default:
		log.Warn("MediaItem App From: unexpected type %T", main)
		return false
	}
	return true
}

type MediaItemWeb struct {
	Oid      string `json:"oid"`
	Cid      string `json:"cid"`
	OType    string `json:"otype"`
	Cover    string `json:"cover"`
	Title    string `json:"title"`
	Desc     string `json:"desc"`
	Duration int64  `json:"duration"`
}

func (i *MediaItemWeb) FromMediaItemWeb(id int64, gt string, main interface{}) bool {
	switch main := main.(type) {
	case map[int64]*arcgrpc.Arc:
		am := main
		a, ok := am[id]
		if !ok {
			return false
		}
		if a.AttrVal(arcgrpc.AttrBitSteinsGate) == arcgrpc.AttrYes {
			return false
		}
		i.Oid = strconv.FormatInt(a.Aid, 10)
		i.Cid = strconv.FormatInt(a.FirstCid, 10)
		i.OType = _ugcType
		i.Cover = a.Pic
		i.Title = a.Title
		i.Desc = a.Desc
		i.Duration = a.Duration
	case map[int32]*episodegrpc.EpisodeCardsProto:
		sm := main
		s, ok := sm[int32(id)]
		if !ok {
			return false
		}
		i.Oid = strconv.Itoa(int(s.Season.SeasonId))
		i.Cid = strconv.Itoa(int(s.EpisodeId))
		i.OType = _pgcType
		i.Cover = s.Cover
		i.Title = s.Season.Title
		i.Duration = int64(s.Duration)
	case map[int32]*seasongrpc.CardInfoProto:
		sm := main
		s, ok := sm[int32(id)]
		if !ok {
			return false
		}
		i.Oid = strconv.Itoa(int(s.SeasonId))
		i.Cid = strconv.Itoa(int(s.NewEp.Id))
		i.OType = _pgcType
		i.Cover = s.Cover
		i.Title = s.Title
		i.Duration = int64(s.NewEp.Duration)
		i.Desc = s.Evaluate
	default:
		log.Warn("MediaItemWeb H5 From: unexpected type %T", main)
	}
	return true
}
