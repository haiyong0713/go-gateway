package card

import (
	"strconv"

	"go-common/library/log"
	"go-common/library/time"
	"go-gateway/app/app-svr/app-car/interface/model"
	"go-gateway/app/app-svr/app-car/interface/model/card/ai"
	"go-gateway/app/app-svr/app-car/interface/model/card/operate"
	"go-gateway/app/app-svr/app-car/interface/model/favorite"
	arcgrpc "go-gateway/app/app-svr/archive/service/api"

	hisApi "git.bilibili.co/bapis/bapis-go/community/interface/history"
	cardappgrpc "git.bilibili.co/bapis/bapis-go/pgc/service/card/app"
	episodegrpc "git.bilibili.co/bapis/bapis-go/pgc/service/season/episode"
	seasongrpc "git.bilibili.co/bapis/bapis-go/pgc/service/season/season"
)

func h5Handle(cardGoto model.CardGt, cardType model.CardType, rcmd *ai.Item, materials *Materials) (hander Handler) {
	base := &Base{CardType: cardType, Rcmd: rcmd, Materials: materials}
	switch cardType {
	case model.SmallCoverV1:
		hander = &H5SmallCoverV1{Base: base}
	case model.SmallCoverV2:
		hander = &H5SmallCoverV2{Base: base}
	case model.SmallCoverV3:
		hander = &H5SmallCoverV3{Base: base}
	case model.SmallCoverV4:
		hander = &H5SmallCoverV4{Base: base}
	case model.VerticalCoverV1:
		hander = &H5VerticalCoverV1{Base: base}
	default:
		switch cardGoto {
		case model.CardGotoAv:
			base.CardType = model.SmallCoverV1
			hander = &H5SmallCoverV1{Base: base}
		case model.CardGotoPGC:
			base.CardType = model.VerticalCoverV1
			hander = &H5VerticalCoverV1{Base: base}
		default:
			log.Error("h5Handle Fail to build handler, cardType=%s cardGoto=%s ai={%+v}", cardType, cardGoto, rcmd)
		}
	}
	return
}

type H5SmallCoverV1 struct {
	*Base
	CoverLeftText1 string     `json:"cover_left_text_1,omitempty"`
	CoverLeftIcon1 model.Icon `json:"cover_left_icon_1,omitempty"`
	CoverLeftText2 string     `json:"cover_left_text_2,omitempty"`
	CoverLeftIcon2 model.Icon `json:"cover_left_icon_2,omitempty"`
	CoverRightText string     `json:"cover_right_text,omitempty"`
	Desc           string     `json:"desc,omitempty"`
	CoverRightIcon model.Icon `json:"cover_right_icon,omitempty"`
}

func (c *H5SmallCoverV1) From(main interface{}, op *operate.Card) bool {
	if op == nil {
		return false
	}
	switch main := main.(type) {
	case map[int64]*arcgrpc.Arc:
		am := main
		a, ok := am[op.ID]
		if !ok {
			log.Warn("SmallCoverV1 H5 From: aid(%d) is null", op.ID)
			return false
		}
		if !model.AvIsNormal(a) {
			return false
		}
		// 过滤原因：互动视频
		if a.AttrVal(arcgrpc.AttrBitSteinsGate) == arcgrpc.AttrYes {
			c.Filter = model.FilterAttrBitSteinsGate
		}
		c.Base.fromH5(strconv.FormatInt(a.Aid, 10), strconv.FormatInt(a.FirstCid, 10), a.Pic, a.Title, model.GotoAv)
		c.Base.FromRequestParam(c.Materials.Prune, a.FirstCid, op.Rid, op.Entrance, op.FollowType, op.KeyWord)
		c.CoverLeftText1 = model.StatString(a.Stat.View, "")
		c.CoverLeftIcon1 = model.IconPlay
		c.CoverLeftText2 = model.StatString(a.Stat.Danmaku, "")
		c.CoverLeftIcon2 = model.IconDanmaku
		c.CoverRightText = model.DurationString(a.Duration)
		c.Desc = a.Author.Name
		if op.Entrance == model.EntranceSpace {
			// 空间页展示投稿时间
			c.Desc = model.PubDataString(a.PubDate.Time())
		}
		// 收藏夹
		c.RequestParam.FavID = op.FavID
		c.RequestParam.Vmid = op.Vmid
	case map[int32]*episodegrpc.EpisodeCardsProto:
		sm := main
		s, ok := sm[int32(op.ID)]
		if !ok {
			log.Warn("SmallCoverV1 H5 From: epid(%d) is null", op.ID)
			return false
		}
		c.Base.fromH5(strconv.Itoa(int(s.GetSeason().SeasonId)), strconv.Itoa(int(s.EpisodeId)), s.Cover, s.Season.Title, model.GotoPGC)
		c.Base.FromRequestParam(c.Materials.Prune, int64(s.EpisodeId), op.Rid, op.Entrance, op.FollowType, op.KeyWord)
		if s.Season.Stat != nil {
			c.CoverLeftText1 = model.StatString(int32(s.Season.Stat.View), "")
			c.CoverLeftIcon1 = model.IconPlay
		}
		c.Desc = s.Season.NewEpShow
	case *favorite.MediaList:
		f := main
		// 如果收藏夹内一个内容都没有直接不下发
		if f.MediaCount == 0 {
			return false
		}
		c.Base.fromH5(strconv.FormatInt(op.ID, 10), "", f.Cover, f.Title, model.GotoFavorite)
		c.Base.FromRequestParam(nil, 0, 0, op.Entrance, op.FollowType, op.KeyWord)
		c.RequestParam.FavID = op.ID
		c.RequestParam.Vmid = f.Mid
		c.Desc = model.FavoriteCountString(int32(f.MediaCount))
		switch op.Business {
		case model.EntranceUpFavorite:
			c.Desc = c.Desc + " · " + f.Upper.Name
		}
		c.CoverRightIcon = model.IconFavoritePlay
	case nil:
		switch op.CardGoto {
		case model.CardGotoTopView:
			c.Base.fromH5("", "", op.Cover, "稍后再看", model.GotoTopView)
			c.Desc = op.Desc
		default:
			return false
		}
	default:
		log.Warn("SmallCoverV1 H5 From: unexpected type %T", main)
		return false
	}
	return true
}

func (c *H5SmallCoverV1) Get() *Base {
	return c.Base
}

type H5SmallCoverV2 struct {
	*Base
	CoverLeftText1  string       `json:"cover_left_text_1,omitempty"`
	Desc            string       `json:"desc,omitempty"`
	CoverBadgeStyle *ReasonStyle `json:"cover_bage_style,omitempty"`
}

func (c *H5SmallCoverV2) From(main interface{}, op *operate.Card) bool {
	if op == nil {
		return false
	}
	switch main := main.(type) {
	case *cardappgrpc.CardSeasonProto:
		s := main
		if s.NewEp == nil {
			return false
		}
		cid := int64(s.NewEp.Id)
		if op.Cid > 0 {
			cid = op.Cid
		}
		c.Base.fromH5(strconv.Itoa(int(s.SeasonId)), strconv.FormatInt(cid, 10), s.Cover, s.Title, model.GotoPGC)
		c.Base.FromRequestParam(c.Materials.Prune, cid, op.Rid, op.Entrance, op.FollowType, op.KeyWord)
		c.CoverBadgeStyle = reasonStyleFrom(model.PGCBageType[s.BadgeType], s.Badge)
		if s.NewEp != nil {
			c.CoverLeftText1 = s.NewEp.IndexShow
		}
		if s.Progress != nil {
			// time = 0表示未观看
			c.Desc = model.HighLightString(s.Progress.IndexShow)
			if s.Progress.Time == 0 {
				c.Desc = s.Progress.IndexShow
			}
		}
	default:
		log.Warn("SmallCoverV2 H5 From: unexpected type %T", main)
		return false
	}
	return true
}

func (c *H5SmallCoverV2) Get() *Base {
	return c.Base
}

type H5SmallCoverV3 struct {
	*Base
	CoverRightText string `json:"cover_right_text,omitempty"`
	Desc           string `json:"desc,omitempty"`
	// 历史记录字段
	ViewAt      int64        `json:"view_at,omitempty"`
	Kid         int64        `json:"kid,omitempty"`
	HistoryArgs *HistoryArgs `json:"history_args,omitempty"`
}

func (c *H5SmallCoverV3) From(main interface{}, op *operate.Card) bool {
	if op == nil {
		return false
	}
	his := c.FromHis(c.Rcmd.Card)
	if his == nil {
		log.Warn("SmallCoverV3 H5 his is nil id(%d)", op.ID)
		return false
	}
	switch main := main.(type) {
	case map[int64]*arcgrpc.ViewReply:
		am := main
		a, ok := am[op.ID]
		if !ok {
			log.Warn("SmallCoverV3 H5 From: aid(%d) is null", op.ID)
			return false
		}
		if !model.AvIsNormalView(a) {
			return false
		}
		// 过滤原因：互动视频
		if a.AttrVal(arcgrpc.AttrBitSteinsGate) == arcgrpc.AttrYes {
			c.Filter = model.FilterAttrBitSteinsGate
		}
		c.Base.fromH5(strconv.FormatInt(a.Aid, 10), strconv.FormatInt(a.FirstCid, 10), a.Pic, a.Title, model.GotoAv)
		c.Base.FromRequestParam(c.Materials.Prune, a.FirstCid, op.Rid, op.Entrance, op.FollowType, op.KeyWord)
		c.HistoryArgs.Mid = a.Author.Mid
		c.HistoryArgs.Name = a.Author.Name
		for _, p := range a.Pages {
			if p.Cid == his.Cid {
				if a.AttrVal(arcgrpc.AttrBitSteinsGate) != arcgrpc.AttrYes { // 互动视频进度不展示
					c.HistoryArgs.Progress = p.Duration
					c.HistoryArgs.Page = p.Page
					c.HistoryArgs.Subtitle = p.Part
					c.HistoryArgs.Duration = p.Duration
					c.HistoryArgs.Cid = p.Cid
				}
				c.Cid = strconv.FormatInt(p.Cid, 10)
				c.Base.FromRequestParam(c.Materials.Prune, a.FirstCid, p.Cid, op.Entrance, op.FollowType, op.KeyWord)
				break
			}
		}
	case map[int32]*episodegrpc.EpisodeCardsProto:
		eps := main
		ep, eok := eps[int32(op.Cid)]
		a, aok := c.Materials.ViewReplym[op.ID]
		if !eok || !aok || ep.Season == nil {
			log.Warn("SmallCoverV3 H5 From: epid(%d) is null", op.Cid)
			return false
		}
		c.Base.fromH5(strconv.Itoa(int(ep.Season.SeasonId)), strconv.FormatInt(op.Cid, 10), ep.Cover, ep.Season.Title, model.GotoPGC)
		c.Base.FromRequestParam(c.Materials.Prune, op.Cid, op.Rid, op.Entrance, op.FollowType, op.KeyWord)
		for _, p := range a.Pages {
			if p.Cid == his.Cid {
				c.HistoryArgs.Duration = p.Duration
				c.HistoryArgs.Progress = p.Duration
				break
			}
		}
	default:
		log.Warn("SmallCoverV3 H5 From: unexpected goto %s", op.CardGoto)
		return false
	}
	if his.Pro > -1 {
		// l.pro == -1表示已看完
		c.HistoryArgs.Progress = his.Pro
	}
	if c.HistoryArgs.Duration > 0 {
		c.CoverRightText = model.HisDurationString(c.HistoryArgs.Progress) + " / " + model.HisDurationString(int64(c.HistoryArgs.Duration))
	}
	return true
}

func (c *H5SmallCoverV3) FromHis(main interface{}) *hisApi.ModelResource {
	switch main := main.(type) {
	case *hisApi.ModelResource:
		l := main
		c.Kid = l.Kid
		c.ViewAt = l.Unix
		c.HistoryArgs = &HistoryArgs{
			Business: l.Business,
		}
		c.Desc = model.HisPubDataString(time.Time(l.Unix).Time())
		return l
	default:
		log.Warn("SmallCoverV3 H5 his is nil")
	}
	return nil
}

func (c *H5SmallCoverV3) Get() *Base {
	return c.Base
}

type H5VerticalCoverV1 struct {
	*Base
	Desc            string       `json:"desc,omitempty"`
	CoverBadgeStyle *ReasonStyle `json:"cover_bage_style,omitempty"`
}

func (c *H5VerticalCoverV1) From(main interface{}, op *operate.Card) bool {
	if op == nil {
		return false
	}
	switch main := main.(type) {
	case map[int32]*seasongrpc.CardInfoProto:
		scm := main
		sc, ok := scm[int32(op.ID)]
		if !ok || sc.FirstEpInfo == nil {
			log.Warn("VerticalCoverV1 H5 From: seasonId(%d) is null", op.ID)
			return false
		}
		c.Base.fromH5(strconv.Itoa(int(sc.SeasonId)), strconv.Itoa(int(sc.FirstEpInfo.Id)), sc.Cover, sc.Title, model.GotoPGC)
		c.Base.FromRequestParam(c.Materials.Prune, int64(sc.FirstEpInfo.Id), op.Rid, op.Entrance, op.FollowType, op.KeyWord)
		c.CoverBadgeStyle = reasonStyleFrom(model.PGCBageType[sc.BadgeType], sc.Badge)
		if sc.Stat != nil {
			c.Desc = model.BangumiFavString(int32(sc.Stat.Follow), sc.SeasonType)
		}
	default:
		log.Warn("VerticalCoverV1 H5 From: unexpected type %T", main)
		return false
	}
	return true
}

func (c *H5VerticalCoverV1) Get() *Base {
	return c.Base
}

type H5SmallCoverV4 struct {
	*Base
	CoverLeftText1  string       `json:"cover_left_text_1,omitempty"`
	CoverLeftIcon1  model.Icon   `json:"cover_left_icon_1,omitempty"`
	CoverRightText  string       `json:"cover_right_text,omitempty"`
	Desc            string       `json:"desc,omitempty"`
	CoverBadgeStyle *ReasonStyle `json:"cover_bage_style,omitempty"`
}

func (c *H5SmallCoverV4) From(main interface{}, op *operate.Card) bool {
	if op == nil {
		return false
	}
	switch main := main.(type) {
	case map[int64]*arcgrpc.ViewReply:
		am := main
		a, ok := am[op.ID]
		if !ok {
			log.Warn("SmallCoverV4 H5 From: aid(%d) is null", op.ID)
			return false
		}
		if !model.AvIsNormalView(a) {
			return false
		}
		// 过滤原因：互动视频
		if a.AttrVal(arcgrpc.AttrBitSteinsGate) == arcgrpc.AttrYes {
			c.Filter = model.FilterAttrBitSteinsGate
		}
		// 默认给首个cid
		cid := a.FirstCid
		// 获取历史记录里面的cid
		if op.Cid > 0 {
			cid = op.Cid
		}
		c.Base.fromH5(strconv.FormatInt(a.Aid, 10), strconv.FormatInt(cid, 10), a.Pic, a.Title, model.GotoAv)
		c.CoverLeftText1 = model.StatString(a.Stat.View, "")
		c.CoverLeftIcon1 = model.IconPlay
		c.CoverRightText = model.DurationString(a.Duration)
		c.Desc = a.Author.Name
	case map[int64]*arcgrpc.Arc:
		am := main
		a, ok := am[op.ID]
		if !ok {
			log.Warn("SmallCoverV4 H5 From: aid(%d) is null", op.ID)
			return false
		}
		if !model.AvIsNormal(a) {
			return false
		}
		// 过滤原因：互动视频
		if a.AttrVal(arcgrpc.AttrBitSteinsGate) == arcgrpc.AttrYes {
			c.Filter = model.FilterAttrBitSteinsGate
		} // 默认给首个cid
		cid := a.FirstCid
		// 获取历史记录里面的cid
		if op.Cid > 0 {
			cid = op.Cid
		}
		c.Base.fromH5(strconv.FormatInt(a.Aid, 10), strconv.FormatInt(cid, 10), a.Pic, a.Title, model.GotoAv)
		c.CoverLeftText1 = model.StatString(a.Stat.View, "")
		c.CoverLeftIcon1 = model.IconPlay
		c.CoverRightText = model.DurationString(a.Duration)
		c.Desc = a.Author.Name
		// 空间页展示投稿时间
		if op.Entrance == model.EntranceSpace {
			c.Desc = model.PubDataString(a.PubDate.Time())
		}
	case map[int32]*seasongrpc.CardInfoProto:
		scm := main
		sc, ok := scm[int32(op.ID)]
		if !ok || sc.FirstEpInfo == nil {
			log.Warn("SmallCoverV4 H5 From: seasonId(%d) or ep is null", op.ID)
			return false
		}
		cid := int64(sc.FirstEpInfo.Id)
		if op.Cid > 0 {
			cid = op.Cid
		}
		c.Base.fromH5(strconv.Itoa(int(sc.SeasonId)), strconv.FormatInt(cid, 10), sc.FirstEpInfo.Cover, sc.Title, model.GotoPGC)
		if sc.Stat != nil {
			c.CoverLeftIcon1 = model.IconPlay
			c.CoverLeftText1 = model.StatString(int32(sc.Stat.View), "")
		}
		c.CoverBadgeStyle = reasonStyleFrom(model.PGCBageType[sc.BadgeType], sc.Badge)
		c.Desc = sc.NewEp.IndexShow
	case map[int32]*episodegrpc.EpisodeCardsProto:
		sm := main
		s, ok := sm[int32(op.ID)]
		if !ok {
			log.Warn("SmallCoverV4 H5 From: epid(%d) is null", op.ID)
			return false
		}
		c.Base.fromH5(strconv.Itoa(int(s.GetSeason().SeasonId)), strconv.FormatInt(int64(s.EpisodeId), 10), s.Cover, s.Season.Title, model.GotoPGC)
		if s.Season.Stat != nil {
			c.CoverLeftText1 = model.StatString(int32(s.Season.Stat.View), "")
			c.CoverLeftIcon1 = model.IconPlay
		}
		c.CoverBadgeStyle = reasonStyleFrom(model.PGCBageType[s.Season.BadgeType], s.Season.Badge)
		c.Desc = s.Season.NewEpShow
	default:
		log.Warn("SmallCoverV4 H5 From: unexpected type %T", main)
		return false
	}
	return true
}

func (c *H5SmallCoverV4) Get() *Base {
	return c.Base
}
