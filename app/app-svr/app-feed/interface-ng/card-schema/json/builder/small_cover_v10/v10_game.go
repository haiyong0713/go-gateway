package small_cover_v10

import (
	"fmt"
	"math"
	"net/url"

	"go-common/library/log"
	appcardmodel "go-gateway/app/app-svr/app-card/interface/model"
	"go-gateway/app/app-svr/app-card/interface/model/card"
	"go-gateway/app/app-svr/app-card/interface/model/card/ai"
	"go-gateway/app/app-svr/app-card/interface/model/card/game"
	cardschema "go-gateway/app/app-svr/app-feed/interface-ng/card-schema"
	jsoncard "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json"
	jsonbuilder "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder"
	jsonavatar "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder/avatar"
	jsoncommon "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder/common"
	jsonreasonstyle "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder/reason_style"

	tunnelV2 "git.bilibili.co/bapis/bapis-go/ai/feed/mgr/service"
	"github.com/pkg/errors"
)

var (
	badgeMap = map[string]*card.ReasonStyle{
		"1_1": {
			IconURL:      "https://i0.hdslb.com/bfs/activity-plat/static/20220314/0977767b2e79d8ad0a36a731068a83d7/XD3YCJIhhv.png",
			IconURLNight: "https://i0.hdslb.com/bfs/activity-plat/static/20220314/0977767b2e79d8ad0a36a731068a83d7/iTh2CQg1YX.png",
			IconWidth:    111,
			IconHeight:   22,
		},
		"1_2": {
			IconURL:      "https://i0.hdslb.com/bfs/activity-plat/static/20220314/0977767b2e79d8ad0a36a731068a83d7/GByBp5LMWF.png",
			IconURLNight: "https://i0.hdslb.com/bfs/activity-plat/static/20220314/0977767b2e79d8ad0a36a731068a83d7/3cNU6BZ1aL.png",
			IconWidth:    113,
			IconHeight:   22,
		},
		"1_3": {
			IconURL:      "https://i0.hdslb.com/bfs/activity-plat/static/20220314/0977767b2e79d8ad0a36a731068a83d7/MxbANa3J72.png",
			IconURLNight: "https://i0.hdslb.com/bfs/activity-plat/static/20220314/0977767b2e79d8ad0a36a731068a83d7/B2Oksm4t9V.png",
			IconWidth:    113,
			IconHeight:   22,
		},
		"1_4": {
			IconURL:      "https://i0.hdslb.com/bfs/activity-plat/static/20220314/0977767b2e79d8ad0a36a731068a83d7/MB4IQgpoA8.png",
			IconURLNight: "https://i0.hdslb.com/bfs/activity-plat/static/20220314/0977767b2e79d8ad0a36a731068a83d7/zrM6QfENFF.png",
			IconWidth:    113,
			IconHeight:   22,
		},
		"1_5": {
			IconURL:      "https://i0.hdslb.com/bfs/activity-plat/static/20220314/0977767b2e79d8ad0a36a731068a83d7/gJN43PzdI1.png",
			IconURLNight: "https://i0.hdslb.com/bfs/activity-plat/static/20220314/0977767b2e79d8ad0a36a731068a83d7/heoPGI2Wvt.png",
			IconWidth:    113,
			IconHeight:   22,
		},
		"1_6": {
			IconURL:      "https://i0.hdslb.com/bfs/activity-plat/static/20220314/0977767b2e79d8ad0a36a731068a83d7/LqXv6ZZN86.png",
			IconURLNight: "https://i0.hdslb.com/bfs/activity-plat/static/20220314/0977767b2e79d8ad0a36a731068a83d7/avdT6Nqdke.png",
			IconWidth:    113,
			IconHeight:   22,
		},
		"1_7": {
			IconURL:      "https://i0.hdslb.com/bfs/activity-plat/static/20220314/0977767b2e79d8ad0a36a731068a83d7/O1qfv4jvk6.png",
			IconURLNight: "https://i0.hdslb.com/bfs/activity-plat/static/20220314/0977767b2e79d8ad0a36a731068a83d7/FOLDEjPnDw.png",
			IconWidth:    113,
			IconHeight:   22,
		},
		"1_8": {
			IconURL:      "https://i0.hdslb.com/bfs/activity-plat/static/20220314/0977767b2e79d8ad0a36a731068a83d7/vXYa3ICiOb.png",
			IconURLNight: "https://i0.hdslb.com/bfs/activity-plat/static/20220314/0977767b2e79d8ad0a36a731068a83d7/S4Ihi6v77I.png",
			IconWidth:    113,
			IconHeight:   22,
		},
		"1_9": {
			IconURL:      "https://i0.hdslb.com/bfs/activity-plat/static/20220314/0977767b2e79d8ad0a36a731068a83d7/9LIbd98muA.png",
			IconURLNight: "https://i0.hdslb.com/bfs/activity-plat/static/20220314/0977767b2e79d8ad0a36a731068a83d7/HZpncjmheP.png",
			IconWidth:    113,
			IconHeight:   22,
		},
		"1_10": {
			IconURL:      "https://i0.hdslb.com/bfs/activity-plat/static/20220314/0977767b2e79d8ad0a36a731068a83d7/mqqaJmcqBA.png",
			IconURLNight: "https://i0.hdslb.com/bfs/activity-plat/static/20220314/0977767b2e79d8ad0a36a731068a83d7/f6PS2KMtxV.png",
			IconWidth:    117,
			IconHeight:   22,
		},
		"5_1": {
			IconURL:      "https://i0.hdslb.com/bfs/activity-plat/static/20220314/0977767b2e79d8ad0a36a731068a83d7/lOa1gNvG3f.png",
			IconURLNight: "https://i0.hdslb.com/bfs/activity-plat/static/20220314/0977767b2e79d8ad0a36a731068a83d7/id3IYpMSzE.png",
			IconWidth:    111,
			IconHeight:   22,
		},
		"5_2": {
			IconURL:      "https://i0.hdslb.com/bfs/activity-plat/static/20220314/0977767b2e79d8ad0a36a731068a83d7/ucDkinU4hl.png",
			IconURLNight: "https://i0.hdslb.com/bfs/activity-plat/static/20220314/0977767b2e79d8ad0a36a731068a83d7/2WeyBDzGBj.png",
			IconWidth:    113,
			IconHeight:   22,
		},
		"5_3": {
			IconURL:      "https://i0.hdslb.com/bfs/activity-plat/static/20220314/0977767b2e79d8ad0a36a731068a83d7/Y1imhOGTIH.png",
			IconURLNight: "https://i0.hdslb.com/bfs/activity-plat/static/20220314/0977767b2e79d8ad0a36a731068a83d7/xDGL1iQbBM.png",
			IconWidth:    113,
			IconHeight:   22,
		},
		"5_4": {
			IconURL:      "https://i0.hdslb.com/bfs/activity-plat/static/20220314/0977767b2e79d8ad0a36a731068a83d7/Yt0jQfwp9Y.png",
			IconURLNight: "https://i0.hdslb.com/bfs/activity-plat/static/20220314/0977767b2e79d8ad0a36a731068a83d7/0olaKlOpmA.png",
			IconWidth:    113,
			IconHeight:   22,
		},
		"5_5": {
			IconURL:      "https://i0.hdslb.com/bfs/activity-plat/static/20220314/0977767b2e79d8ad0a36a731068a83d7/1R1eBplOgJ.png",
			IconURLNight: "https://i0.hdslb.com/bfs/activity-plat/static/20220314/0977767b2e79d8ad0a36a731068a83d7/UR8yIgnhqT.png",
			IconWidth:    113,
			IconHeight:   22,
		},
		"5_6": {
			IconURL:      "https://i0.hdslb.com/bfs/activity-plat/static/20220314/0977767b2e79d8ad0a36a731068a83d7/5XdKov3Cu4.png",
			IconURLNight: "https://i0.hdslb.com/bfs/activity-plat/static/20220314/0977767b2e79d8ad0a36a731068a83d7/9bSe9TWdse.png",
			IconWidth:    113,
			IconHeight:   22,
		},
		"5_7": {
			IconURL:      "https://i0.hdslb.com/bfs/activity-plat/static/20220314/0977767b2e79d8ad0a36a731068a83d7/ZXvhyYQt2y.png",
			IconURLNight: "https://i0.hdslb.com/bfs/activity-plat/static/20220314/0977767b2e79d8ad0a36a731068a83d7/znce4KWeVi.png",
			IconWidth:    113,
			IconHeight:   22,
		},
		"5_8": {
			IconURL:      "https://i0.hdslb.com/bfs/activity-plat/static/20220314/0977767b2e79d8ad0a36a731068a83d7/rzYDmuK9PW.png",
			IconURLNight: "https://i0.hdslb.com/bfs/activity-plat/static/20220314/0977767b2e79d8ad0a36a731068a83d7/8h3HD3vk88.png",
			IconWidth:    113,
			IconHeight:   22,
		},
		"5_9": {
			IconURL:      "https://i0.hdslb.com/bfs/activity-plat/static/20220314/0977767b2e79d8ad0a36a731068a83d7/tcJWViZgln.png",
			IconURLNight: "https://i0.hdslb.com/bfs/activity-plat/static/20220314/0977767b2e79d8ad0a36a731068a83d7/mBdoBHwRcd.png",
			IconWidth:    113,
			IconHeight:   22,
		},
		"5_10": {
			IconURL:      "https://i0.hdslb.com/bfs/activity-plat/static/20220314/0977767b2e79d8ad0a36a731068a83d7/7AhMGAmm8V.png",
			IconURLNight: "https://i0.hdslb.com/bfs/activity-plat/static/20220314/0977767b2e79d8ad0a36a731068a83d7/1fCID70ZIH.png",
			IconWidth:    117,
			IconHeight:   22,
		},
		"6_1": {
			IconURL:      "https://i0.hdslb.com/bfs/activity-plat/static/20220314/0977767b2e79d8ad0a36a731068a83d7/pjt3KPuZAG.png",
			IconURLNight: "https://i0.hdslb.com/bfs/activity-plat/static/20220314/0977767b2e79d8ad0a36a731068a83d7/PptWhh0ZRN.png",
			IconWidth:    111,
			IconHeight:   22,
		},
		"6_2": {
			IconURL:      "https://i0.hdslb.com/bfs/activity-plat/static/20220314/0977767b2e79d8ad0a36a731068a83d7/k80MJ43lQ8.png",
			IconURLNight: "https://i0.hdslb.com/bfs/activity-plat/static/20220314/0977767b2e79d8ad0a36a731068a83d7/uWrNpSae20.png",
			IconWidth:    113,
			IconHeight:   22,
		},
		"6_3": {
			IconURL:      "https://i0.hdslb.com/bfs/activity-plat/static/20220314/0977767b2e79d8ad0a36a731068a83d7/IMTymWX39V.png",
			IconURLNight: "https://i0.hdslb.com/bfs/activity-plat/static/20220314/0977767b2e79d8ad0a36a731068a83d7/4UtEAgPfvZ.png",
			IconWidth:    113,
			IconHeight:   22,
		},
		"6_4": {
			IconURL:      "https://i0.hdslb.com/bfs/activity-plat/static/20220314/0977767b2e79d8ad0a36a731068a83d7/xxdBzJoyLz.png",
			IconURLNight: "https://i0.hdslb.com/bfs/activity-plat/static/20220314/0977767b2e79d8ad0a36a731068a83d7/nD1PXLgghH.png",
			IconWidth:    113,
			IconHeight:   22,
		},
		"6_5": {
			IconURL:      "https://i0.hdslb.com/bfs/activity-plat/static/20220314/0977767b2e79d8ad0a36a731068a83d7/PhYWAFMLCh.png",
			IconURLNight: "https://i0.hdslb.com/bfs/activity-plat/static/20220314/0977767b2e79d8ad0a36a731068a83d7/PCxfqIEjbI.png",
			IconWidth:    113,
			IconHeight:   22,
		},
		"6_6": {
			IconURL:      "https://i0.hdslb.com/bfs/activity-plat/static/20220314/0977767b2e79d8ad0a36a731068a83d7/q4C37JsUPV.png",
			IconURLNight: "https://i0.hdslb.com/bfs/activity-plat/static/20220314/0977767b2e79d8ad0a36a731068a83d7/E4OGn0gYHd.png",
			IconWidth:    113,
			IconHeight:   22,
		},
		"6_7": {
			IconURL:      "https://i0.hdslb.com/bfs/activity-plat/static/20220314/0977767b2e79d8ad0a36a731068a83d7/86Tf5sSqEO.png",
			IconURLNight: "https://i0.hdslb.com/bfs/activity-plat/static/20220314/0977767b2e79d8ad0a36a731068a83d7/mAAfPptBrH.png",
			IconWidth:    113,
			IconHeight:   22,
		},
		"6_8": {
			IconURL:      "https://i0.hdslb.com/bfs/activity-plat/static/20220314/0977767b2e79d8ad0a36a731068a83d7/oHECVTSzv0.png",
			IconURLNight: "https://i0.hdslb.com/bfs/activity-plat/static/20220314/0977767b2e79d8ad0a36a731068a83d7/5YG0UyAF1c.png",
			IconWidth:    113,
			IconHeight:   22,
		},
		"6_9": {
			IconURL:      "https://i0.hdslb.com/bfs/activity-plat/static/20220314/0977767b2e79d8ad0a36a731068a83d7/xnnDThMC5R.png",
			IconURLNight: "https://i0.hdslb.com/bfs/activity-plat/static/20220314/0977767b2e79d8ad0a36a731068a83d7/NcRnHgA3FV.png",
			IconWidth:    113,
			IconHeight:   22,
		},
		"6_10": {
			IconURL:      "https://i0.hdslb.com/bfs/activity-plat/static/20220314/0977767b2e79d8ad0a36a731068a83d7/Jh577OveiI.png",
			IconURLNight: "https://i0.hdslb.com/bfs/activity-plat/static/20220314/0977767b2e79d8ad0a36a731068a83d7/nNiQxtpU8V.png",
			IconWidth:    117,
			IconHeight:   22,
		},
	}
)

type V10GameBuilder interface {
	ReplaceContext(jsonbuilder.BuilderContext) V10GameBuilder
	SetBase(*jsoncard.Base) V10GameBuilder
	SetRcmd(*ai.Item) V10GameBuilder
	SetGame(*game.Game) V10GameBuilder
	Build() (*jsoncard.SmallCoverV10, error)
	WithAfter(req ...func(*jsoncard.SmallCoverV10)) V10GameBuilder
}

type v10GameBuilder struct {
	jsonbuilder.BuilderContext
	base           *jsoncard.Base
	rcmd           *ai.Item
	game           *game.Game
	threePoint     jsoncommon.ThreePoint
	hideRcmdReason bool
	afterFn        []func(*jsoncard.SmallCoverV10)
}

func NewV10GameBuilder(ctx jsonbuilder.BuilderContext) V10GameBuilder {
	return v10GameBuilder{BuilderContext: ctx}
}

func (b v10GameBuilder) ReplaceContext(ctx jsonbuilder.BuilderContext) V10GameBuilder {
	b.BuilderContext = ctx
	return b
}

func (b v10GameBuilder) SetBase(base *jsoncard.Base) V10GameBuilder {
	b.base = base
	return b
}

func (b v10GameBuilder) SetRcmd(in *ai.Item) V10GameBuilder {
	b.rcmd = in
	return b
}

func (b v10GameBuilder) SetGame(in *game.Game) V10GameBuilder {
	b.game = in
	return b
}

func (b v10GameBuilder) constructURI() string {
	return appcardmodel.FillURI("", 0, 0, b.game.GameLink, appcardmodel.GameHandler(b.rcmd, "100011"))
}

func (b v10GameBuilder) constructThreePointV2() []*jsoncard.ThreePointV2 {
	enableSwitchColumn := b.BuilderContext.FeatureGates().FeatureEnabled(cardschema.FeatureSwitchColumnThreePoint)
	return b.threePoint.ConstructDefaultThreePointV2(b.BuilderContext, enableSwitchColumn)
}

func (b v10GameBuilder) WithAfter(req ...func(v10 *jsoncard.SmallCoverV10)) V10GameBuilder {
	b.afterFn = append(b.afterFn, req...)
	return b
}

//nolint:gomnd
func (b *v10GameBuilder) constructCoverLeftText() string {
	switch b.game.GameStatusV2 {
	case 1:
		if b.game.BookNum <= 0 {
			b.hideRcmdReason = true
			return ""
		}
		return appcardmodel.Stat64String(b.game.BookNum, "人预约")
	case 2:
		if b.game.DownloadNum <= 0 {
			b.hideRcmdReason = true
			return ""
		}
		return appcardmodel.Stat64String(b.game.DownloadNum, "下载")
	default:
		return ""
	}
}

func (b v10GameBuilder) constructCoverRightText() string {
	const _normal = 2
	if b.game.GradeStatus == _normal {
		return fmt.Sprintf("%.1f分", b.game.Grade)
	}
	return ""
}

func (b v10GameBuilder) Build() (*jsoncard.SmallCoverV10, error) {
	if b.game == nil {
		return nil, errors.Errorf("empty `game` field")
	}
	if !b.game.IsOnline {
		return nil, errors.Errorf("game is offline")
	}
	if b.rcmd == nil {
		return nil, errors.Errorf("empty `rcmd` field")
	}
	if err := jsonbuilder.NewBaseUpdater(b.BuilderContext, b.base).
		UpdateCover(b.constructCover()).
		UpdateTitle(b.game.GameName).
		UpdateURI(b.constructURI()).
		UpdateThreePointV2(b.constructThreePointV2()).
		Update(); err != nil {
		return nil, err
	}
	out := &jsoncard.SmallCoverV10{
		Base:                         b.base,
		SubTitle:                     b.game.GameTags,
		CoverLeftText1:               b.constructCoverLeftText(),
		CoverLeft1ContentDescription: b.constructCoverLeftText(),
		CoverRightText:               b.constructCoverRightText(),
		CoverRightContentDescription: b.constructCoverRightText(),
		BadgeStyle:                   jsonreasonstyle.ConstructReasonStyle(appcardmodel.BgColorTransparentRed, "游戏"),
		LeftCoverBadgeNewStyle:       b.constructLeftCoverBadge(),
	}
	avatar, err := jsonavatar.NewAvatarBuilder(b.BuilderContext).
		SetAvatarStatus(&jsoncard.AvatarStatus{
			Cover: b.game.GameIcon,
			Type:  appcardmodel.AvatarSquare,
		}).Build()
	if err != nil {
		log.Error("Failed to build avatar: %+v", err)
	}
	out.Avatar = avatar
	out.RcmdReasonStyle, out.DescButton = b.constructReasonStyleOrDescButton()
	for _, fn := range b.afterFn {
		fn(out)
	}
	return out, nil
}

//nolint:gomnd
func (b v10GameBuilder) constructDescButton() *card.Button {
	text := ""
	switch b.game.GameStatusV2 {
	case 2:
		text = "立即下载"
	case 1:
		text = "立即预约"
	default:
		text = "立即查看"
	}
	if b.game.Notice != "" {
		text = b.game.Notice
	}
	return &card.Button{
		Text:    text,
		Event:   appcardmodel.EventChannelClick,
		EventV2: appcardmodel.EventV2ChannelClick,
		Type:    appcardmodel.ButtonGrey,
	}
}

func (b v10GameBuilder) constructCover() string {
	cover := b.game.Cover
	if b.game.MaterialsInfo != nil && b.game.MaterialsInfo.PromoteStatus == 1 && b.game.MaterialsInfo.ImageURL != "" {
		cover = b.game.MaterialsInfo.ImageURL
	}
	u, err := url.Parse(cover)
	if err != nil {
		log.Error("Failed to parse game cover: %+v", errors.WithStack(err))
		return cover
	}
	u.Scheme = "https"
	return u.String()
}

func (b v10GameBuilder) constructLeftCoverBadge() *card.ReasonStyle {
	if !b.rcmd.AllowGameBadge() {
		return nil
	}
	hasMngBadge := b.game.GameRankInfo != nil && b.game.GameRankInfo.TmDayIconURL != ""
	badge, ok := badgeMap[fmt.Sprintf("%d_%d", b.game.RankType, b.game.GameRank)]
	if !ok && !hasMngBadge {
		return nil
	}
	const _showHeight float64 = 22
	if hasMngBadge {
		return &card.ReasonStyle{
			IconURL:      b.game.GameRankInfo.TmDayIconURL,
			IconURLNight: b.game.GameRankInfo.TmNightIconURL,
			IconHeight:   int32(_showHeight),
			IconWidth:    int32(math.Floor(float64(b.game.GameRankInfo.TmIconWidth) / float64(b.game.GameRankInfo.TmIconHeight) * _showHeight)),
		}
	}
	return &card.ReasonStyle{
		IconURL:      badge.IconURL,
		IconURLNight: badge.IconURLNight,
		IconWidth:    badge.IconWidth,
		IconHeight:   badge.IconHeight,
	}
}

func (b v10GameBuilder) constructReasonStyleOrDescButton() (*jsoncard.ReasonStyle, *jsoncard.Button) {
	if b.rcmd.RcmdReason != nil && !b.hideRcmdReason {
		reasonStyle := jsonreasonstyle.ConstructTopReasonStyle(b.rcmd.RcmdReason.Content,
			jsonreasonstyle.CornerMarkFromAI(b.rcmd),
			jsonreasonstyle.CorverMarkFromContext(b.BuilderContext),
		)
		return reasonStyle, nil
	}
	return nil, b.constructDescButton()
}

func V10FilledByMultiMaterials(arg *tunnelV2.Material, item *ai.Item, needGif bool) func(*jsoncard.SmallCoverV10) {
	return func(card *jsoncard.SmallCoverV10) {
		if arg == nil {
			return
		}
		if arg.Title != "" {
			card.Title = arg.Title
		}
		if arg.Cover != "" {
			card.Cover = arg.Cover
		}
		if needGif && item.AllowGIF() && arg.GifCover != "" && item.StaticCover == 0 {
			card.CoverGif = arg.GifCover
		}
		if arg.Desc != "" && card.DescButton != nil {
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
