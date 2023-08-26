package rank

import (
	"strconv"

	accountgrpc "git.bilibili.co/bapis/bapis-go/account/service"
	accv1 "git.bilibili.co/bapis/bapis-go/account/service"
	serGRPC "git.bilibili.co/bapis/bapis-go/content-flow-control/service"
	cardm "go-gateway/app/app-svr/app-card/interface/model"
	rankmod "go-gateway/app/app-svr/app-show/interface/api/rank"
	"go-gateway/app/app-svr/app-show/interface/model"
	"go-gateway/app/app-svr/archive/service/api"

	relationgrpc "git.bilibili.co/bapis/bapis-go/account/service/relation"
)

type List struct {
	Aid    int64   `json:"aid"`
	Score  int64   `json:"score"`
	Others []*List `json:"others"`
}

type InnerAttr struct {
	OverSeaBlock bool
}

func FromArchiveRank(a *api.Arc, scores map[int64]int64) (i *rankmod.Item) {
	i = &rankmod.Item{
		Title:       a.Title,
		Cover:       a.Pic,
		Param:       strconv.FormatInt(a.Aid, 10),
		Uri:         model.FillURI(model.GotoAv, strconv.FormatInt(a.Aid, 10), model.AvHandler(a)),
		RedirectUrl: a.RedirectURL,
		Goto:        model.GotoAv,
		Play:        a.Stat.View,
		Danmaku:     a.Stat.Danmaku,
		Mid:         a.Author.Mid,
		Name:        a.Author.Name,
		Face:        a.Author.Face,
		Reply:       a.Stat.Reply,
		Favourite:   a.Stat.Fav,
		PubDate:     a.PubDate,
		Rid:         a.TypeID,
		Rname:       a.TypeName,
		Duration:    a.Duration,
		Like:        a.Stat.Like,
		Cid:         a.FirstCid,
	}
	if score, ok := scores[a.Aid]; ok {
		i.Pts = score
	}
	if a.Access > 0 {
		i.Play = 0
	}
	if a.Rights.IsCooperation > 0 {
		i.Cooperation = "等联合创作"
	}
	return
}

func FromOfficialVerify(a accv1.OfficialInfo) (i *rankmod.OfficialVerify) {
	i = &rankmod.OfficialVerify{}
	if a.Role == 0 {
		i.Type = -1
	} else {
		if a.Role <= 2 || a.Role == 7 {
			i.Type = 0
		} else {
			i.Type = 1
		}
		i.Desc = a.Title
	}
	return
}

func (l *List) FromRanks(as map[int64]*api.Arc, authorRelations map[int64]*relationgrpc.InterrelationReply, authorStats map[int64]*relationgrpc.StatReply, authorCards map[int64]*accountgrpc.Card, plat int8, innerAttr map[int64]*InnerAttr) (i *rankmod.Item) {
	i = fromRank(l, as, plat, innerAttr)
	if i == nil {
		return
	}
	for _, other := range l.Others {
		if other != nil {
			if child := fromRank(other, as, plat, innerAttr); child != nil {
				i.Children = append(i.Children, child)
			}
		}
	}
	i.Attribute = int32(cardm.RelationOldChange(i.Mid, authorRelations))
	if len(authorStats) > 0 {
		if stats, ok := authorStats[i.Mid]; ok {
			i.Follower = stats.Follower
		}
	}
	if len(authorCards) > 0 {
		if info, ok := authorCards[i.Mid]; ok {
			i.OfficialVerify = FromOfficialVerify(info.Official)
		}
	}
	rel := cardm.RelationChange(i.Mid, authorRelations)
	if rel != nil {
		i.Relation = &rankmod.Relation{
			Status:     rel.Status,
			IsFollow:   rel.IsFollow,
			IsFollowed: rel.IsFollowed,
		}
	}
	return
}

func ChangeInnerAttr(in []*serGRPC.InfoItem) *InnerAttr {
	out := &InnerAttr{}
	if in == nil {
		return out
	}
	for _, v := range in {
		switch v.Key {
		case "54": //oversea_block
			if v.Value == 1 {
				out.OverSeaBlock = true
			}
		default:
			continue
		}
	}
	return out
}

func fromRank(l *List, as map[int64]*api.Arc, plat int8, innerAttr map[int64]*InnerAttr) (i *rankmod.Item) {
	if a, ok := as[l.Aid]; ok {
		var overSeaFlag bool
		if inva, k := innerAttr[l.Aid]; k && inva != nil {
			overSeaFlag = inva.OverSeaBlock
		}
		if model.IsOverseas(plat) && overSeaFlag {
			return
		}
		i = &rankmod.Item{
			Title:       a.Title,
			Cover:       a.Pic,
			Param:       strconv.FormatInt(a.Aid, 10),
			Uri:         model.FillURI(model.GotoAv, strconv.FormatInt(a.Aid, 10), model.AvHandler(a)),
			RedirectUrl: a.RedirectURL,
			Goto:        model.GotoAv,
			Play:        a.Stat.View,
			Danmaku:     a.Stat.Danmaku,
			Mid:         a.Author.Mid,
			Name:        a.Author.Name,
			Face:        a.Author.Face,
			Reply:       a.Stat.Reply,
			Favourite:   a.Stat.Fav,
			PubDate:     a.PubDate,
			Rid:         a.TypeID,
			Rname:       a.TypeName,
			Duration:    a.Duration,
			Like:        a.Stat.Like,
			Cid:         a.FirstCid,
		}
		i.Pts = l.Score
		if a.Access > 0 {
			i.Play = 0
		}
		if a.Rights.IsCooperation > 0 {
			i.Cooperation = "等联合创作"
		}
		if !model.IsIPad(plat) {
			if a.RedirectURL != "" {
				i.Uri = a.RedirectURL
				i.Goto = model.GotoBangumi
			}
		}
	}
	return
}
