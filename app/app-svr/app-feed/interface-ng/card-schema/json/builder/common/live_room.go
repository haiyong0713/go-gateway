package jsoncommon

import (
	"fmt"
	"strconv"

	"go-gateway/app/app-svr/app-card/interface/model"
	appcardmodel "go-gateway/app/app-svr/app-card/interface/model"
	"go-gateway/app/app-svr/app-card/interface/model/card/live"
	"go-gateway/app/app-svr/app-card/interface/model/card/operate"
	cardschema "go-gateway/app/app-svr/app-feed/interface-ng/card-schema"
	jsoncard "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json"

	accountgrpc "git.bilibili.co/bapis/bapis-go/account/service"
	livexroomgate "git.bilibili.co/bapis/bapis-go/live/xroom-gate"
)

type LiveRoomCommon struct{}

func (LiveRoomCommon) ConstructThreePointFromLiveRoom(in *live.Room) *jsoncard.ThreePoint {
	return &jsoncard.ThreePoint{DislikeReasons: constructThreePointFromLiveRoom(in)}
}

func constructThreePointFromLiveRoom(in *live.Room) []*jsoncard.DislikeReason {
	dislikeReasons := []*jsoncard.DislikeReason{}
	if in.Uname != "" {
		dislikeReasons = append(dislikeReasons, &jsoncard.DislikeReason{
			ID:    _upper,
			Name:  fmt.Sprintf("UP主:%s", in.Uname),
			Toast: _dislikeToast,
		})
	}
	if in.AreaV2Name != "" {
		dislikeReasons = append(dislikeReasons, &jsoncard.DislikeReason{
			ID:    _region,
			Name:  fmt.Sprintf("分区:%s", in.AreaV2Name),
			Toast: _dislikeToast,
		})
	}
	dislikeReasons = append(dislikeReasons, &jsoncard.DislikeReason{
		ID:    _noSeason,
		Name:  "不感兴趣",
		Toast: _dislikeToast,
	})
	return dislikeReasons
}

func (LiveRoomCommon) ConstructThreePointV2FromLiveRoom(ctx cardschema.FeedContext, in *live.Room) []*jsoncard.ThreePointV2 {
	out := []*jsoncard.ThreePointV2{}
	dislikeSubTitle, _, dislikeTitle := dislikeText(ctx)
	reason := constructThreePointFromLiveRoom(in)
	replaceDislikeReason(ctx, reason, &threePointConfig{})
	out = append(out, &jsoncard.ThreePointV2{
		Title:    dislikeTitle,
		Subtitle: dislikeSubTitle,
		Reasons:  reason,
		Type:     model.ThreePointDislike,
	})
	return out
}

func (LiveRoomCommon) ConstructDescButtonFromLiveRoom(in *live.Room) *jsoncard.Button {
	return &jsoncard.Button{
		Type:    appcardmodel.ButtonGrey,
		Text:    in.AreaV2Name,
		URI:     model.FillURI(model.GotoLiveTag, 0, 0, strconv.FormatInt(in.AreaV2ParentID, 10), appcardmodel.LiveRoomTagHandler(in)),
		Event:   appcardmodel.EventChannelClick,
		EventV2: appcardmodel.EventV2ChannelClick,
	}
}

func (LiveRoomCommon) ConstructPlayerArgsFromLiveRoom(in *live.Room, contentMode int64) *jsoncard.PlayerArgs {
	return &jsoncard.PlayerArgs{
		RoomID:      in.RoomID,
		IsLive:      1,
		Type:        appcardmodel.GotoLive,
		ContentMode: contentMode,
	}
}

func (LiveRoomCommon) ConstructArgsFromLiveRoom(in *live.Room) jsoncard.Args {
	out := jsoncard.Args{}
	out.UpID = in.UID
	out.UpName = in.Uname
	out.Rid = int32(in.AreaV2ParentID)
	out.Rname = in.AreaV2ParentName
	out.Tid = in.AreaV2ID
	out.Tname = in.AreaV2Name
	out.RoomID = in.RoomID
	out.Online = in.Online
	return out
}

func (LiveRoomCommon) ConstructCoverLeftMeta(in *live.Room) (string, appcardmodel.Icon, bool) {
	if in.WatchedShow == nil {
		return "", 0, false
	}
	icon := appcardmodel.IconLiveOnline
	if in.WatchedShow.Switch {
		icon = appcardmodel.IconLiveWatched
	}
	return appcardmodel.Stat64String(in.WatchedShow.Num, ""), icon, true
}

type LiveEntryRoomCommon struct{}

func (LiveEntryRoomCommon) ConstructThreePointFromLiveEntryRoom(in *livexroomgate.EntryRoomInfoResp_EntryList, author *accountgrpc.Card) *jsoncard.ThreePoint {
	return &jsoncard.ThreePoint{DislikeReasons: constructThreePointFromLiveEntryRoom(in, author)}
}

func constructThreePointFromLiveEntryRoom(in *livexroomgate.EntryRoomInfoResp_EntryList, author *accountgrpc.Card) []*jsoncard.DislikeReason {
	dislikeReasons := []*jsoncard.DislikeReason{}
	if author.Name != "" {
		dislikeReasons = append(dislikeReasons, &jsoncard.DislikeReason{
			ID:    _upper,
			Name:  fmt.Sprintf("UP主:%s", author.Name),
			Toast: "将减少相似内容推荐",
		})
	}
	if in.AreaName != "" {
		dislikeReasons = append(dislikeReasons, &jsoncard.DislikeReason{
			ID:    _region,
			Name:  fmt.Sprintf("分区:%s", in.AreaName),
			Toast: _dislikeToast,
		})
	}
	dislikeReasons = append(dislikeReasons, &jsoncard.DislikeReason{
		ID:    _noSeason,
		Name:  "不感兴趣",
		Toast: _dislikeToast,
	})
	return dislikeReasons
}

func (LiveEntryRoomCommon) ConstructThreePointV2FromLiveEntryRoom(in *livexroomgate.EntryRoomInfoResp_EntryList, author *accountgrpc.Card) []*jsoncard.ThreePointV2 {
	out := []*jsoncard.ThreePointV2{}
	out = append(out, &jsoncard.ThreePointV2{
		Title:    "不感兴趣",
		Subtitle: "(选择后将减少相似内容推荐)",
		Reasons:  constructThreePointFromLiveEntryRoom(in, author),
		Type:     model.ThreePointDislike,
	})
	return out
}

func (LiveEntryRoomCommon) ConstructCoverLeftMeta(in *livexroomgate.EntryRoomInfoResp_EntryList) (string, appcardmodel.Icon, bool) {
	if in.WatchedShow == nil {
		return "", 0, false
	}
	if in.WatchedShow.Switch {
		return in.WatchedShow.TextSmall, appcardmodel.IconLiveWatched, true
	}
	return in.WatchedShow.TextSmall, appcardmodel.IconLiveOnline, true
}

func (LiveEntryRoomCommon) ConstructDescButtonFromLiveEntryRoom(in *livexroomgate.EntryRoomInfoResp_EntryList) *jsoncard.Button {
	return &jsoncard.Button{
		Type:    appcardmodel.ButtonGrey,
		Text:    in.AreaName,
		URI:     model.FillURI(model.GotoLiveTag, 0, 0, strconv.FormatInt(0, 10), appcardmodel.LiveEntryRoomTagHandler(in)),
		Event:   appcardmodel.EventChannelClick,
		EventV2: appcardmodel.EventV2ChannelClick,
	}
}

func (LiveEntryRoomCommon) ConstructPlayerArgsFromLiveEntryRoom(in *livexroomgate.EntryRoomInfoResp_EntryList) *jsoncard.PlayerArgs {
	return &jsoncard.PlayerArgs{
		RoomID: in.RoomId,
		IsLive: 1,
		Type:   appcardmodel.GotoLive,
	}
}

func (LiveEntryRoomCommon) ConstructArgsFromLiveEntryRoom(in *livexroomgate.EntryRoomInfoResp_EntryList, author *accountgrpc.Card) jsoncard.Args {
	out := jsoncard.Args{}
	out.UpID = in.Uid
	out.UpName = author.Name
	out.Rid = int32(in.ParentAreaId)
	out.Rname = in.ParentAreaName
	out.Tid = in.AreaId
	out.Tname = in.AreaName
	out.RoomID = in.RoomId
	out.Online = int32(in.PopularityCount)
	return out
}

func (LiveRoomCommon) ConstructLeftBottomRcmdReasonStyle(in *operate.LiveBottomBadge) *jsoncard.ReasonStyle {
	out := &jsoncard.ReasonStyle{
		Text:             in.Text,
		TextColor:        in.TextColor,
		BgColor:          in.BgColor,
		BorderColor:      in.BorderColor,
		IconURL:          in.IconURL,
		TextColorNight:   in.TextColorNight,
		BgColorNight:     in.BgColorNight,
		BorderColorNight: in.BorderColorNight,
		BgStyle:          in.BgStyle,
	}
	return out
}
