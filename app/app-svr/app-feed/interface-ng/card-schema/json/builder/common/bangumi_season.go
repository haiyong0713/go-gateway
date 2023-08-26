package jsoncommon

import (
	"fmt"
	"strconv"

	appcardmodel "go-gateway/app/app-svr/app-card/interface/model"
	"go-gateway/app/app-svr/app-card/interface/model/card/ai"
	cardschema "go-gateway/app/app-svr/app-feed/interface-ng/card-schema"
	jsoncard "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json"
	jsonreasonstyle "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder/reason_style"

	taggrpc "git.bilibili.co/bapis/bapis-go/community/interface/tag"
	episodegrpc "git.bilibili.co/bapis/bapis-go/pgc/service/season/episode"
)

type BangumiSeason struct{}

type EpCover struct {
	CoverLeftText1               string
	CoverLeftIcon1               appcardmodel.Icon
	CoverLeft1ContentDescription string
	CoverLeftText2               string
	CoverLeftIcon2               appcardmodel.Icon
	CoverLeft2ContentDescription string
}

func (BangumiSeason) ConstructSeasonURI(episodeID string) string {
	return appcardmodel.FillURI(appcardmodel.GotoBangumi, 0, 0, episodeID, nil)
}

func (BangumiSeason) ConstructDescButtonFromTag(in *taggrpc.Tag) *jsoncard.Button {
	return ConstructDescButtonFromTag(in)
}

func (BangumiSeason) ConstructDescButtonFromNewEpShow(in string) *jsoncard.Button {
	out := &jsoncard.Button{
		Text:    in,
		Event:   appcardmodel.EventChannelClick,
		EventV2: appcardmodel.EventV2ChannelClick,
		Type:    appcardmodel.ButtonGrey,
	}
	return out
}

func (BangumiSeason) ConstructEpTitle(in *episodegrpc.EpisodeCardsProto) string {
	title := in.Season.Title
	if in.ShowTitle != "" {
		title = fmt.Sprintf("%s：%s", title, in.ShowTitle)
	}
	return title
}

func (BangumiSeason) ConstructEpParam(in *episodegrpc.EpisodeCardsProto) string {
	return strconv.FormatInt(int64(in.EpisodeId), 10)
}

func (BangumiSeason) ConstructEpURI(in *episodegrpc.EpisodeCardsProto, device cardschema.Device, rcmd *ai.Item) string {
	param := in.Url
	if param == "" {
		plat := device.Plat()
		build := int(device.Build())
		param = appcardmodel.FillURI(appcardmodel.GotoBangumi, plat, build, strconv.FormatInt(int64(in.EpisodeId), 10), nil)
	}
	return appcardmodel.FillURI("", 0, 0, param, appcardmodel.PGCTrackIDHandler(rcmd))
}

func (BangumiSeason) ConstructEpCover(in *episodegrpc.EpisodeCardsProto) *EpCover {
	if in.Season.Stat == nil {
		return nil
	}
	return &EpCover{
		CoverLeftText1: appcardmodel.StatString(int32(in.Season.Stat.View), ""),
		CoverLeftIcon1: appcardmodel.IconPlay,
		CoverLeftText2: appcardmodel.StatString(int32(in.Season.Stat.Follow), ""),
		CoverLeftIcon2: appcardmodel.IconFavorite,
	}
}

func (BangumiSeason) ConstructEpBadge(in *episodegrpc.EpisodeCardsProto) (string, *jsoncard.ReasonStyle) {
	return in.Season.SeasonTypeName, jsonreasonstyle.ConstructReasonStyle(appcardmodel.BgColorTransparentRed, in.Season.SeasonTypeName)
}
