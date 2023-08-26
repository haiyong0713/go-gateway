package large_cover_v1

import (
	"go-common/library/log"
	appcardmodel "go-gateway/app/app-svr/app-card/interface/model"
	"go-gateway/app/app-svr/app-card/interface/model/card/ai"
	cardschema "go-gateway/app/app-svr/app-feed/interface-ng/card-schema"
	jsoncard "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json"
	jsonbuilder "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder"
	jsonavatar "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder/avatar"
	jsoncommon "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder/common"
	jsonreasonstyle "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder/reason_style"
	"strconv"

	pgcinline "git.bilibili.co/bapis/bapis-go/pgc/service/card/inline"

	"github.com/pkg/errors"
)

type V1InlinePGCBuilder interface {
	Parent() LargeCoverV1BuilderFactory
	SetBase(*jsoncard.Base) V1InlinePGCBuilder
	SetRcmd(*ai.Item) V1InlinePGCBuilder
	SetEpisode(*pgcinline.EpisodeCard) V1InlinePGCBuilder
	SetHasLike(map[int64]int8) V1InlinePGCBuilder
	Build() (*jsoncard.LargeCoverV1, error)
	WithAfter(req ...func(*jsoncard.LargeCoverV1)) v1InlinePGCBuilder
}

type v1InlinePGCBuilder struct {
	jsoncommon.ThreePoint
	parent  *largeCoverV1BuilderFactory
	base    *jsoncard.Base
	rcmd    *ai.Item
	hasLike map[int64]int8
	episode *pgcinline.EpisodeCard
	afterFn []func(*jsoncard.LargeCoverV1)
}

func (b v1InlinePGCBuilder) Parent() LargeCoverV1BuilderFactory {
	return b.parent
}

func (b v1InlinePGCBuilder) SetBase(base *jsoncard.Base) V1InlinePGCBuilder {
	b.base = base
	return b
}

func (b v1InlinePGCBuilder) SetRcmd(in *ai.Item) V1InlinePGCBuilder {
	b.rcmd = in
	return b
}

func (b v1InlinePGCBuilder) SetHasLike(in map[int64]int8) V1InlinePGCBuilder {
	b.hasLike = in
	return b
}

func (b v1InlinePGCBuilder) SetEpisode(in *pgcinline.EpisodeCard) V1InlinePGCBuilder {
	b.episode = in
	return b
}

func (b v1InlinePGCBuilder) constructThreePointV2() []*jsoncard.ThreePointV2 {
	enableSwitchColumn := b.parent.BuilderContext.FeatureGates().FeatureEnabled(cardschema.FeatureSwitchColumnThreePoint)
	if b.parent.BuilderContext.VersionControl().Can("feed.usingNewThreePointV2") {
		return b.ConstructDefaultThreePointV2(b.parent.BuilderContext, enableSwitchColumn)
	}
	return b.ConstructDefaultThreePointV2Legacy(b.parent.BuilderContext, enableSwitchColumn)
}

func (b v1InlinePGCBuilder) constructURI(in *pgcinline.EpisodeCard, device cardschema.Device, rcmd *ai.Item) string {
	param := in.Url
	if param == "" {
		plat := device.Plat()
		build := int(device.Build())
		param = appcardmodel.FillURI(appcardmodel.GotoBangumi, plat, build, strconv.FormatInt(int64(in.EpisodeId), 10), nil)
	}
	return appcardmodel.FillURI("", 0, 0, param, appcardmodel.PGCTrackIDHandler(rcmd))
}

func (b v1InlinePGCBuilder) constructPlayerArgs() *jsoncard.PlayerArgs {
	return &jsoncard.PlayerArgs{
		Aid:       b.episode.Aid,
		Cid:       b.episode.Cid,
		EpID:      int64(b.episode.EpisodeId),
		IsPreview: b.episode.IsPreview,
		Type:      appcardmodel.GotoBangumi,
		Duration:  b.episode.Duration,
		SubType:   b.episode.Season.Type,
		SeasonID:  int64(b.episode.Season.SeasonId),
	}
}

func (b v1InlinePGCBuilder) constructAvatar() *jsoncard.Avatar {
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

func (b v1InlinePGCBuilder) Build() (*jsoncard.LargeCoverV1, error) {
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
	if err := jsonbuilder.NewBaseUpdater(b.parent.BuilderContext, b.base).
		UpdateCover(b.episode.Cover).
		UpdateTitle(b.episode.Season.Title).
		UpdateURI(b.constructURI(b.episode, b.parent.Device(), b.rcmd)).
		UpdatePlayerArgs(b.constructPlayerArgs()).
		UpdateThreePoint(b.ConstructDefaultThreePoint()).
		UpdateThreePointV2(b.constructThreePointV2()).
		Update(); err != nil {
		return nil, err
	}
	out := &jsoncard.LargeCoverV1{
		Base:           b.base,
		CoverLeftText1: appcardmodel.DurationString(b.episode.Duration),
		Avatar:         b.constructAvatar(),
		Desc:           b.episode.NewDesc,
	}
	if b.episode.PlayerInfo != nil &&
		b.parent.BuilderContext.VersionControl().Can("pgc.inlinePGCAutoPlaySupported") {
		out.CanPlay = 1
	}
	if b.episode.Stat != nil {
		out.CoverLeftText2 = appcardmodel.ArchiveViewString(int32(b.episode.Stat.Play))
		out.CoverLeftText3 = appcardmodel.DanmakuString(int32(b.episode.Stat.Danmaku))
	}
	topRcmdReason, bottomRcmdReason := jsonreasonstyle.BuildTopBottomRecommendReasonText(
		b.parent.BuilderContext,
		b.rcmd.RcmdReason,
		b.rcmd.Goto,
		false,
	)
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
	for _, fn := range b.afterFn {
		fn(out)
	}

	return out, nil
}

func (b v1InlinePGCBuilder) WithAfter(req ...func(*jsoncard.LargeCoverV1)) v1InlinePGCBuilder {
	b.afterFn = append(b.afterFn, req...)
	return b
}
