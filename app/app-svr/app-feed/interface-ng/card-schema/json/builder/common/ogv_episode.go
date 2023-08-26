package jsoncommon

import (
	"fmt"
	"strconv"

	appcardmodel "go-gateway/app/app-svr/app-card/interface/model"
	"go-gateway/app/app-svr/app-card/interface/model/card/ai"
	cardschema "go-gateway/app/app-svr/app-feed/interface-ng/card-schema"

	pgccard "git.bilibili.co/bapis/bapis-go/pgc/service/card"
)

type OgvEpisode struct{}

func (OgvEpisode) ConstructOgvURI(in *pgccard.EpisodeCard, device cardschema.Device, rcmd *ai.Item) string {
	param := in.Url
	if param == "" {
		plat := device.Plat()
		build := int(device.Build())
		param = appcardmodel.FillURI(appcardmodel.GotoBangumi, plat, build, strconv.FormatInt(int64(in.EpisodeId), 10), nil)
	}
	return appcardmodel.FillURI("", 0, 0, param, appcardmodel.PGCTrackIDHandler(rcmd))
}

func (OgvEpisode) ConstructEpCover(in *pgccard.EpisodeCard) *EpCover {
	if in.Season.Stat == nil {
		return nil
	}
	epCover := &EpCover{
		CoverLeftText1: appcardmodel.StatString(int32(in.Season.Stat.View), ""),
		CoverLeftIcon1: appcardmodel.IconPlay,
		CoverLeftText2: appcardmodel.StatString(int32(in.Season.Stat.Follow), ""),
		CoverLeftIcon2: appcardmodel.IconFavorite,
	}
	epCover.CoverLeft1ContentDescription = appcardmodel.CoverIconContentDescription(epCover.CoverLeftIcon1,
		epCover.CoverLeftText1)
	epCover.CoverLeft2ContentDescription = appcardmodel.CoverIconContentDescription(epCover.CoverLeftIcon2,
		epCover.CoverLeftText2)
	return epCover
}

func (e OgvEpisode) ConstructOgvRightText(episode *pgccard.EpisodeCard, hasScore bool, enableScoreVersion bool) string {
	if hasScore && episode.Season.RatingInfo != nil && episode.Season.RatingInfo.Score > 0 && enableScoreVersion {
		return fmt.Sprintf("%.1f分", episode.Season.RatingInfo.Score)
	}
	return ""
}
