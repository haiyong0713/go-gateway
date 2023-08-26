package cm

import (
	"fmt"
	"strconv"

	"go-common/library/log"
	appcardmodel "go-gateway/app/app-svr/app-card/interface/model"
	"go-gateway/app/app-svr/app-card/interface/model/card"
	"go-gateway/app/app-svr/app-card/interface/model/card/ai"
	"go-gateway/app/app-svr/app-card/interface/model/card/cm"
	"go-gateway/app/app-svr/app-card/interface/model/card/live"
	cardschema "go-gateway/app/app-svr/app-feed/interface-ng/card-schema"
	jsoncard "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json"
	jsonbuilder "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder"
	jsonavatar "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder/avatar"
	jsoncommon "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder/common"
	jsonreasonstyle "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder/reason_style"

	accountgrpc "git.bilibili.co/bapis/bapis-go/account/service"

	"github.com/pkg/errors"
)

type V2AdInlineLiveBuilder interface {
	Parent() CmV2BuilderFactory
	SetBase(*jsoncard.Base) V2AdInlineLiveBuilder
	SetRcmd(*ai.Item) V2AdInlineLiveBuilder
	SetAdInfo(*cm.AdInfo) V2AdInlineLiveBuilder
	SetLiveRoom(*live.Room) V2AdInlineLiveBuilder
	SetAuthorCard(*accountgrpc.Card) V2AdInlineLiveBuilder

	Build() (*jsoncard.LargeCoverInline, error)
	WithAfter(req ...func(*jsoncard.LargeCoverInline)) V2AdInlineLiveBuilder
}

type v2AdInlineLiveBuilder struct {
	jsoncommon.LiveRoomCommon
	parent     *cmV2BuilderFactory
	base       *jsoncard.Base
	rcmd       *ai.Item
	adInfo     *cm.AdInfo
	liveRoom   *live.Room
	authorCard *accountgrpc.Card
	afterFn    []func(*jsoncard.LargeCoverInline)
}

func (b v2AdInlineLiveBuilder) Parent() CmV2BuilderFactory {
	return b.parent
}

func (b v2AdInlineLiveBuilder) SetBase(base *jsoncard.Base) V2AdInlineLiveBuilder {
	b.base = base
	return b
}

func (b v2AdInlineLiveBuilder) SetRcmd(in *ai.Item) V2AdInlineLiveBuilder {
	b.rcmd = in
	return b
}

func (b v2AdInlineLiveBuilder) SetAdInfo(in *cm.AdInfo) V2AdInlineLiveBuilder {
	b.adInfo = in
	return b
}

func (b v2AdInlineLiveBuilder) SetLiveRoom(in *live.Room) V2AdInlineLiveBuilder {
	b.liveRoom = in
	return b
}

func (b v2AdInlineLiveBuilder) SetAuthorCard(in *accountgrpc.Card) V2AdInlineLiveBuilder {
	b.authorCard = in
	return b
}

func (b v2AdInlineLiveBuilder) constructOfficialIcon() appcardmodel.Icon {
	return appcardmodel.OfficialIcon(b.authorCard)
}

func (b v2AdInlineLiveBuilder) constructURI() string {
	device := b.parent.BuilderContext.Device()
	uri := appcardmodel.FillURI(appcardmodel.GotoLive, device.Plat(), int(device.Build()), strconv.FormatInt(b.liveRoom.RoomID, 10), appcardmodel.LiveRoomHandler(b.liveRoom, device.Network()))
	if b.liveRoom.Link != "" {
		uri = appcardmodel.FillURI("", 0, 0, b.liveRoom.Link, appcardmodel.URLTrackIDHandler(b.rcmd))
	}
	return uri
}

func (b v2AdInlineLiveBuilder) constructArgs() jsoncard.Args {
	out := jsoncard.Args{}
	out.UpID = b.liveRoom.UID
	out.UpName = b.liveRoom.Uname
	out.Rid = int32(b.liveRoom.AreaV2ParentID)
	out.Rname = b.liveRoom.AreaV2ParentName
	out.Tid = b.liveRoom.AreaV2ID
	out.Tname = b.liveRoom.AreaV2Name
	out.RoomID = b.liveRoom.RoomID
	out.Online = b.liveRoom.Online
	out.IsFollow = 0
	if b.parent.BuilderContext.IsAttentionTo(b.liveRoom.UID) {
		out.IsFollow = 1
	}
	return out
}

func (b v2AdInlineLiveBuilder) constructPlayerArgs() *jsoncard.PlayerArgs {
	return &jsoncard.PlayerArgs{
		RoomID:         b.liveRoom.RoomID,
		IsLive:         1,
		Type:           appcardmodel.GotoLive,
		ManualPlay:     b.rcmd.ManualInline(),
		HidePlayButton: appcardmodel.HidePlayButton,
	}
}

func (b v2AdInlineLiveBuilder) constructThreePoint() *jsoncard.ThreePoint {
	if !b.parent.BuilderContext.FeatureGates().FeatureEnabled(cardschema.FeatureNewDislike) {
		return nil
	}
	return b.ConstructThreePointFromLiveRoom(b.liveRoom)
}

func (b v2AdInlineLiveBuilder) constructThreePointV2() []*jsoncard.ThreePointV2 {
	if !b.parent.BuilderContext.FeatureGates().FeatureEnabled(cardschema.FeatureNewDislike) {
		return nil
	}
	return b.ConstructThreePointV2FromLiveRoom(b.parent.BuilderContext, b.liveRoom)
}

func (b v2AdInlineLiveBuilder) Build() (*jsoncard.LargeCoverInline, error) {
	if b.base == nil {
		return nil, errors.Errorf("empty `base` field")
	}
	if b.rcmd == nil {
		return nil, errors.Errorf("empty `rcmd` field")
	}
	if b.liveRoom == nil {
		return nil, errors.Errorf("empty `liveRoom` field")
	}
	if card.CheckMidMaxInt32(b.liveRoom.UID) && b.parent.BuilderContext.VersionControl().Can("feed.disableInt64Mid") {
		return nil, errors.Errorf("ignore on maxint32 mid: %d", b.liveRoom.UID)
	}

	if err := jsonbuilder.NewBaseUpdater(b.parent.BuilderContext, b.base).
		UpdateCover(b.liveRoom.Cover).
		UpdateTitle(b.liveRoom.Title).
		UpdateURI(b.constructURI()).
		UpdateArgs(b.constructArgs()).
		UpdatePlayerArgs(b.constructPlayerArgs()).
		UpdateThreePoint(b.constructThreePoint()).
		UpdateThreePointV2(b.constructThreePointV2()).
		Update(); err != nil {
		return nil, err
	}
	out := &jsoncard.LargeCoverInline{
		OfficialIcon:      b.constructOfficialIcon(),
		OfficialIconV2:    b.constructOfficialIcon(),
		RightTopLiveBadge: card.ConstructRightTopLiveBadge(b.liveRoom.LiveStatus),
	}
	out.CanPlay = 1
	out.CoverLeftText1 = appcardmodel.StatString(b.liveRoom.Online, "")
	out.CoverLeftIcon1 = appcardmodel.IconOnline
	out.CoverLeftText2 = fmt.Sprintf("%s · %s", b.liveRoom.AreaV2ParentName, b.liveRoom.AreaV2Name)
	out.BadgeStyle = jsonreasonstyle.ConstructReasonStyle(appcardmodel.BgColorTransparentRed, "直播")
	if b.rcmd.RcmdReason != nil || b.parent.BuilderContext.IsAttentionTo(b.liveRoom.UID) {
		reasonText, _ := jsonreasonstyle.BuildRecommendReasonText(
			b.parent.BuilderContext,
			b.rcmd.RcmdReason,
			b.rcmd.Goto,
			"",
			b.parent.BuilderContext.IsAttentionTo(b.liveRoom.UID),
		)
		out.RcmdReasonStyle = jsonreasonstyle.ConstructTopReasonStyle(
			reasonText,
			jsonreasonstyle.CornerMarkFromAI(b.rcmd),
			jsonreasonstyle.CorverMarkFromContext(b.parent.BuilderContext),
		)
	}
	if b.parent.BuilderContext.IsAttentionTo(b.liveRoom.UID) {
		out.OfficialIcon = appcardmodel.IconIsAttenm
		out.IsAtten = true
	}
	avatar, err := jsonavatar.NewAvatarBuilder(b.parent.BuilderContext).
		SetAvatarStatus(&jsoncard.AvatarStatus{
			Cover: b.liveRoom.Face,
			Text:  b.liveRoom.Uname,
			Goto:  appcardmodel.GotoMid,
			Param: strconv.FormatInt(b.liveRoom.UID, 10),
			Type:  appcardmodel.AvatarRound,
		}).Build()
	if err != nil {
		log.Error("Failed to build avatar: %+v", err)
	}
	out.Avatar = avatar
	out.Base = b.base
	for _, fn := range b.afterFn {
		fn(out)
	}
	return out, nil
}

func (b v2AdInlineLiveBuilder) WithAfter(req ...func(*jsoncard.LargeCoverInline)) V2AdInlineLiveBuilder {
	b.afterFn = append(b.afterFn, req...)
	return b
}
