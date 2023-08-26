package jsoncommon

import (
	"strconv"

	appcardmodel "go-gateway/app/app-svr/app-card/interface/model"
	cardschema "go-gateway/app/app-svr/app-feed/interface-ng/card-schema"
	jsoncard "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json"
	arcgrpc "go-gateway/app/app-svr/archive/service/api"

	taggrpc "git.bilibili.co/bapis/bapis-go/community/interface/tag"
)

type ArchiveCommon struct{}

func (ArchiveCommon) ConstructArchiveURI(aid int64, extraFn func(string) string) string {
	return appcardmodel.FillURI(appcardmodel.GotoAv, 0, 0, strconv.FormatInt(aid, 10), extraFn)
}

func (ArchiveCommon) ConstructPGCRedirectURI(redirectURL string, extraFn func(string) string) string {
	return appcardmodel.FillURI("", 0, 0, redirectURL, extraFn)
}

func (ArchiveCommon) ConstructVerticalArchiveURI(aid int64, device cardschema.Device, extraFn func(string) string) string {
	return appcardmodel.FillURI(appcardmodel.GotoVerticalAv, device.Plat(),
		int(device.Build()), strconv.FormatInt(aid, 10), extraFn)
}

func (ArchiveCommon) ConstructPlayerArgs(in *arcgrpc.ArcPlayer) *jsoncard.PlayerArgs {
	if in == nil {
		return nil
	}
	if in.Arc.AttrVal(arcgrpc.AttrBitIsPGC) == arcgrpc.AttrNo && in.Arc.Rights.Autoplay != 1 {
		return nil
	}
	if in.Arc.AttrVal(arcgrpc.AttrBitIsPGC) == arcgrpc.AttrYes && in.Arc.AttrVal(arcgrpc.AttrBitBadgepay) == arcgrpc.AttrYes {
		return nil
	}
	return &jsoncard.PlayerArgs{
		Aid:      in.Arc.Aid,
		Cid:      in.DefaultPlayerCid,
		Type:     appcardmodel.GotoAv,
		Duration: in.Arc.Duration,
	}
}

func (ArchiveCommon) ConstructDescButtonFromChannel(channelName string, channelID int64) *jsoncard.Button {
	out := &jsoncard.Button{
		Type:    appcardmodel.ButtonGrey,
		Text:    channelName,
		URI:     appcardmodel.FillURI(appcardmodel.GotoChannel, 0, 0, strconv.FormatInt(channelID, 10), nil),
		Event:   appcardmodel.EventChannelClick,
		EventV2: appcardmodel.EventV2ChannelClick,
	}
	return out
}

func (ArchiveCommon) ConstructDescButtonFromTag(in *taggrpc.Tag) *jsoncard.Button {
	return ConstructDescButtonFromTag(in)
}

func (ArchiveCommon) ConstructDescButtonFromArchvieType(in string) *jsoncard.Button {
	out := &jsoncard.Button{
		Text:    in,
		Event:   appcardmodel.EventChannelClick,
		EventV2: appcardmodel.EventV2ChannelClick,
		Type:    appcardmodel.ButtonGrey,
	}
	return out
}

func (ArchiveCommon) ConstructDescButtonFromDesc(in string) *jsoncard.Button {
	out := &jsoncard.Button{
		Text:    in,
		Event:   appcardmodel.EventChannelClick,
		EventV2: appcardmodel.EventV2ChannelClick,
		Type:    appcardmodel.ButtonGrey,
	}
	return out
}

func (ArchiveCommon) ConstructDescButtonFromMid(ctx cardschema.FeedContext, mid int64) *jsoncard.Button {
	out := &jsoncard.Button{
		Text:     "+ 关注",
		Param:    strconv.FormatInt(mid, 10),
		Event:    appcardmodel.EventUpFollow,
		Selected: 0,
		EventV2:  appcardmodel.EventV2UpFollow,
		Type:     appcardmodel.ButtonTheme,
	}
	if ctx.IsAttentionTo(mid) {
		out.Selected = 1
	}
	return out
}

func (ArchiveCommon) ConstructArgs(arcPlayer *arcgrpc.ArcPlayer, tag *taggrpc.Tag) jsoncard.Args {
	out := jsoncard.Args{}
	if arcPlayer != nil {
		out.Aid = arcPlayer.Arc.Aid
		out.UpID = arcPlayer.Arc.Author.Mid
		out.UpName = arcPlayer.Arc.Author.Name
		out.Rid = arcPlayer.Arc.TypeID
		out.Rname = arcPlayer.Arc.TypeName
	}
	if tag != nil {
		out.Tid = tag.Id
		out.Tname = tag.Name
	}
	return out
}
