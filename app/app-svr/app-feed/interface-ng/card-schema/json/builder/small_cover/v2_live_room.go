package jsonsmallcover

import (
	"strconv"

	appcardmodel "go-gateway/app/app-svr/app-card/interface/model"
	appcard "go-gateway/app/app-svr/app-card/interface/model/card"
	"go-gateway/app/app-svr/app-card/interface/model/card/ai"
	"go-gateway/app/app-svr/app-card/interface/model/card/live"
	cardschema "go-gateway/app/app-svr/app-feed/interface-ng/card-schema"
	jsoncard "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json"
	jsonbuilder "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder"
	jsoncommon "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder/common"
	jsonreasonstyle "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder/reason_style"

	accountgrpc "git.bilibili.co/bapis/bapis-go/account/service"
	"github.com/pkg/errors"
)

type V2LiveRoomBuilder interface {
	Parent() SmallCoverV2BuilderFactory
	SetBase(*jsoncard.Base) V2LiveRoomBuilder
	SetRcmd(*ai.Item) V2LiveRoomBuilder
	SetLiveRoom(*live.Room) V2LiveRoomBuilder
	SetAuthorCard(*accountgrpc.Card) V2LiveRoomBuilder

	Build() (*jsoncard.SmallCoverV2, error)
	WithAfter(req ...func(*jsoncard.SmallCoverV2)) V2LiveRoomBuilder
}

type v2LiveRoomBuilder struct {
	jsoncommon.LiveRoomCommon
	parent     *smallCoverV2BuilderFactory
	base       *jsoncard.Base
	rcmd       *ai.Item
	liveRoom   *live.Room
	authorCard *accountgrpc.Card
	afterFn    []func(*jsoncard.SmallCoverV2)
}

func (b v2LiveRoomBuilder) Parent() SmallCoverV2BuilderFactory {
	return b.parent
}

func (b v2LiveRoomBuilder) SetBase(base *jsoncard.Base) V2LiveRoomBuilder {
	b.base = base
	return b
}

func (b v2LiveRoomBuilder) SetRcmd(in *ai.Item) V2LiveRoomBuilder {
	b.rcmd = in
	return b
}

func (b v2LiveRoomBuilder) SetLiveRoom(in *live.Room) V2LiveRoomBuilder {
	b.liveRoom = in
	return b
}

func (b v2LiveRoomBuilder) SetAuthorCard(in *accountgrpc.Card) V2LiveRoomBuilder {
	b.authorCard = in
	return b
}

func (b v2LiveRoomBuilder) constructOfficialIcon() appcardmodel.Icon {
	return appcardmodel.OfficialIcon(b.authorCard)
}

func (b v2LiveRoomBuilder) constructURI() string {
	device := b.parent.BuilderContext.Device()
	uri := appcardmodel.FillURI(appcardmodel.GotoLive, device.Plat(), int(device.Build()), strconv.FormatInt(b.liveRoom.RoomID, 10), appcardmodel.LiveRoomHandler(b.liveRoom, device.Network()))
	if b.liveRoom.Link != "" {
		uri = appcardmodel.FillURI("", 0, 0, b.liveRoom.Link, appcardmodel.URLLiveHandler(b.rcmd, "29015"))
	}
	return uri
}

func (b v2LiveRoomBuilder) constructArgs() jsoncard.Args {
	return b.ConstructArgsFromLiveRoom(b.liveRoom)
}

func (b v2LiveRoomBuilder) constructPlayerArgs() *jsoncard.PlayerArgs {
	return b.ConstructPlayerArgsFromLiveRoom(b.liveRoom, 0)
}

func (b v2LiveRoomBuilder) constructDescButton() *jsoncard.Button {
	return b.ConstructDescButtonFromLiveRoom(b.liveRoom)
}

func (b v2LiveRoomBuilder) Build() (*jsoncard.SmallCoverV2, error) {
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
		UpdateTitle(b.liveRoom.Title).
		UpdateURI(b.constructURI()).
		UpdateArgs(b.constructArgs()).
		UpdatePlayerArgs(b.constructPlayerArgs()).
		UpdateThreePoint(b.constructThreePoint()).
		UpdateThreePointV2(b.constructThreePointV2()).
		Update(); err != nil {
		return nil, err
	}
	out := &jsoncard.SmallCoverV2{
		OfficialIcon: b.constructOfficialIcon(),
		DescButton:   b.constructDescButton(),
	}
	out.CanPlay = 1
	out.CoverLeftText1 = appcardmodel.StatString(b.liveRoom.Online, "")
	out.CoverLeftIcon1 = appcardmodel.IconOnline
	out.CoverLeft1ContentDescription = appcardmodel.CoverIconContentDescription(out.CoverLeftIcon1,
		out.CoverLeftText1)
	out.CoverRightText = b.liveRoom.Uname
	out.CoverRightContentDescription = b.liveRoom.Uname
	out.Badge = "直播"
	out.BadgeStyle = jsonreasonstyle.ConstructReasonStyle(appcardmodel.BgColorTransparentRed, "直播")
	if b.rcmd.RcmdReason != nil || b.parent.BuilderContext.IsAttentionTo(b.liveRoom.UID) {
		out.DescButton = nil
		reasonText, desc := jsonreasonstyle.BuildRecommendReasonText(
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
		out.RcmdReason = reasonText
		out.Desc = desc
	}
	out.Base = b.base
	for _, fn := range b.afterFn {
		fn(out)
	}
	return out, nil
}

func (b v2LiveRoomBuilder) constructThreePoint() *jsoncard.ThreePoint {
	if !b.parent.BuilderContext.FeatureGates().FeatureEnabled(cardschema.FeatureNewDislike) {
		return nil
	}
	return b.ConstructThreePointFromLiveRoom(b.liveRoom)
}

func (b v2LiveRoomBuilder) constructThreePointV2() []*jsoncard.ThreePointV2 {
	if !b.parent.BuilderContext.FeatureGates().FeatureEnabled(cardschema.FeatureNewDislike) {
		return nil
	}
	return b.ConstructThreePointV2FromLiveRoom(b.parent.BuilderContext, b.liveRoom)
}

func (b v2LiveRoomBuilder) WithAfter(req ...func(*jsoncard.SmallCoverV2)) V2LiveRoomBuilder {
	b.afterFn = append(b.afterFn, req...)
	return b
}
