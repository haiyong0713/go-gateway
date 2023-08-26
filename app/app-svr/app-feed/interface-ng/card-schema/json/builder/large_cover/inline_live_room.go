package large_cover

import (
	"strconv"

	"go-common/library/log"
	appcardmodel "go-gateway/app/app-svr/app-card/interface/model"
	"go-gateway/app/app-svr/app-card/interface/model/card"
	"go-gateway/app/app-svr/app-card/interface/model/card/ai"
	"go-gateway/app/app-svr/app-card/interface/model/card/live"
	"go-gateway/app/app-svr/app-card/interface/model/card/threePointMeta"
	cardschema "go-gateway/app/app-svr/app-feed/interface-ng/card-schema"
	jsoncard "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json"
	jsonbuilder "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder"
	jsonavatar "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder/avatar"
	jsoncommon "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder/common"
	jsonreasonstyle "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder/reason_style"

	accountgrpc "git.bilibili.co/bapis/bapis-go/account/service"
	"github.com/pkg/errors"
)

type InlineLiveRoomBuilder interface {
	Parent() LargeCoverInlineBuilderFactory
	SetBase(*jsoncard.Base) InlineLiveRoomBuilder
	SetRcmd(*ai.Item) InlineLiveRoomBuilder
	SetLiveRoom(*live.Room) InlineLiveRoomBuilder
	SetInline(*Inline) InlineLiveRoomBuilder
	SetAuthorCard(*accountgrpc.Card) InlineLiveRoomBuilder

	Build() (*jsoncard.LargeCoverInline, error)
	WithAfter(req ...func(*jsoncard.LargeCoverInline)) InlineLiveRoomBuilder
}

type v8InlineLiveRoomBuilder struct {
	jsoncommon.LiveRoomCommon
	parent     *largeCoverInlineBuilderFactory
	base       *jsoncard.Base
	rcmd       *ai.Item
	liveRoom   *live.Room
	inline     *Inline
	authorCard *accountgrpc.Card
	afterFn    []func(*jsoncard.LargeCoverInline)
}

func (b v8InlineLiveRoomBuilder) Parent() LargeCoverInlineBuilderFactory {
	return b.parent
}

func (b v8InlineLiveRoomBuilder) SetBase(base *jsoncard.Base) InlineLiveRoomBuilder {
	b.base = base
	return b
}

func (b v8InlineLiveRoomBuilder) SetRcmd(in *ai.Item) InlineLiveRoomBuilder {
	b.rcmd = in
	return b
}

func (b v8InlineLiveRoomBuilder) SetLiveRoom(in *live.Room) InlineLiveRoomBuilder {
	b.liveRoom = in
	return b
}

func (b v8InlineLiveRoomBuilder) SetInline(in *Inline) InlineLiveRoomBuilder {
	b.inline = in
	return b
}

func (b v8InlineLiveRoomBuilder) SetAuthorCard(in *accountgrpc.Card) InlineLiveRoomBuilder {
	b.authorCard = in
	return b
}

func (b v8InlineLiveRoomBuilder) constructOfficialIcon() appcardmodel.Icon {
	return appcardmodel.OfficialIcon(b.authorCard)
}

func (b v8InlineLiveRoomBuilder) constructURI() string {
	device := b.parent.BuilderContext.Device()
	uri := appcardmodel.FillURI(appcardmodel.GotoLive, device.Plat(), int(device.Build()), strconv.FormatInt(b.liveRoom.RoomID, 10), appcardmodel.LiveRoomHandler(b.liveRoom, device.Network()))
	if b.liveRoom.Link != "" {
		uri = appcardmodel.FillURI("", 0, 0, b.liveRoom.Link, appcardmodel.URLTrackIDHandler(b.rcmd))
	}
	return uri
}

func (b v8InlineLiveRoomBuilder) constructArgs() jsoncard.Args {
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

func (b v8InlineLiveRoomBuilder) constructPlayerArgs() *jsoncard.PlayerArgs {
	return &jsoncard.PlayerArgs{
		RoomID:         b.liveRoom.RoomID,
		IsLive:         1,
		Type:           appcardmodel.GotoLive,
		ManualPlay:     b.rcmd.ManualInline(),
		HidePlayButton: appcardmodel.HidePlayButton,
	}
}

func (b v8InlineLiveRoomBuilder) constructThreePoint() *jsoncard.ThreePoint {
	if !b.parent.BuilderContext.FeatureGates().FeatureEnabled(cardschema.FeatureNewDislike) {
		return nil
	}
	return b.ConstructThreePointFromLiveRoom(b.liveRoom)
}

func (b v8InlineLiveRoomBuilder) constructThreePointV2() []*jsoncard.ThreePointV2 {
	if !b.parent.BuilderContext.FeatureGates().FeatureEnabled(cardschema.FeatureNewDislike) {
		return nil
	}
	return b.ConstructThreePointV2FromLiveRoom(b.parent.BuilderContext, b.liveRoom)
}

func (b v8InlineLiveRoomBuilder) Build() (*jsoncard.LargeCoverInline, error) {
	if b.base == nil {
		return nil, errors.Errorf("empty `base` field")
	}
	if b.rcmd == nil {
		return nil, errors.Errorf("empty `rcmd` field")
	}
	if b.liveRoom == nil {
		return nil, errors.Errorf("empty `liveRoom` field")
	}
	if b.inline == nil {
		return nil, errors.Errorf("empty `inline` field")
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
		UpdateThreePointMeta(b.constructThreePointPanelMeta()).
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
	text, icon, ok := b.LiveRoomCommon.ConstructCoverLeftMeta(b.liveRoom)
	if ok && b.parent.BuilderContext.VersionControl().Can("feed.enableLiveWatched") {
		out.CoverLeftIcon1 = icon
		out.CoverLeftText1 = text
	}
	out.CoverLeft1ContentDescription = appcardmodel.CoverIconContentDescription(out.CoverLeftIcon1,
		out.CoverLeftText1)
	out.CoverLeftText2 = b.liveRoom.AreaV2Name
	out.BadgeStyle = jsonreasonstyle.ConstructReasonStyle(appcardmodel.BgColorTransparentRed, "直播")
	if b.rcmd.RcmdReason != nil {
		reasonText, _ := jsonreasonstyle.BuildInlineReasonText(
			b.rcmd.RcmdReason,
			"",
			b.parent.BuilderContext.IsAttentionTo(b.liveRoom.UID),
			b.enableInlineRcmd(),
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
	out.SharePlane = b.constructSharePlane()
	out.Base = b.base
	for _, fn := range b.afterFn {
		fn(out)
	}
	return out, nil
}

func (b v8InlineLiveRoomBuilder) WithAfter(req ...func(*jsoncard.LargeCoverInline)) InlineLiveRoomBuilder {
	b.afterFn = append(b.afterFn, req...)
	return b
}

func SingleInlineLiveHideMeta() func(*jsoncard.LargeCoverInline) {
	return func(card *jsoncard.LargeCoverInline) {
		if card.PlayerArgs == nil {
			return
		}
		card.PlayerArgs.HidePlayButton = true
	}
}

func SingleInlineLivePrivateVal(room *live.Room, item *ai.Item) func(*jsoncard.LargeCoverInline) {
	return func(card *jsoncard.LargeCoverInline) {
		card.Desc = room.Uname
		// 因单双列直播inline使用同一卡片构造逻辑，但单双列jump_from不同，因此需对新单列特殊处理
		card.URI = appcardmodel.FillURI("", 0, 0, card.URI,
			appcardmodel.URLLiveHandler(item, "29016"))
	}
}

func (b v8InlineLiveRoomBuilder) constructThreePointPanelMeta() *threePointMeta.PanelMeta {
	const (
		_inlineShareOrigin = "tm_inline"
		_inlineLiveShareId = "tm.recommend.live.0"
	)
	if b.inline.ThreePointPanelType == 0 {
		return nil
	}
	return &threePointMeta.PanelMeta{
		PanelType:   int8(b.inline.ThreePointPanelType),
		ShareOrigin: _inlineShareOrigin,
		ShareId:     _inlineLiveShareId,
		FunctionalButtons: threePointMeta.ConstructFunctionalButton(true,
			b.parent.BuilderContext.FeatureGates().FeatureEnabled(cardschema.FeatureSwitchColumnThreePoint),
			appcardmodel.ColumnStatus(b.parent.BuilderContext.IndexParam().Column()),
			b.parent.BuilderContext.FeatureGates().FeatureEnabled(cardschema.FeatureDislikeText)),
	}
}

func (b v8InlineLiveRoomBuilder) constructSharePlane() *appcardmodel.SharePlane {
	return &appcardmodel.SharePlane{
		Title:      b.liveRoom.Title,
		Cover:      b.liveRoom.Cover,
		RoomId:     b.liveRoom.RoomID,
		ShareTo:    appcardmodel.ShareTo,
		Author:     b.liveRoom.Uname,
		AuthorId:   b.liveRoom.UID,
		AreaName:   b.liveRoom.AreaV2Name,
		AuthorFace: b.liveRoom.Face,
	}
}

func (b v8InlineLiveRoomBuilder) enableInlineRcmd() bool {
	if appcardmodel.Columnm[appcardmodel.ColumnStatus(b.parent.BuilderContext.IndexParam().Column())] == appcardmodel.ColumnSvrDouble {
		return true
	}
	return b.parent.BuilderContext.FeatureGates().FeatureEnabled(cardschema.FeatureSingleRcmdReason)
}
