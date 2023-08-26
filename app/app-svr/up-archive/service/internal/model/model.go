package model

import (
	"time"

	"go-gateway/app/app-svr/up-archive/service/api"

	arcmdl "git.bilibili.co/bapis/bapis-go/archive/service"
)

const _upArcScoreMaxRand = 512999

type UpArcPubMsg struct {
	Mid   int64 `json:"mid"`
	Ctime int64 `json:"ctime"`
}

// nolint:gomnd
func MaxScore() int64 {
	return (time.Now().Unix()+1)<<21 | _upArcScoreMaxRand<<2 | 3
}

func CopyFromArc(from *arcmdl.Arc) *api.Arc {
	to := &api.Arc{
		Aid:         from.Aid,
		Videos:      from.Videos,
		TypeID:      from.TypeID,
		TypeName:    from.TypeName,
		Copyright:   from.Copyright,
		Pic:         from.Pic,
		Title:       from.Title,
		PubDate:     from.PubDate,
		Ctime:       from.Ctime,
		Desc:        from.Desc,
		State:       from.State,
		Access:      from.Access,
		Attribute:   from.Attribute,
		Tag:         from.Tag,
		Tags:        from.Tags,
		Duration:    from.Duration,
		MissionID:   from.MissionID,
		OrderID:     from.OrderID,
		RedirectURL: from.RedirectURL,
		Forward:     from.Forward,
		Rights: api.Rights{
			Bp:              from.Rights.Bp,
			Elec:            from.Rights.Elec,
			Download:        from.Rights.Download,
			Movie:           from.Rights.Movie,
			Pay:             from.Rights.Pay,
			HD5:             from.Rights.HD5,
			NoReprint:       from.Rights.NoReprint,
			Autoplay:        from.Rights.Autoplay,
			UGCPay:          from.Rights.UGCPay,
			IsCooperation:   from.Rights.IsCooperation,
			UGCPayPreview:   from.Rights.UGCPayPreview,
			NoBackground:    from.Rights.NoBackground,
			ArcPay:          from.Rights.ArcPay,
			ArcPayFreeWatch: from.Rights.ArcPayFreeWatch,
		},
		Author: api.Author{
			Mid:  from.Author.Mid,
			Name: from.Author.Name,
			Face: from.Author.Face,
		},
		Stat: api.Stat{
			Aid:     from.Stat.Aid,
			View:    from.Stat.View,
			Danmaku: from.Stat.Danmaku,
			Reply:   from.Stat.Reply,
			Fav:     from.Stat.Fav,
			Coin:    from.Stat.Coin,
			Share:   from.Stat.Share,
			NowRank: from.Stat.NowRank,
			HisRank: from.Stat.HisRank,
			Like:    from.Stat.Like,
			DisLike: from.Stat.DisLike,
		},
		ReportResult: from.ReportResult,
		Dynamic:      from.Dynamic,
		FirstCid:     from.FirstCid,
		Dimension: api.Dimension{
			Width:  from.Dimension.Width,
			Height: from.Dimension.Height,
			Rotate: from.Dimension.Rotate,
		},
		SeasonID:    from.SeasonID,
		AttributeV2: from.AttributeV2,
		UpFrom:      from.GetUpFromV2(),
	}
	for _, v := range from.StaffInfo {
		if v == nil {
			continue
		}
		to.StaffInfo = append(to.StaffInfo, &api.StaffInfo{
			Mid:       v.Mid,
			Title:     v.Title,
			Attribute: v.Attribute,
		})
	}
	return to
}
