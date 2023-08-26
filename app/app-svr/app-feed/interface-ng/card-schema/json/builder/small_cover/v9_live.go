package jsonsmallcover

import (
	"bytes"
	"fmt"
	"math"
	"sort"
	"strconv"

	appcardmodel "go-gateway/app/app-svr/app-card/interface/model"
	"go-gateway/app/app-svr/app-card/interface/model/card"
	"go-gateway/app/app-svr/app-card/interface/model/card/ai"
	"go-gateway/app/app-svr/app-card/interface/model/card/live"
	"go-gateway/app/app-svr/app-card/interface/model/card/operate"
	cardschema "go-gateway/app/app-svr/app-feed/interface-ng/card-schema"
	jsoncard "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json"
	jsonbuilder "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder"
	jsoncommon "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder/common"
	jsonreasonstyle "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder/reason_style"

	accountgrpc "git.bilibili.co/bapis/bapis-go/account/service"
	tunnelV2 "git.bilibili.co/bapis/bapis-go/ai/feed/mgr/service"
	"github.com/pkg/errors"
)

type V9LiveBuilder interface {
	ReplaceContext(jsonbuilder.BuilderContext) V9LiveBuilder
	SetBase(*jsoncard.Base) V9LiveBuilder
	SetLiveRoom(*live.Room) V9LiveBuilder
	SetAuthorCard(*accountgrpc.Card) V9LiveBuilder
	SetRcmd(*ai.Item) V9LiveBuilder
	SetLeftBottomBadgeStyle(*operate.LiveBottomBadge) V9LiveBuilder
	SetLeftCoverBadgeStyle([]*operate.V9LiveLeftCoverBadge) V9LiveBuilder
	Build() (*jsoncard.SmallCoverV9, error)

	WithAfter(req ...func(*jsoncard.SmallCoverV9)) V9LiveBuilder
}

type v9LiveBuilder struct {
	jsoncommon.LiveRoomCommon
	jsonbuilder.BuilderContext
	base                 *jsoncard.Base
	rcmd                 *ai.Item
	liveRoom             *live.Room
	authorCard           *accountgrpc.Card
	leftBottomBadgeStyle *operate.LiveBottomBadge
	leftCoverBadgeStyle  []*operate.V9LiveLeftCoverBadge
	afterFn              []func(*jsoncard.SmallCoverV9)
}

func NewV9LiveBuilder(ctx jsonbuilder.BuilderContext) V9LiveBuilder {
	return v9LiveBuilder{BuilderContext: ctx}
}

func (b v9LiveBuilder) ReplaceContext(ctx jsonbuilder.BuilderContext) V9LiveBuilder {
	b.BuilderContext = ctx
	return b
}

func (b v9LiveBuilder) SetBase(base *jsoncard.Base) V9LiveBuilder {
	b.base = base
	return b
}

func (b v9LiveBuilder) SetRcmd(in *ai.Item) V9LiveBuilder {
	b.rcmd = in
	return b
}

func (b v9LiveBuilder) SetLiveRoom(in *live.Room) V9LiveBuilder {
	b.liveRoom = in
	return b
}

func (b v9LiveBuilder) SetAuthorCard(in *accountgrpc.Card) V9LiveBuilder {
	b.authorCard = in
	return b
}

func (b v9LiveBuilder) SetLeftBottomBadgeStyle(in *operate.LiveBottomBadge) V9LiveBuilder {
	b.leftBottomBadgeStyle = in
	return b
}

func (b v9LiveBuilder) SetLeftCoverBadgeStyle(in []*operate.V9LiveLeftCoverBadge) V9LiveBuilder {
	b.leftCoverBadgeStyle = in
	return b
}

func (b v9LiveBuilder) WithAfter(req ...func(v9 *jsoncard.SmallCoverV9)) V9LiveBuilder {
	b.afterFn = append(b.afterFn, req...)
	return b
}

func (b v9LiveBuilder) constructURI() string {
	device := b.BuilderContext.Device()
	uri := appcardmodel.FillURI(appcardmodel.GotoLive, device.Plat(), int(device.Build()), strconv.FormatInt(b.liveRoom.RoomID, 10), appcardmodel.LiveRoomHandler(b.liveRoom, device.Network()))
	if b.liveRoom.Link != "" {
		uri = appcardmodel.FillURI("", 0, 0, b.liveRoom.Link, appcardmodel.URLLiveHandler(b.rcmd, "29015"))
	}
	return uri
}

func (b v9LiveBuilder) constructArgs() jsoncard.Args {
	return b.ConstructArgsFromLiveRoom(b.liveRoom)
}

func (b v9LiveBuilder) constructPlayerArgs(contentMode int64) *jsoncard.PlayerArgs {
	return b.ConstructPlayerArgsFromLiveRoom(b.liveRoom, contentMode)
}

func (b v9LiveBuilder) constructThreePoint() *jsoncard.ThreePoint {
	if !b.BuilderContext.FeatureGates().FeatureEnabled(cardschema.FeatureNewDislike) {
		return nil
	}
	return b.ConstructThreePointFromLiveRoom(b.liveRoom)
}

func (b v9LiveBuilder) constructThreePointV2() []*jsoncard.ThreePointV2 {
	if !b.BuilderContext.FeatureGates().FeatureEnabled(cardschema.FeatureNewDislike) {
		return nil
	}
	return b.ConstructThreePointV2FromLiveRoom(b.BuilderContext, b.liveRoom)
}

func (b v9LiveBuilder) constructOfficialIcon() appcardmodel.Icon {
	return appcardmodel.OfficialIcon(b.authorCard)
}

func (b v9LiveBuilder) constructLiveLeftCoverBadge(badgeStyle []*operate.V9LiveLeftCoverBadge) *jsoncard.ReasonStyle {
	if len(badgeStyle) == 0 {
		return nil
	}
	if len(b.liveRoom.AllPendants) == 0 && b.liveRoom.HotRank <= 0 {
		return nil
	}
	badgeMap := card.ConvertMap(badgeStyle)
	pendants := card.ConstructPendants(b.liveRoom)
	var badgeList []*operate.V9LiveLeftCoverBadge
	for _, pendant := range pendants {
		key := fmt.Sprintf("%s:%s", pendant.Type, pendant.Name)
		if b.rcmd.LiveCornerMark == 0 && key == "mobile_index_badge:红包抽奖" { // 0不展示礼物红包
			continue
		}
		badge, ok := badgeMap[key]
		if !ok {
			continue
		}
		badgeList = append(badgeList, badge)
	}
	if len(badgeList) == 0 {
		return nil
	}
	sort.Slice(badgeList, func(i, j int) bool { return badgeList[i].Priority < badgeList[j].Priority })
	newStyle := &jsoncard.ReasonStyle{
		Text:         badgeList[0].Text,
		IconURL:      badgeList[0].NewStyleIconURL,
		IconURLNight: badgeList[0].NewStyleIconURLNight,
		IconWidth:    badgeList[0].IconWidth,
		IconHeight:   badgeList[0].IconHeight,
	}
	return newStyle
}

func (b v9LiveBuilder) Build() (*jsoncard.SmallCoverV9, error) {
	if b.base == nil {
		return nil, errors.Errorf("empty `base` field")
	}
	if b.rcmd == nil {
		return nil, errors.Errorf("empty `rcmd` field")
	}
	if b.liveRoom == nil {
		return nil, errors.Errorf("empty `liveRoom` field")
	}
	if card.CheckMidMaxInt32(b.liveRoom.UID) && b.BuilderContext.VersionControl().Can("feed.disableInt64Mid") {
		return nil, errors.Errorf("ignore on maxint32 mid: %d", b.liveRoom.UID)
	}
	if b.liveRoom.LiveStatus != 1 {
		return nil, errors.Errorf("ignore on live room live status: %+v", b.liveRoom)
	}
	if err := jsonbuilder.NewBaseUpdater(b.BuilderContext, b.base).
		UpdateCover(b.liveRoom.Cover).
		UpdateTitle(b.liveRoom.Title).
		UpdateURI(b.constructURI()).
		UpdateArgs(b.constructArgs()).
		UpdatePlayerArgs(b.constructPlayerArgs(b.BuilderContext.FeatureGates().FeatureState(cardschema.FeatureLiveContentMode))).
		UpdateThreePoint(b.constructThreePoint()).
		UpdateThreePointV2(b.constructThreePointV2()).
		Update(); err != nil {
		return nil, err
	}
	up := &jsoncard.Up{
		ID:           b.liveRoom.UID,
		Name:         b.liveRoom.Uname,
		OfficialIcon: b.constructOfficialIcon(),
		Avatar: &jsoncard.Avatar{
			Cover: b.liveRoom.Face,
			URI:   b.constructURI(),
			Event: appcardmodel.EventMainCard,
		},
	}
	out := &jsoncard.SmallCoverV9{
		Base:                         b.base,
		OfficialIconV2:               b.constructOfficialIcon(),
		CanPlay:                      int32(b.rcmd.LiveInlineLight),
		CoverLeftText1:               appcardmodel.StatString(b.liveRoom.Online, ""),
		CoverLeftIcon1:               appcardmodel.IconOnline,
		CoverRightText:               b.liveRoom.AreaV2Name,
		LeftBottomRcmdReasonStyle:    b.ConstructLeftBottomRcmdReasonStyle(b.leftBottomBadgeStyle),
		LeftCoverBadgeNewStyle:       b.constructLiveLeftCoverBadge(b.leftCoverBadgeStyle),
		CoverRightContentDescription: b.liveRoom.AreaV2Name,
		OffBadgeStyle:                jsonreasonstyle.ConstructReasonStyle(appcardmodel.BgColorTransparentRed, "直播"),
	}
	text, icon, ok := b.LiveRoomCommon.ConstructCoverLeftMeta(b.liveRoom)
	if ok && b.BuilderContext.VersionControl().Can("feed.enableLiveWatched") {
		out.CoverLeftIcon1 = icon
		out.CoverLeftText1 = text
	}
	out.CoverLeft1ContentDescription = appcardmodel.CoverIconContentDescription(out.CoverLeftIcon1, out.CoverLeftText1)
	if b.BuilderContext.IsAttentionTo(b.liveRoom.UID) {
		up.OfficialIcon = appcardmodel.IconIsAttenm
		out.IsAtten = true
	}
	out.Up = up
	if b.rcmd.LiveInlineLightDanmu == 0 {
		out.DisableDanmu = true
		out.HideDanmuSwitch = true
	}
	for _, fn := range b.afterFn {
		fn(out)
	}
	return out, nil
}

func V9FilledByMultiMaterials(arg *tunnelV2.Material) func(*jsoncard.SmallCoverV9) {
	return func(card *jsoncard.SmallCoverV9) {
		if arg == nil {
			return
		}
		if arg.Title != "" {
			card.Title = arg.Title
		}
		if arg.Cover != "" {
			card.Cover = arg.Cover
		}
		if card.DescButton != nil && arg.Desc != "" {
			card.DescButton.Text = arg.Desc
		}
		if arg.GetPowerCorner().GetPowerPicSun() != "" && arg.GetPowerCorner().GetPowerPicNight() != "" &&
			arg.GetPowerCorner().GetWidth() > 0 && arg.GetPowerCorner().GetHeight() > 0 {
			card.LeftCoverBadgeNewStyle = &jsoncard.ReasonStyle{
				IconURL:      arg.GetPowerCorner().GetPowerPicSun(),
				IconURLNight: arg.GetPowerCorner().GetPowerPicNight(),
				IconWidth:    int32(math.Floor(float64(arg.GetPowerCorner().GetWidth()) / float64(arg.GetPowerCorner().GetHeight()) * float64(21))),
				IconHeight:   21,
			}
		}
	}
}

func SmallCoverV9TalkBack() func(*jsoncard.SmallCoverV9) {
	return func(card *jsoncard.SmallCoverV9) {
		buffer := bytes.Buffer{}
		buffer.WriteString(appcardmodel.TalkBackCardType(card.Goto) + ",")
		buffer.WriteString(card.Title + ",")
		buffer.WriteString(card.CoverLeft1ContentDescription + ",")
		buffer.WriteString("分区" + card.CoverRightContentDescription + ",")
		if card.Args.UpName != "" {
			buffer.WriteString("UP主" + card.Args.UpName)
		}
		card.TalkBack = buffer.String()
	}
}
