package large_cover_v1

import (
	"strconv"

	"go-common/library/log"
	appcardmodel "go-gateway/app/app-svr/app-card/interface/model"
	appcard "go-gateway/app/app-svr/app-card/interface/model/card"
	"go-gateway/app/app-svr/app-card/interface/model/card/ai"
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

type V1LiveRoomBuilder interface {
	Parent() LargeCoverV1BuilderFactory
	SetBase(*jsoncard.Base) V1LiveRoomBuilder
	SetRcmd(*ai.Item) V1LiveRoomBuilder
	SetLiveRoom(*live.Room) V1LiveRoomBuilder
	SetAuthorCard(*accountgrpc.Card) V1LiveRoomBuilder

	Build() (*jsoncard.LargeCoverV1, error)
	WithAfter(req ...func(*jsoncard.LargeCoverV1)) V1LiveRoomBuilder
}

type v1LiveRoomBuilder struct {
	jsoncommon.LiveRoomCommon
	parent     *largeCoverV1BuilderFactory
	base       *jsoncard.Base
	rcmd       *ai.Item
	liveRoom   *live.Room
	authorCard *accountgrpc.Card
	afterFn    []func(*jsoncard.LargeCoverV1)
}

func (b v1LiveRoomBuilder) Parent() LargeCoverV1BuilderFactory {
	return b.parent
}

func (b v1LiveRoomBuilder) SetBase(base *jsoncard.Base) V1LiveRoomBuilder {
	b.base = base
	return b
}

func (b v1LiveRoomBuilder) SetRcmd(in *ai.Item) V1LiveRoomBuilder {
	b.rcmd = in
	return b
}

func (b v1LiveRoomBuilder) SetLiveRoom(in *live.Room) V1LiveRoomBuilder {
	b.liveRoom = in
	return b
}

func (b v1LiveRoomBuilder) SetAuthorCard(in *accountgrpc.Card) V1LiveRoomBuilder {
	b.authorCard = in
	return b
}

func (b v1LiveRoomBuilder) constructOfficialIcon() appcardmodel.Icon {
	return appcardmodel.OfficialIcon(b.authorCard)
}

func (b v1LiveRoomBuilder) constructURI() string {
	device := b.parent.BuilderContext.Device()
	uri := appcardmodel.FillURI(appcardmodel.GotoLive, device.Plat(), int(device.Build()), strconv.FormatInt(b.liveRoom.RoomID, 10), appcardmodel.LiveRoomHandler(b.liveRoom, device.Network()))
	if b.liveRoom.Link != "" {
		uri = appcardmodel.FillURI("", 0, 0, b.liveRoom.Link, appcardmodel.URLLiveHandler(b.rcmd, "29016"))
	}
	return uri
}

func (b v1LiveRoomBuilder) constructArgs() jsoncard.Args {
	return b.ConstructArgsFromLiveRoom(b.liveRoom)
}

func (b v1LiveRoomBuilder) constructPlayerArgs() *jsoncard.PlayerArgs {
	return b.ConstructPlayerArgsFromLiveRoom(b.liveRoom, 0)
}

func (b v1LiveRoomBuilder) constructThreePoint() *jsoncard.ThreePoint {
	if !b.parent.BuilderContext.FeatureGates().FeatureEnabled(cardschema.FeatureNewDislike) {
		return nil
	}
	return b.ConstructThreePointFromLiveRoom(b.liveRoom)
}

func (b v1LiveRoomBuilder) constructThreePointV2() []*jsoncard.ThreePointV2 {
	if !b.parent.BuilderContext.FeatureGates().FeatureEnabled(cardschema.FeatureNewDislike) {
		return nil
	}
	return b.ConstructThreePointV2FromLiveRoom(b.parent.BuilderContext, b.liveRoom)
}

func (b v1LiveRoomBuilder) constructDescButton() *jsoncard.Button {
	return b.ConstructDescButtonFromLiveRoom(b.liveRoom)
}

func (b v1LiveRoomBuilder) constructAvatar() *jsoncard.Avatar {
	avatar, err := jsonavatar.NewAvatarBuilder(b.parent.BuilderContext).
		SetAvatarStatus(&jsoncard.AvatarStatus{
			Cover: b.liveRoom.Cover,
			Goto:  appcardmodel.GotoMid,
			Param: strconv.FormatInt(b.liveRoom.UID, 10),
			Type:  appcardmodel.AvatarRound,
		}).Build()
	if err != nil {
		log.Warn("Failed to build avatar: %+v", err)
	}
	return avatar
}

func (b v1LiveRoomBuilder) Build() (*jsoncard.LargeCoverV1, error) {
	if b.base == nil {
		return nil, errors.Errorf("empty `base` field")
	}
	if b.rcmd == nil {
		return nil, errors.Errorf("empty `rcmd` field")
	}
	if b.liveRoom == nil {
		return nil, errors.Errorf("empty `liveRoom` field")
	}
	if appcard.CheckMidMaxInt32(b.liveRoom.UID) && b.parent.BuilderContext.VersionControl().Can("feed.disableInt64Mid") {
		return nil, errors.Errorf("ignore on maxint32 mid: %d", b.liveRoom.UID)
	}
	if b.liveRoom.LiveStatus != 1 {
		return nil, errors.Errorf("ignore on live room live status: %+v", b.liveRoom)
	}

	if err := jsonbuilder.NewBaseUpdater(b.parent.BuilderContext, b.base).
		UpdateCover(b.liveRoom.Cover).
		UpdateTitle(b.liveRoom.Uname).
		UpdateURI(b.constructURI()).
		UpdateArgs(b.constructArgs()).
		UpdatePlayerArgs(b.constructPlayerArgs()).
		UpdateThreePoint(b.constructThreePoint()).
		UpdateThreePointV2(b.constructThreePointV2()).
		UpdateBaseInnerDescButton(b.constructDescButton()).
		Update(); err != nil {
		return nil, err
	}

	out := &jsoncard.LargeCoverV1{
		CanPlay:         1,
		Desc:            b.liveRoom.Title,
		CoverBadge:      "直播",
		CoverBadgeStyle: jsonreasonstyle.ConstructReasonStyle(appcardmodel.BgColorRed, "直播"),
		OfficialIcon:    b.constructOfficialIcon(),
		Avatar:          b.constructAvatar(),
	}
	topRcmdReason, bottomRcmdReason := jsonreasonstyle.BuildTopBottomRecommendReasonText(
		b.parent.BuilderContext,
		b.rcmd.RcmdReason,
		b.rcmd.Goto,
		b.parent.BuilderContext.IsAttentionTo(b.liveRoom.UID),
	)
	b.resolveCover(out)
	out.TopRcmdReason = topRcmdReason
	out.BottomRcmdReason = bottomRcmdReason
	out.TopRcmdReasonStyle = jsonreasonstyle.ConstructTopReasonStyle(
		topRcmdReason,
		jsonreasonstyle.CornerMarkFromAI(b.rcmd),
		jsonreasonstyle.CorverMarkFromContext(b.parent.BuilderContext),
	)
	out.BottomRcmdReasonStyle = jsonreasonstyle.ConstructBottomReasonStyle(
		bottomRcmdReason,
		jsonreasonstyle.CornerMarkFromAI(b.rcmd),
		jsonreasonstyle.CorverMarkFromContext(b.parent.BuilderContext),
	)
	out.Base = b.base
	for _, fn := range b.afterFn {
		fn(out)
	}

	return out, nil
}

func (b v1LiveRoomBuilder) resolveCover(in *jsoncard.LargeCoverV1) {
	if b.parent.BuilderContext.VersionControl().Can("feed.enablePadNewCover") {
		in.CoverLeftText1 = appcardmodel.StatString(b.liveRoom.Online, "")
		in.CoverLeftIcon1 = appcardmodel.IconOnline
		return
	}
	in.CoverLeftText2 = appcardmodel.LiveOnlineString(b.liveRoom.Online)
}

func (b v1LiveRoomBuilder) WithAfter(req ...func(*jsoncard.LargeCoverV1)) V1LiveRoomBuilder {
	b.afterFn = append(b.afterFn, req...)
	return b
}
