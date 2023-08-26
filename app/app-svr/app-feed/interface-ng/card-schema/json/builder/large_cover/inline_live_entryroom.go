package large_cover

import (
	"strconv"

	"go-common/library/log"
	appcardmodel "go-gateway/app/app-svr/app-card/interface/model"
	"go-gateway/app/app-svr/app-card/interface/model/card"
	"go-gateway/app/app-svr/app-card/interface/model/card/ai"
	"go-gateway/app/app-svr/app-card/interface/model/card/threePointMeta"
	cardschema "go-gateway/app/app-svr/app-feed/interface-ng/card-schema"
	jsoncard "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json"
	jsonbuilder "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder"
	jsonavatar "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder/avatar"
	jsoncommon "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder/common"
	jsonreasonstyle "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder/reason_style"

	accountgrpc "git.bilibili.co/bapis/bapis-go/account/service"
	livexroomgate "git.bilibili.co/bapis/bapis-go/live/xroom-gate"
	"github.com/pkg/errors"
)

type InlineLiveEntryRoomBuilder interface {
	Parent() LargeCoverInlineBuilderFactory
	SetBase(*jsoncard.Base) InlineLiveEntryRoomBuilder
	SetRcmd(*ai.Item) InlineLiveEntryRoomBuilder
	SetEntryFrom(string) InlineLiveEntryRoomBuilder
	SetInline(*Inline) InlineLiveEntryRoomBuilder
	SetLiveRoom(*livexroomgate.EntryRoomInfoResp_EntryList) InlineLiveEntryRoomBuilder
	SetAuthorCard(*accountgrpc.Card) InlineLiveEntryRoomBuilder

	Build() (*jsoncard.LargeCoverInline, error)
	WithAfter(req ...func(*jsoncard.LargeCoverInline)) InlineLiveEntryRoomBuilder
}

type v8InlineLiveEntryRoomBuilder struct {
	jsoncommon.LiveEntryRoomCommon
	parent     *largeCoverInlineBuilderFactory
	base       *jsoncard.Base
	rcmd       *ai.Item
	inline     *Inline
	entryFrom  string
	liveRoom   *livexroomgate.EntryRoomInfoResp_EntryList
	authorCard *accountgrpc.Card
	afterFn    []func(*jsoncard.LargeCoverInline)
}

func (b v8InlineLiveEntryRoomBuilder) Parent() LargeCoverInlineBuilderFactory {
	return b.parent
}

func (b v8InlineLiveEntryRoomBuilder) SetBase(base *jsoncard.Base) InlineLiveEntryRoomBuilder {
	b.base = base
	return b
}

func (b v8InlineLiveEntryRoomBuilder) SetRcmd(in *ai.Item) InlineLiveEntryRoomBuilder {
	b.rcmd = in
	return b
}

func (b v8InlineLiveEntryRoomBuilder) SetLiveRoom(in *livexroomgate.EntryRoomInfoResp_EntryList) InlineLiveEntryRoomBuilder {
	b.liveRoom = in
	return b
}

func (b v8InlineLiveEntryRoomBuilder) SetAuthorCard(in *accountgrpc.Card) InlineLiveEntryRoomBuilder {
	b.authorCard = in
	return b
}

func (b v8InlineLiveEntryRoomBuilder) SetEntryFrom(in string) InlineLiveEntryRoomBuilder {
	b.entryFrom = in
	return b
}

func (b v8InlineLiveEntryRoomBuilder) SetInline(in *Inline) InlineLiveEntryRoomBuilder {
	b.inline = in
	return b
}

func (b v8InlineLiveEntryRoomBuilder) constructOfficialIcon() appcardmodel.Icon {
	return appcardmodel.OfficialIcon(b.authorCard)
}

func (b v8InlineLiveEntryRoomBuilder) constructURI() string {
	device := b.parent.BuilderContext.Device()
	uri := appcardmodel.FillURI(appcardmodel.GotoLive,
		device.Plat(), int(device.Build()),
		strconv.FormatInt(b.liveRoom.RoomId, 10),
		appcardmodel.LiveEntryHandler(b.liveRoom, b.entryFrom))
	return uri
}

func (b v8InlineLiveEntryRoomBuilder) constructArgs() jsoncard.Args {
	out := jsoncard.Args{}
	out.UpID = b.liveRoom.Uid
	out.UpName = b.authorCard.Name
	out.Rid = int32(b.liveRoom.ParentAreaId)
	out.Rname = b.liveRoom.ParentAreaName
	out.Tid = b.liveRoom.AreaId
	out.Tname = b.liveRoom.AreaName
	out.RoomID = b.liveRoom.RoomId
	out.Online = int32(b.liveRoom.PopularityCount)
	out.IsFollow = 0
	if b.parent.BuilderContext.IsAttentionTo(b.liveRoom.Uid) {
		out.IsFollow = 1
	}
	return out
}

func (b v8InlineLiveEntryRoomBuilder) constructPlayerArgs() *jsoncard.PlayerArgs {
	return &jsoncard.PlayerArgs{
		RoomID:         b.liveRoom.RoomId,
		IsLive:         1,
		Type:           appcardmodel.GotoLive,
		ManualPlay:     b.rcmd.ManualInline(),
		HidePlayButton: appcardmodel.HidePlayButton,
	}
}

func (b v8InlineLiveEntryRoomBuilder) constructThreePoint() *jsoncard.ThreePoint {
	if !b.parent.BuilderContext.FeatureGates().FeatureEnabled(cardschema.FeatureNewDislike) {
		return nil
	}
	return b.ConstructThreePointFromLiveEntryRoom(b.liveRoom, b.authorCard)
}

func (b v8InlineLiveEntryRoomBuilder) constructThreePointV2() []*jsoncard.ThreePointV2 {
	if !b.parent.BuilderContext.FeatureGates().FeatureEnabled(cardschema.FeatureNewDislike) {
		return nil
	}
	return b.ConstructThreePointV2FromLiveEntryRoom(b.liveRoom, b.authorCard)
}

func isPublicLiveRoom(in *livexroomgate.EntryRoomInfoResp_EntryList) bool {
	// 一旦有一个是 true 就不是公开的直播间
	return !(in.IsEncryptRoom || in.IsPayRoom || in.IsLockRoom || in.IsHiddenRoom)
}

func (b v8InlineLiveEntryRoomBuilder) Build() (*jsoncard.LargeCoverInline, error) {
	if b.base == nil {
		return nil, errors.Errorf("empty `base` field")
	}
	if b.rcmd == nil {
		return nil, errors.Errorf("empty `rcmd` field")
	}
	if b.liveRoom == nil {
		return nil, errors.Errorf("empty `liveRoom` field")
	}
	if b.authorCard == nil {
		return nil, errors.Errorf("empty `authorCard` field")
	}
	if b.liveRoom.LiveStatus != 1 {
		return nil, errors.Errorf("ignore on live room live status: %+v", b.liveRoom)
	}
	if !isPublicLiveRoom(b.liveRoom) {
		return nil, errors.Errorf("ignore on private live room: %+v", b.liveRoom)
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
		RightTopLiveBadge: card.ConstructRightTopLiveBadge(int8(b.liveRoom.LiveStatus)),
	}
	out.CanPlay = 1
	out.CoverLeftText1, out.CoverLeftIcon1 = b.constructCoverLeftMetaFromLiveRoom()
	out.CoverLeftText2 = b.liveRoom.AreaName
	out.BadgeStyle = jsonreasonstyle.ConstructReasonStyle(appcardmodel.BgColorTransparentRed, "直播")
	if b.rcmd.RcmdReason != nil || b.parent.BuilderContext.IsAttentionTo(b.liveRoom.Uid) {
		reasonText, _ := jsonreasonstyle.BuildRecommendReasonText(
			b.parent.BuilderContext,
			b.rcmd.RcmdReason,
			b.rcmd.Goto,
			"",
			b.parent.BuilderContext.IsAttentionTo(b.liveRoom.Uid),
		)
		out.RcmdReasonStyle = jsonreasonstyle.ConstructTopReasonStyle(
			reasonText,
			jsonreasonstyle.CornerMarkFromAI(b.rcmd),
			jsonreasonstyle.CorverMarkFromContext(b.parent.BuilderContext),
		)
	}
	if b.parent.BuilderContext.IsAttentionTo(b.liveRoom.Uid) {
		out.OfficialIcon = appcardmodel.IconIsAttenm
		out.IsAtten = true
	}
	avatar, err := jsonavatar.NewAvatarBuilder(b.parent.BuilderContext).
		SetAvatarStatus(&jsoncard.AvatarStatus{
			Cover:      b.authorCard.Face,
			Text:       b.authorCard.Name,
			Goto:       appcardmodel.GotoMid,
			Param:      strconv.FormatInt(b.liveRoom.Uid, 10),
			Type:       appcardmodel.AvatarRound,
			FaceNftNew: b.authorCard.FaceNftNew,
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

func (b v8InlineLiveEntryRoomBuilder) WithAfter(req ...func(*jsoncard.LargeCoverInline)) InlineLiveEntryRoomBuilder {
	b.afterFn = append(b.afterFn, req...)
	return b
}

func SingleV8InlineDesc(authorCard *accountgrpc.Card) func(*jsoncard.LargeCoverInline) {
	return func(card *jsoncard.LargeCoverInline) {
		card.Desc = authorCard.Name
	}
}

func (b v8InlineLiveEntryRoomBuilder) constructThreePointPanelMeta() *threePointMeta.PanelMeta {
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
			false,
			appcardmodel.ColumnStatus(b.parent.BuilderContext.IndexParam().Column()),
			false),
	}
}

func (b v8InlineLiveEntryRoomBuilder) constructSharePlane() *appcardmodel.SharePlane {
	return &appcardmodel.SharePlane{
		Title:      b.liveRoom.Title,
		Cover:      b.liveRoom.Cover,
		RoomId:     b.liveRoom.RoomId,
		ShareTo:    appcardmodel.ShareTo,
		Author:     b.authorCard.Name,
		AuthorId:   b.authorCard.Mid,
		AreaName:   b.liveRoom.AreaName,
		AuthorFace: b.authorCard.Face,
	}
}

func (b v8InlineLiveEntryRoomBuilder) constructCoverLeftMetaFromLiveRoom() (string, appcardmodel.Icon) {
	if b.liveRoom.WatchedShow != nil {
		text, icon, ok := b.LiveEntryRoomCommon.ConstructCoverLeftMeta(b.liveRoom)
		if ok && b.parent.BuilderContext.VersionControl().Can("feed.enableLiveWatched") {
			return text, icon
		}
	}
	return appcardmodel.StatString(int32(b.liveRoom.PopularityCount), ""), appcardmodel.IconOnline
}
