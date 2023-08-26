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

func singleHandle(cardGoto model.CardGt, cardType model.CardType, rcmd *ai.Item, materials *Materials) (hander Handler) {
	base := &Base{CardType: cardType, CardGoto: cardGoto, Rcmd: rcmd, Materials: materials}
	switch cardType {
	case model.SmallCoverV1:
		hander = &SmallCoverV1{Base: base}
	case model.SmallCoverV2:
		hander = &SmallCoverV2{Base: base}
	case model.SmallCoverV3:
		hander = &SmallCoverV3{Base: base}
	case model.SmallCoverV4:
		hander = &SmallCoverV4{Base: base}
	case model.VerticalCoverV1:
		hander = &VerticalCoverV1{Base: base}
	case model.BannerV1:
		hander = &BannerV1{Base: base}
	case model.FmV1:
		hander = &FM{Base: base}
	default:
		switch cardGoto {
		case model.CardGotoAv:
			base.CardType = model.SmallCoverV1
			hander = &SmallCoverV1{Base: base}
		case model.CardGotoPGC:
			base.CardType = model.VerticalCoverV1
			hander = &VerticalCoverV1{Base: base}
		case model.CardGotoDefalutFavorite, model.CardGotoUserFavorite:
			base.CardType = model.SmallCoverV1
			hander = &SmallCoverV1{Base: base}
		default:
			log.Error("singleHandle Fail to build handler, cardType=%s cardGoto=%s ai={%+v}", cardType, cardGoto, rcmd)
		}
	}
	return
}

type SmallCoverV1 struct {
	*Base
	CoverLeftText1 string     `json:"cover_left_text_1,omitempty"`
	CoverLeftIcon1 model.Icon `json:"cover_left_icon_1,omitempty"`
	CoverLeftText2 string     `json:"cover_left_text_2,omitempty"`
	CoverLeftIcon2 model.Icon `json:"cover_left_icon_2,omitempty"`
	CoverRightText string     `json:"cover_right_text,omitempty"`
	DescIcon       model.Icon `json:"desc_icon,omitempty"`
	Desc           string     `json:"desc,omitempty"`
	CoverRightIcon model.Icon `json:"cover_right_icon,omitempty"`
	DynCtime       int64      `json:"dyn_ctime,omitempty"`
	Banner         *Banner    `json:"banner,omitempty"`
}

type Banner struct {
	CoverLeftText1 string `json:"cover_left_text_1,omitempty"`
	CoverLeftText2 string `json:"cover_left_text_2,omitempty"`
	CoverLeftText3 string `json:"cover_left_text_3,omitempty"`
}

func (c *SmallCoverV1) From(main interface{}, op *operate.Card) bool {
	if op == nil {
		return false
	}
	switch main := main.(type) {
	case map[int64]*arcgrpc.Arc:
		am := main
		a, ok := am[op.ID]
		if !ok {
			log.Warn("SmallCoverV1 From: aid(%d) is null", op.ID)
			return false
		}
		// mid > int32直接抛弃这张卡
		if model.CheckMidMaxInt32(a.Author.Mid, op.Build) {
			return false
		}
		// 过滤原因：互动视频
		if a.AttrVal(arcgrpc.AttrBitSteinsGate) == arcgrpc.AttrYes {
			c.Filter = model.FilterAttrBitSteinsGate
		}
		c.Base.from(op.Plat, op.Build, strconv.FormatInt(a.Aid, 10), strconv.FormatInt(a.FirstCid, 10), a.Pic, a.Title, model.GotoAv, strconv.FormatInt(a.Aid, 10), model.ParamHandler(c.Materials.Prune, a.FirstCid, op.Rid, op.Entrance, op.FollowType, op.KeyWord))
		c.CoverLeftText1 = model.StatString(a.Stat.View, "")
		c.CoverLeftIcon1 = model.IconPlay
		c.CoverLeftText2 = model.StatString(a.Stat.Danmaku, "")
		c.CoverLeftIcon2 = model.IconDanmaku
		c.CoverRightText = model.DurationString(a.Duration)
		c.Desc = a.Author.Name
		if op.Entrance != model.EntranceSpace {
			c.DescIcon = model.IconUp
		}
		switch op.Entrance {
		case model.EntranceSpace:
			// 空间页展示投稿时间
			c.Desc = model.PubDataString(a.PubDate.Time())
			c.URI = model.FillURI("", op.Plat, op.Build, c.URI, model.SpaceHandler(a.Author.Mid))
		case model.EntranceMediaList:
			c.URI = model.FillURI("", op.Plat, op.Build, c.URI, model.FavHandler(op.FavID, op.Vmid))
		}
		c.Cid = strconv.FormatInt(a.FirstCid, 10)
	case map[int64]*arcgrpc.ArcPlayer:
		am := main
		ap, ok := am[op.ID]
		if !ok {
			log.Warn("SmallCoverV1 From: aid(%d) is null", op.ID)
			return false
		}
		a := ap.Arc
		// 过滤原因：互动视频
		if a.AttrVal(arcgrpc.AttrBitSteinsGate) == arcgrpc.AttrYes {
			c.Filter = model.FilterAttrBitSteinsGate
		}
		c.Base.from(op.Plat, op.Build, strconv.FormatInt(a.Aid, 10), strconv.FormatInt(a.FirstCid, 10), a.Pic, a.Title, model.GotoAv, strconv.FormatInt(a.Aid, 10), model.ParamHandler(c.Materials.Prune, a.FirstCid, op.Rid, op.Entrance, op.FollowType, op.KeyWord))
		c.URI = model.FillURI("", op.Plat, op.Build, c.URI, model.ArcPlayHandler(ap.PlayerInfo[a.FirstCid]))
		c.CoverLeftText1 = model.StatString(a.Stat.View, "")
		c.CoverLeftIcon1 = model.IconPlay
		c.CoverLeftText2 = model.StatString(a.Stat.Danmaku, "")
		c.CoverLeftIcon2 = model.IconDanmaku
		c.CoverRightText = model.DurationString(a.Duration)
		c.Desc = a.Author.Name
		if op.Entrance != model.EntranceSpace {
			c.DescIcon = model.IconUp
		}
		switch op.Entrance {
		case model.EntranceSpace:
			// 空间页展示投稿时间
			c.Desc = model.PubDataString(a.PubDate.Time())
			c.URI = model.FillURI("", op.Plat, op.Build, c.URI, model.SpaceHandler(a.Author.Mid))
			// 空间页banner字段
			c.Banner = &Banner{
				CoverLeftText1: model.DurationString(a.Duration),
				CoverLeftText2: model.StatString(a.Stat.View, "播放"),
				CoverLeftText3: model.StatString(a.Stat.Danmaku, "弹幕"),
			}
		case model.EntranceMediaList:
			c.URI = model.FillURI("", op.Plat, op.Build, c.URI, model.FavHandler(op.FavID, op.Vmid))
		}
		c.Cid = strconv.FormatInt(a.FirstCid, 10)
	case map[int32]*episodegrpc.EpisodeCardsProto:
		sm := main
		s, ok := sm[int32(op.ID)]
		if !ok {
			log.Warn("SmallCoverV1 From: epid(%d) is null", op.ID)
			return false
		}
		c.Base.from(op.Plat, op.Build, strconv.Itoa(int(s.GetSeason().SeasonId)), strconv.Itoa(int(s.EpisodeId)), s.Cover, s.Season.Title, model.GotoPGC, strconv.Itoa(int(s.GetSeason().SeasonId)), model.ParamHandler(c.Materials.Prune, int64(s.EpisodeId), op.Rid, op.Entrance, op.FollowType, op.KeyWord))
		if s.Season.Stat != nil {
			c.CoverLeftText1 = model.StatString(int32(s.Season.Stat.View), "")
			c.CoverLeftIcon1 = model.IconPlay
		}
		c.Desc = s.Season.NewEpShow
		switch op.Entrance {
		case model.EntranceMediaList:
			c.URI = model.FillURI("", op.Plat, op.Build, c.URI, model.FavHandler(op.FavID, op.Vmid))
		}
	case *favorite.MediaList:
		f := main
		// 如果收藏夹内一个内容都没有直接不下发
		if f.MediaCount == 0 {
			return false
		}
		// mid > int32直接抛弃这张卡
		if model.CheckMidMaxInt32(f.Mid, op.Build) {
			return false
		}
		c.Base.from(op.Plat, op.Build, strconv.FormatInt(op.ID, 10), "", f.Cover, f.Title, model.GotoFavorite, strconv.FormatInt(op.ID, 10), model.MediaHandler(f.Mid))
		c.Desc = model.FavoriteCountString(int32(f.MediaCount))
		switch op.Business {
		case model.EntranceUpFavorite:
			c.Desc = c.Desc + " · " + f.Upper.Name
		}
		c.CoverRightIcon = model.IconFavoritePlay
	case map[int64]*arcgrpc.ViewReply:
		am := main
		a, ok := am[op.ID]
		if !ok {
			log.Warn("SmallCoverV1 From: aid(%d) is null", op.ID)
			return false
		}
		// mid > int32直接抛弃这张卡
		if model.CheckMidMaxInt32(a.Author.Mid, op.Build) {
			return false
		}
		// 过滤原因：互动视频
		if a.AttrVal(arcgrpc.AttrBitSteinsGate) == arcgrpc.AttrYes {
			c.Filter = model.FilterAttrBitSteinsGate
		}
		c.Base.from(op.Plat, op.Build, strconv.FormatInt(a.Aid, 10), strconv.FormatInt(a.FirstCid, 10), a.Pic, a.Title, model.GotoAv, strconv.FormatInt(a.Aid, 10), model.ParamHandler(c.Materials.Prune, a.FirstCid, op.Rid, op.Entrance, op.FollowType, op.KeyWord))
		c.CoverLeftText1 = model.StatString(a.Stat.View, "")
		c.CoverLeftIcon1 = model.IconPlay
		c.CoverLeftText2 = model.StatString(a.Stat.Danmaku, "")
		c.CoverLeftIcon2 = model.IconDanmaku
		c.CoverRightText = model.DurationString(a.Duration)
		c.Desc = a.Author.Name
		if op.Entrance != model.EntranceSpace {
			c.DescIcon = model.IconUp
		}
		switch op.Entrance {
		case model.EntranceSpace:
			// 空间页展示投稿时间
			c.Desc = model.PubDataString(a.PubDate.Time())
			c.URI = model.FillURI("", op.Plat, op.Build, c.URI, model.SpaceHandler(a.Author.Mid))
		case model.EntranceMediaList:
			c.URI = model.FillURI("", op.Plat, op.Build, c.URI, model.FavHandler(op.FavID, op.Vmid))
		}
		c.Cid = strconv.FormatInt(a.FirstCid, 10)
	case nil:
		switch op.CardGoto {
		case model.CardGotoTopView:
			c.Base.from(op.Plat, op.Build, "", "", op.Cover, "稍后再看", model.GotoTopView, "", nil)
			c.Desc = op.Desc
		default:
			return false
		}
	default:
		log.Warn("SmallCoverV1 From: unexpected type %T", main)
		return false
	}
	// ext
	c.DynCtime = op.DynCtime
	return true
}

func (c *SmallCoverV1) Get() *Base {
	return c.Base
}

type SmallCoverV2 struct {
	*Base
	CoverLeftText1  string       `json:"cover_left_text_1,omitempty"`
	Desc            string       `json:"desc,omitempty"`
	CoverBadgeStyle *ReasonStyle `json:"cover_bage_style,omitempty"`
}

func (c *SmallCoverV2) From(main interface{}, op *operate.Card) bool {
	if op == nil {
		return false
	}
	switch main := main.(type) {
	case *cardappgrpc.CardSeasonProto:
		s := main
		c.Base.from(op.Plat, op.Build, strconv.Itoa(int(s.SeasonId)), strconv.FormatInt(op.Cid, 10), s.Cover, s.Title, model.GotoPGC, strconv.Itoa(int(s.SeasonId)), model.ParamHandler(c.Materials.Prune, op.Cid, op.Rid, op.Entrance, op.FollowType, op.KeyWord))
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
		log.Warn("SmallCoverV2 From: unexpected type %T", main)
		return false
	}
	return true
}

func (c *SmallCoverV2) Get() *Base {
	return c.Base
}

type SmallCoverV3 struct {
	*Base
	CoverRightText string `json:"cover_right_text,omitempty"`
	Desc           string `json:"desc,omitempty"`
	// 历史记录字段
	ViewAt      int64        `json:"view_at,omitempty"`
	Kid         int64        `json:"kid,omitempty"`
	HistoryArgs *HistoryArgs `json:"history_args,omitempty"`
}

func (c *SmallCoverV3) From(main interface{}, op *operate.Card) bool {
	if op == nil {
		return false
	}
	switch main := main.(type) {
	case *hisApi.ModelResource:
		l := main
		c.Kid = l.Kid
		c.ViewAt = l.Unix
		c.HistoryArgs = &HistoryArgs{
			Business: l.Business,
		}
		c.Desc = model.HisPubDataString(time.Time(l.Unix).Time())
		switch op.CardGoto {
		case model.CardGotoAv:
			a, ok := c.Materials.ViewReplym[l.Oid]
			if !ok {
				log.Warn("SmallCoverV3 From: id(%d) is null", l.Oid)
				return false
			}
			// mid > int32直接抛弃这张卡
			if model.CheckMidMaxInt32(a.Author.Mid, op.Build) {
				return false
			}
			// 过滤原因：互动视频
			if a.AttrVal(arcgrpc.AttrBitSteinsGate) == arcgrpc.AttrYes {
				c.Filter = model.FilterAttrBitSteinsGate
			}
			c.Base.from(op.Plat, op.Build, strconv.FormatInt(a.Aid, 10), strconv.FormatInt(a.FirstCid, 10), a.Pic, a.Title, model.GotoAv, strconv.FormatInt(a.Aid, 10), model.ParamHandler(c.Materials.Prune, a.FirstCid, op.Rid, op.Entrance, op.FollowType, op.KeyWord))
			c.HistoryArgs.Mid = a.Author.Mid
			c.HistoryArgs.Name = a.Author.Name
			for _, p := range a.Pages {
				if p.Cid == l.Cid {
					if a.AttrVal(arcgrpc.AttrBitSteinsGate) != arcgrpc.AttrYes { // 互动视频进度不展示
						c.HistoryArgs.Progress = p.Duration
						c.HistoryArgs.Page = p.Page
						c.HistoryArgs.Subtitle = p.Part
						c.HistoryArgs.Duration = p.Duration
						c.HistoryArgs.Cid = p.Cid
					}
					c.URI = model.FillURI(model.GotoAv, op.Plat, op.Build, strconv.FormatInt(a.Aid, 10), model.ParamHandler(c.Materials.Prune, p.Cid, op.Rid, op.Entrance, op.FollowType, op.KeyWord))
					break
				}
			}
		case model.CardGotoPGC:
			s, smok := c.Materials.EpisodeCardsProtom[int32(l.Epid)]
			a, arcok := c.Materials.ViewReplym[l.Oid]
			if !arcok || !smok || s.Season == nil {
				log.Warn("SmallCoverV3 From: epid(%d) is null", l.Epid)
				return false
			}
			c.Base.from(op.Plat, op.Build, strconv.Itoa(int(s.GetSeason().SeasonId)), strconv.Itoa(int(l.Epid)), s.Cover, s.Season.Title, model.GotoPGC, strconv.Itoa(int(s.GetSeason().SeasonId)), model.ParamHandler(c.Materials.Prune, l.Epid, op.Rid, op.Entrance, op.FollowType, op.KeyWord))
			for _, p := range a.Pages {
				if p.Cid == l.Cid {
					c.HistoryArgs.Duration = p.Duration
					c.HistoryArgs.Progress = p.Duration
					break
				}
			}
		default:
			log.Warn("SmallCoverV3 From: unexpected goto %s", op.CardGoto)
			return false
		}
		if l.Pro > -1 {
			// l.pro == -1表示已看完
			c.HistoryArgs.Progress = l.Pro
		}
		if c.HistoryArgs.Duration > 0 {
			c.CoverRightText = model.HisDurationString(c.HistoryArgs.Progress) + " / " + model.HisDurationString(int64(c.HistoryArgs.Duration))
		}
	default:
		log.Warn("SmallCoverV3 From: unexpected type %T", main)
		return false
	}
	return true
}

func (c *SmallCoverV3) Get() *Base {
	return c.Base
}

type VerticalCoverV1 struct {
	*Base
	Desc            string       `json:"desc,omitempty"`
	CoverBadgeStyle *ReasonStyle `json:"cover_bage_style,omitempty"`
	SeasonID        int32        `json:"season_id,omitempty"`
}

func (c *VerticalCoverV1) From(main interface{}, op *operate.Card) bool {
	if op == nil {
		return false
	}
	switch main := main.(type) {
	case map[int32]*seasongrpc.CardInfoProto:
		scm := main
		sc, ok := scm[int32(op.ID)]
		if !ok || sc.FirstEpInfo == nil {
			log.Warn("VerticalCoverV1 From: seasonId(%d) is null", op.ID)
			return false
		}
		c.Base.from(op.Plat, op.Build, strconv.Itoa(int(sc.SeasonId)), strconv.Itoa(int(sc.FirstEpInfo.Id)), sc.Cover, sc.Title, model.GotoPGC, strconv.Itoa(int(sc.SeasonId)), model.ParamHandler(c.Materials.Prune, int64(sc.FirstEpInfo.Id), op.Rid, op.Entrance, op.FollowType, op.KeyWord))
		c.CoverBadgeStyle = reasonStyleFrom(model.PGCBageType[sc.BadgeType], sc.Badge)
		if sc.Stat != nil {
			c.Desc = model.BangumiFavString(int32(sc.Stat.Follow), sc.SeasonType)
		}
		c.SeasonID = sc.SeasonId
	case map[int32]*episodegrpc.EpisodeCardsProto:
		sm := main
		s, ok := sm[int32(op.ID)]
		if !ok || s.Season == nil {
			log.Warn("VerticalCoverV1 From: epid(%d) is null", op.ID)
			return false
		}
		c.Base.from(op.Plat, op.Build, strconv.Itoa(int(s.Season.SeasonId)), strconv.Itoa(int(s.EpisodeId)), s.Season.Cover, s.Season.Title, model.GotoPGC, strconv.Itoa(int(s.Season.SeasonId)), model.ParamHandler(c.Materials.Prune, int64(s.EpisodeId), op.Rid, op.Entrance, op.FollowType, op.KeyWord))
		c.CoverBadgeStyle = reasonStyleFrom(model.PGCBageType[s.Season.BadgeType], s.Season.Badge)
		if s.Season.Stat != nil {
			c.Desc = model.BangumiFavString(int32(s.Season.Stat.Follow), s.Season.SeasonType)
		}
		c.SeasonID = s.Season.SeasonId
	default:
		log.Warn("VerticalCoverV1 From: unexpected type %T", main)
		return false
	}
	return true
}

func (c *VerticalCoverV1) Get() *Base {
	return c.Base
}

type SmallCoverV4 struct {
	*Base
	CoverLeftText1  string       `json:"cover_left_text_1,omitempty"`
	CoverLeftIcon1  model.Icon   `json:"cover_left_icon_1,omitempty"`
	CoverRightText  string       `json:"cover_right_text,omitempty"`
	Desc            string       `json:"desc,omitempty"`
	DescIcon        model.Icon   `json:"desc_icon,omitempty"`
	CoverBadgeStyle *ReasonStyle `json:"cover_bage_style,omitempty"`
	ViewAt          int64        `json:"view_at,omitempty"`
	HistoryArgs     *HistoryArgs `json:"history_args,omitempty"`
	Owner           *Owner       `json:"owner,omitempty"`
	Videos          int64        `json:"videos,omitempty"`
}

func (c *SmallCoverV4) From(main interface{}, op *operate.Card) bool {
	if op == nil {
		return false
	}
	switch main := main.(type) {
	case map[int64]*arcgrpc.ViewReply:
		am := main
		a, ok := am[op.ID]
		if !ok {
			log.Warn("SmallCoverV4 From: aid(%d) is null", op.ID)
			return false
		}
		// mid > int32直接抛弃这张卡
		if model.CheckMidMaxInt32(a.Author.Mid, op.Build) {
			return false
		}
		// 过滤原因：互动视频
		if a.AttrVal(arcgrpc.AttrBitSteinsGate) == arcgrpc.AttrYes {
			c.Filter = model.FilterAttrBitSteinsGate
		}
		// 默认给首个cid
		cid := a.FirstCid
		// 获取历史记录里面的cid
		if op.Cid != 0 {
			cid = op.Cid
		}
		c.Base.from(op.Plat, op.Build, strconv.FormatInt(a.Aid, 10), strconv.FormatInt(cid, 10), a.Pic, a.Title, model.GotoAv, strconv.FormatInt(a.Aid, 10), model.ParamHandler(c.Materials.Prune, cid, op.Rid, op.Entrance, op.FollowType, op.KeyWord))
		c.CoverLeftText1 = model.StatString(a.Stat.View, "")
		c.CoverLeftIcon1 = model.IconPlay
		c.CoverRightText = model.DurationString(a.Duration)
		c.Desc = a.Author.Name
		c.DescIcon = model.IconUp
		c.ViewAt = op.ViewAt
		if op.Duration != 0 && op.Progress != 0 {
			c.HistoryArgs = &HistoryArgs{
				Duration: op.Duration,
				Progress: op.Progress,
			}
		}
		c.Owner = &Owner{
			Mid:  a.Arc.Author.Mid,
			Name: a.Arc.Author.Name,
			Face: a.Arc.Author.Face,
		}
		c.Videos = a.Videos
	case map[int64]*arcgrpc.Arc:
		am := main
		a, ok := am[op.ID]
		if !ok {
			log.Warn("SmallCoverV4 From: aid(%d) is null", op.ID)
			return false
		}
		// mid > int32直接抛弃这张卡
		if model.CheckMidMaxInt32(a.Author.Mid, op.Build) {
			return false
		}
		// 过滤原因：互动视频
		if a.AttrVal(arcgrpc.AttrBitSteinsGate) == arcgrpc.AttrYes {
			c.Filter = model.FilterAttrBitSteinsGate
		}
		c.Base.from(op.Plat, op.Build, strconv.FormatInt(a.Aid, 10), strconv.FormatInt(a.FirstCid, 10), a.Pic, a.Title, model.GotoAv, strconv.FormatInt(a.Aid, 10), model.ParamHandler(c.Materials.Prune, a.FirstCid, op.Rid, op.Entrance, op.FollowType, op.KeyWord))
		c.CoverLeftText1 = model.StatString(a.Stat.View, "")
		c.CoverLeftIcon1 = model.IconPlay
		c.CoverRightText = model.DurationString(a.Duration)
		c.Desc = a.Author.Name
		// 空间页展示投稿时间
		if op.Entrance == model.EntranceSpace {
			c.Desc = model.PubDataString(a.PubDate.Time())
		} else {
			c.DescIcon = model.IconUp
		}
		c.Owner = &Owner{
			Mid:  a.Author.Mid,
			Name: a.Author.Name,
			Face: a.Author.Face,
		}
		c.Videos = a.Videos
	case map[int64]*arcgrpc.ArcPlayer:
		am := main
		a, ok := am[op.ID]
		if !ok {
			log.Warn("SmallCoverV4 From: aid(%d) is null", op.ID)
			return false
		}
		// mid > int32直接抛弃这张卡
		if model.CheckMidMaxInt32(a.Arc.Author.Mid, op.Build) {
			return false
		}
		// 过滤原因：互动视频
		if a.Arc.AttrVal(arcgrpc.AttrBitSteinsGate) == arcgrpc.AttrYes {
			c.Filter = model.FilterAttrBitSteinsGate
		}
		// 默认给首个cid
		cid := a.Arc.FirstCid
		// 获取历史记录里面的cid
		if op.Cid != 0 {
			cid = op.Cid
		}
		c.Base.from(op.Plat, op.Build, strconv.FormatInt(a.Arc.Aid, 10), strconv.FormatInt(a.Arc.FirstCid, 10), a.Arc.Pic, a.Arc.Title, model.GotoAv, strconv.FormatInt(a.Arc.Aid, 10), model.ParamHandler(c.Materials.Prune, cid, op.Rid, op.Entrance, op.FollowType, op.KeyWord))
		c.URI = model.FillURI("", op.Plat, op.Build, c.URI, model.ArcPlayHandler(a.PlayerInfo[a.Arc.FirstCid]))
		c.CoverLeftText1 = model.StatString(a.Arc.Stat.View, "")
		c.CoverLeftIcon1 = model.IconPlay
		c.CoverRightText = model.DurationString(a.Arc.Duration)
		c.Desc = a.Arc.Author.Name
		// 空间页展示投稿时间
		if op.Entrance == model.EntranceSpace {
			c.Desc = model.PubDataString(a.Arc.PubDate.Time())
		} else {
			c.DescIcon = model.IconUp
		}
		c.Owner = &Owner{
			Mid:  a.Arc.Author.Mid,
			Name: a.Arc.Author.Name,
			Face: a.Arc.Author.Face,
		}
		c.Videos = a.Arc.Videos
	case map[int32]*seasongrpc.CardInfoProto:
		scm := main
		sc, ok := scm[int32(op.ID)]
		if !ok || sc.FirstEpInfo == nil {
			log.Warn("SmallCoverV4 From: seasonId(%d) or ep is null", op.ID)
			return false
		}
		cid := int64(sc.FirstEpInfo.Id)
		if op.Cid > 0 {
			cid = op.Cid
		}
		c.Base.from(op.Plat, op.Build, strconv.Itoa(int(sc.SeasonId)), strconv.FormatInt(cid, 10), sc.FirstEpInfo.Cover, sc.Title, model.GotoPGC, strconv.Itoa(int(sc.SeasonId)), model.ParamHandler(c.Materials.Prune, cid, op.Rid, op.Entrance, op.FollowType, op.KeyWord))
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
			log.Warn("SmallCoverV4 From: epid(%d) is null", op.ID)
			return false
		}
		c.Base.from(op.Plat, op.Build, strconv.Itoa(int(s.GetSeason().SeasonId)), strconv.Itoa(int(s.EpisodeId)), s.Cover, s.Season.Title, model.GotoPGC, strconv.Itoa(int(s.GetSeason().SeasonId)), model.ParamHandler(c.Materials.Prune, int64(s.EpisodeId), op.Rid, op.Entrance, op.FollowType, op.KeyWord))
		if s.Season.Stat != nil {
			c.CoverLeftText1 = model.StatString(int32(s.Season.Stat.View), "")
			c.CoverLeftIcon1 = model.IconPlay
		}
		c.CoverBadgeStyle = reasonStyleFrom(model.PGCBageType[s.Season.BadgeType], s.Season.Badge)
		c.Desc = s.Season.NewEpShow
		c.ViewAt = op.ViewAt
		if epInline, ok := c.Materials.EpInlinem[op.Epid]; ok {
			c.URI = model.FillURI("", op.Plat, op.Build, c.URI, model.PGCPlayHandler(epInline))
		}
		if op.Duration != 0 && op.Progress != 0 {
			c.HistoryArgs = &HistoryArgs{
				Duration: op.Duration,
				Progress: op.Progress,
			}
		}
	default:
		log.Warn("SmallCoverV4 From: unexpected type %T", main)
		return false
	}
	return true
}

func (c *SmallCoverV4) Get() *Base {
	return c.Base
}

type BannerV1 struct {
	*Base
	CoverLeftText1 string `json:"cover_left_text_1,omitempty"`
	CoverLeftText2 string `json:"cover_left_text_2,omitempty"`
	CoverLeftText3 string `json:"cover_left_text_3,omitempty"`
	Owner          *Owner `json:"owner,omitempty"`
}

type Owner struct {
	Mid  int64  `json:"mid,omitempty"`
	Name string `json:"name,omitempty"`
	Face string `json:"face,omitempty"`
	URI  string `json:"uri,omitempty"`
}

func (c *BannerV1) Get() *Base {
	return c.Base
}

func (c *BannerV1) From(main interface{}, op *operate.Card) bool {
	if op == nil {
		return false
	}
	switch main := main.(type) {
	case map[int64]*arcgrpc.ArcPlayer:
		am := main
		a, ok := am[op.ID]
		if !ok {
			log.Warn("BannerV1 From: aid(%d) is null", op.ID)
			return false
		}
		// mid > int32直接抛弃这张卡
		if model.CheckMidMaxInt32(a.Arc.Author.Mid, op.Build) {
			return false
		}
		// 过滤原因：互动视频
		if a.Arc.AttrVal(arcgrpc.AttrBitSteinsGate) == arcgrpc.AttrYes {
			c.Filter = model.FilterAttrBitSteinsGate
		}
		c.Base.from(op.Plat, op.Build, strconv.FormatInt(a.Arc.Aid, 10), strconv.FormatInt(a.Arc.FirstCid, 10), a.Arc.Pic, a.Arc.Title, model.GotoAv, strconv.FormatInt(a.Arc.Aid, 10), model.ParamHandler(c.Materials.Prune, a.Arc.FirstCid, op.Rid, op.Entrance, op.FollowType, op.KeyWord))
		c.URI = model.FillURI("", op.Plat, op.Build, c.URI, model.ArcPlayHandler(a.PlayerInfo[a.Arc.FirstCid]))
		c.CoverLeftText1 = model.DurationString(a.Arc.Duration)
		c.CoverLeftText2 = model.StatString(a.Arc.Stat.View, "播放")
		c.CoverLeftText3 = model.StatString(a.Arc.Stat.Danmaku, "弹幕")
		c.Owner = &Owner{
			Mid:  a.Arc.Author.Mid,
			Name: a.Arc.Author.Name,
			Face: a.Arc.Author.Face,
			URI:  model.FillURI(model.GotoSpace, op.Plat, op.Build, strconv.FormatInt(a.Arc.Author.Mid, 10), nil),
		}
		c.Cid = strconv.FormatInt(a.Arc.FirstCid, 10)
	case map[int32]*episodegrpc.EpisodeCardsProto:
		epm := main
		ep, ok := epm[int32(op.ID)]
		if !ok {
			log.Warn("BannerV1 From: epid(%d) is null", op.ID)
			return false
		}
		epInline, ok := c.Materials.EpInlinem[op.Epid]
		if !ok {
			log.Warn("BannerV1 ep_inline From: epid(%d) is null", op.Epid)
			return false
		}
		c.Base.from(op.Plat, op.Build, strconv.Itoa(int(ep.Season.SeasonId)), strconv.Itoa(int(ep.EpisodeId)), ep.Cover, ep.Season.Title, model.GotoPGC, strconv.Itoa(int(ep.Season.SeasonId)), model.ParamHandler(c.Materials.Prune, int64(ep.EpisodeId), op.Rid, op.Entrance, op.FollowType, op.KeyWord))
		c.URI = model.FillURI("", op.Plat, op.Build, c.URI, model.PGCPlayHandler(epInline))
		c.CoverLeftText1 = ep.Season.NewEpShow
		c.Cid = strconv.Itoa(int(ep.EpisodeId))
	default:
		log.Warn("BannerV1 From: unexpected type %T", main)
		return false
	}
	return true
}

type FM struct {
	*Base
	Aid  int64  `json:"aid,omitempty"`
	Cid  int64  `json:"cid,omitempty"`
	Desc string `json:"desc,omitempty"`
}

func (c *FM) Get() *Base {
	return c.Base
}

func (c *FM) From(main interface{}, op *operate.Card) bool {
	if op == nil {
		return false
	}
	switch main := main.(type) {
	case map[int64]*arcgrpc.Arc:
		am := main
		a, ok := am[op.ID]
		if !ok {
			log.Warn("FM From: aid(%d) is null", op.ID)
			return false
		}
		// mid > int32直接抛弃这张卡
		if model.CheckMidMaxInt32(a.Author.Mid, op.Build) {
			return false
		}
		// 过滤原因：互动视频
		if a.AttrVal(arcgrpc.AttrBitSteinsGate) == arcgrpc.AttrYes {
			c.Filter = model.FilterAttrBitSteinsGate
		}
		// 过滤掉PGC视频
		if a.AttrVal(arcgrpc.AttrBitIsPGC) == arcgrpc.AttrYes {
			return false
		}
		c.Title = a.Title
		c.Cover = a.Pic
		c.Goto = model.GotoAv
		c.Aid = a.Aid
		c.Cid = a.FirstCid
		c.Desc = a.Author.Name
	default:
		log.Warn("FM From: unexpected type %T", main)
		return false
	}
	return true
}
