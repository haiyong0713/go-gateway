package search

import (
	"fmt"
	"go-gateway/app/app-svr/app-search/configs"

	"go-common/library/xstr"

	"go-gateway/app/app-svr/app-search/internal/model"

	pgcsearch "git.bilibili.co/bapis/bapis-go/pgc/service/card/search/v1"
)

// BuildPgcReq builds the pgc request
func (m *Media) BuildPgcReq() (sepReq *pgcsearch.SeasonEpReq) {
	sepReq = &pgcsearch.SeasonEpReq{
		SeasonId: int32(m.SeasonID),
		MediaId:  int32(m.MediaID),
		Type:     int32(m.MediaType),
	}
	if m.HitEpids != "" { // 541 搜索新增命中单集
		if hitepids, err := xstr.SplitInts(m.HitEpids); err == nil {
			for _, v := range hitepids {
				sepReq.EpIds = append(sepReq.EpIds, int32(v))
			}
		}
	}
	return
}

// FromPgcEp def
func (v *Item) FromPgcEp(ep *pgcsearch.SearchEpProto, cfg *configs.PgcSearchCard) {
	v.URI = ep.Url
	v.Param = fmt.Sprintf("%d", ep.Id)
	v.Cover = ep.Cover
	v.Title = ep.Title
	if ep.ReleaseDate != "" {
		v.Label = fmt.Sprintf(cfg.EpLabel, ep.ReleaseDate)
	}
	if len(ep.Badges) == 0 {
		return
	}
	for _, bdg := range ep.Badges {
		v.Badges = append(v.Badges, &model.ReasonStyle{
			Text:             bdg.Text,
			TextColor:        bdg.TextColor,
			TextColorNight:   bdg.TextColorNight,
			BgColor:          bdg.BgColor,
			BgColorNight:     bdg.BgColorNight,
			BorderColor:      bdg.BorderColor,
			BorderColorNight: bdg.BorderColorNight,
			BgStyle:          int8(bdg.BgStyle),
		})
	}
}

// EpisodesNewReq def.
type EpisodesNewReq struct {
	Pn       int32 `form:"pn" default:"1"`
	Ps       int32 `form:"ps" default:"20"`
	SeasonId int32 `form:"season_id" validate:"required"`
}
