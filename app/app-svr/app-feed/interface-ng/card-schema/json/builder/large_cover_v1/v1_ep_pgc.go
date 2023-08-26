package large_cover_v1

import (
	"math"
	"strconv"

	"go-common/library/log"
	appcardmodel "go-gateway/app/app-svr/app-card/interface/model"
	"go-gateway/app/app-svr/app-card/interface/model/card/ai"
	cardschema "go-gateway/app/app-svr/app-feed/interface-ng/card-schema"
	jsoncard "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json"
	jsonbuilder "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder"
	jsonavatar "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder/avatar"
	jsoncommon "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder/common"
	jsonreasonstyle "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder/reason_style"

	deliverygrpc "git.bilibili.co/bapis/bapis-go/pgc/servant/delivery"
	episodegrpc "git.bilibili.co/bapis/bapis-go/pgc/service/season/episode"
	"github.com/pkg/errors"
)

type V1EpPGCBuilder interface {
	Parent() LargeCoverV1BuilderFactory
	SetBase(*jsoncard.Base) V1EpPGCBuilder
	SetRcmd(*ai.Item) V1EpPGCBuilder
	SetEpisode(*episodegrpc.EpisodeCardsProto) V1EpPGCBuilder

	Build() (*jsoncard.LargeCoverV1, error)
	WithAfter(req ...func(*jsoncard.LargeCoverV1)) V1EpPGCBuilder
}

type v1EpPGCBuilder struct {
	seasonCommon jsoncommon.BangumiSeason
	threePoint   jsoncommon.ThreePoint
	base         *jsoncard.Base
	rcmd         *ai.Item
	parent       *largeCoverV1BuilderFactory
	episode      *episodegrpc.EpisodeCardsProto
	afterFn      []func(*jsoncard.LargeCoverV1)
}

func (b v1EpPGCBuilder) Parent() LargeCoverV1BuilderFactory {
	return b.parent
}

func (b v1EpPGCBuilder) SetBase(base *jsoncard.Base) V1EpPGCBuilder {
	b.base = base
	return b
}

func (b v1EpPGCBuilder) SetRcmd(item *ai.Item) V1EpPGCBuilder {
	b.rcmd = item
	return b
}

func (b v1EpPGCBuilder) SetEpisode(in *episodegrpc.EpisodeCardsProto) V1EpPGCBuilder {
	b.episode = in
	return b
}

func (b v1EpPGCBuilder) constructThreePointV2() []*jsoncard.ThreePointV2 {
	enableSwitchColumn := b.parent.BuilderContext.FeatureGates().FeatureEnabled(cardschema.FeatureSwitchColumnThreePoint)
	if b.parent.BuilderContext.VersionControl().Can("feed.usingNewThreePointV2") {
		return b.threePoint.ConstructDefaultThreePointV2(b.parent.BuilderContext, enableSwitchColumn)
	}
	return b.threePoint.ConstructDefaultThreePointV2Legacy(b.parent.BuilderContext, enableSwitchColumn)
}

func (b v1EpPGCBuilder) constructAvatar() *jsoncard.Avatar {
	avatar, err := jsonavatar.NewAvatarBuilder(b.parent.BuilderContext).
		SetAvatarStatus(&jsoncard.AvatarStatus{
			Cover: b.episode.Season.Cover,
			Type:  appcardmodel.AvatarSquare,
		}).Build()
	if err != nil {
		log.Warn("Failed to build avatar: %+v", err)
	}
	return avatar
}

func (b v1EpPGCBuilder) Build() (*jsoncard.LargeCoverV1, error) {
	if b.base == nil {
		return nil, errors.Errorf("empty `base` field")
	}
	if b.rcmd == nil {
		return nil, errors.Errorf("empty `rcmd` field")
	}
	if b.episode == nil {
		return nil, errors.Errorf("empty `episode` field")
	}
	if b.episode.Season == nil {
		return nil, errors.Errorf("empty `episode.Season` field")
	}
	output := &jsoncard.LargeCoverV1{}
	if err := jsonbuilder.NewBaseUpdater(b.parent.BuilderContext, b.base).
		UpdateURI(b.seasonCommon.ConstructEpURI(b.episode, b.parent.Device(), b.rcmd)).
		UpdateCover(b.episode.Cover).
		UpdateTitle(b.seasonCommon.ConstructEpTitle(b.episode)).
		UpdateThreePoint(b.threePoint.ConstructDefaultThreePoint()).
		UpdateThreePointV2(b.constructThreePointV2()).
		UpdateParam(strconv.FormatInt(int64(b.episode.EpisodeId), 10)).
		Update(); err != nil {
		return nil, err
	}

	if b.episode.Season.Stat != nil {
		output.CoverLeftText2 = appcardmodel.ArchiveViewString(int32(b.episode.Season.Stat.View))
		output.CoverLeftText3 = appcardmodel.BangumiFavString(int32(b.episode.Season.Stat.Follow), b.episode.Season.SeasonType)
	}
	output.Avatar = b.constructAvatar()
	output.CoverBadge = b.episode.Season.SeasonTypeName
	output.CoverBadgeStyle = jsonreasonstyle.ConstructReasonStyle(appcardmodel.BgColorRed, b.episode.Season.SeasonTypeName)

	output.Desc = b.episode.Season.NewEpShow
	topRcmdReason, bottomRcmdReason := jsonreasonstyle.BuildTopBottomRecommendReasonText(
		b.parent.BuilderContext,
		b.rcmd.RcmdReason,
		b.rcmd.Goto,
		false,
	)
	output.TopRcmdReason = topRcmdReason
	output.BottomRcmdReason = bottomRcmdReason
	output.TopRcmdReasonStyle = jsonreasonstyle.ConstructTopReasonStyle(
		topRcmdReason,
		jsonreasonstyle.CornerMarkFromAI(b.rcmd),
		jsonreasonstyle.CorverMarkFromContext(b.parent.BuilderContext),
	)
	output.BottomRcmdReasonStyle = jsonreasonstyle.ConstructBottomReasonStyle(
		bottomRcmdReason,
		jsonreasonstyle.CornerMarkFromAI(b.rcmd),
		jsonreasonstyle.CorverMarkFromContext(b.parent.BuilderContext),
	)
	output.Base = b.base
	for _, fn := range b.afterFn {
		fn(output)
	}

	return output, nil
}

func (b v1EpPGCBuilder) WithAfter(req ...func(*jsoncard.LargeCoverV1)) V1EpPGCBuilder {
	b.afterFn = append(b.afterFn, req...)
	return b
}

func V1FilledByEpMaterials(arg *deliverygrpc.EpMaterial, item *ai.Item) func(v1 *jsoncard.LargeCoverV1) {
	return func(card *jsoncard.LargeCoverV1) {
		if arg == nil {
			return
		}
		if arg.Title != "" {
			card.Title = arg.Title
		}
		if arg.Cover != "" {
			card.Cover = arg.Cover
		}
		if item.AllowGIF() && arg.GifCover != "" && item.StaticCover == 0 {
			card.CoverGif = arg.GifCover
		}
		if arg.Desc != "" {
			if card.DescButton != nil {
				card.DescButton.Text = arg.Desc
			} else {
				card.Desc = arg.Desc
			}
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
