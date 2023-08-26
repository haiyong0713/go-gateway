package notify_tunnel

import (
	"encoding/json"
	"strconv"

	"go-common/library/log"
	"go-gateway/app/app-svr/app-card/interface/model"
	"go-gateway/app/app-svr/app-card/interface/model/card"
	"go-gateway/app/app-svr/app-card/interface/model/card/ai"
	"go-gateway/app/app-svr/app-card/interface/model/card/live"
	cardschema "go-gateway/app/app-svr/app-feed/interface-ng/card-schema"
	jsoncard "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json"
	jsonbuilder "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder"
	jsoncommon "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder/common"
	jsonlargecover "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder/large_cover"
	"go-gateway/app/app-svr/app-feed/interface-ng/card-schema/util/sets"
	arcgrpc "go-gateway/app/app-svr/archive/service/api"

	accountgrpc "git.bilibili.co/bapis/bapis-go/account/service"
	pgcinline "git.bilibili.co/bapis/bapis-go/pgc/service/card/inline"
	tunnelgrpc "git.bilibili.co/bapis/bapis-go/platform/service/tunnel"
	"github.com/pkg/errors"
)

type NotifyTunnelLargeV1Builder interface {
	ReplaceContext(jsonbuilder.BuilderContext) NotifyTunnelLargeV1Builder
	SetBase(*jsoncard.Base) NotifyTunnelLargeV1Builder
	SetTunnel(*tunnelgrpc.FeedCard) NotifyTunnelLargeV1Builder
	SetRcmd(*ai.Item) NotifyTunnelLargeV1Builder
	SetArcPlayer(map[int64]*arcgrpc.ArcPlayer) NotifyTunnelLargeV1Builder
	SetEpisode(map[int32]*pgcinline.EpisodeCard) NotifyTunnelLargeV1Builder
	SetLiveRoom(map[int64]*live.Room) NotifyTunnelLargeV1Builder
	SetAuthorCard(map[int64]*accountgrpc.Card) NotifyTunnelLargeV1Builder
	SetInline(*jsonlargecover.Inline) NotifyTunnelLargeV1Builder

	Build() (*jsoncard.UniversalNotifyTunnelLargeV1, error)
}

type notifyTunnelLargeV1Builder struct {
	jsonbuilder.BuilderContext
	rcmd       *ai.Item
	base       *jsoncard.Base
	threePoint jsoncommon.ThreePoint
	tunnel     *tunnelgrpc.FeedCard
	arcPlayer  map[int64]*arcgrpc.ArcPlayer
	episode    map[int32]*pgcinline.EpisodeCard
	liveRoom   map[int64]*live.Room
	authorCard map[int64]*accountgrpc.Card
	inline     *jsonlargecover.Inline
}

func NewNotifyTunnelLargeV1Builder(ctx jsonbuilder.BuilderContext) NotifyTunnelLargeV1Builder {
	return notifyTunnelLargeV1Builder{BuilderContext: ctx}
}

func (b notifyTunnelLargeV1Builder) ReplaceContext(ctx jsonbuilder.BuilderContext) NotifyTunnelLargeV1Builder {
	b.BuilderContext = ctx
	return b
}

func (b notifyTunnelLargeV1Builder) SetBase(base *jsoncard.Base) NotifyTunnelLargeV1Builder {
	b.base = base
	return b
}

func (b notifyTunnelLargeV1Builder) SetTunnel(tunnel *tunnelgrpc.FeedCard) NotifyTunnelLargeV1Builder {
	b.tunnel = tunnel
	return b
}

func (b notifyTunnelLargeV1Builder) SetRcmd(rcmd *ai.Item) NotifyTunnelLargeV1Builder {
	b.rcmd = rcmd
	return b
}

func (b notifyTunnelLargeV1Builder) SetArcPlayer(arcs map[int64]*arcgrpc.ArcPlayer) NotifyTunnelLargeV1Builder {
	b.arcPlayer = arcs
	return b
}

func (b notifyTunnelLargeV1Builder) SetEpisode(in map[int32]*pgcinline.EpisodeCard) NotifyTunnelLargeV1Builder {
	b.episode = in
	return b
}

func (b notifyTunnelLargeV1Builder) SetLiveRoom(in map[int64]*live.Room) NotifyTunnelLargeV1Builder {
	b.liveRoom = in
	return b
}

func (b notifyTunnelLargeV1Builder) SetAuthorCard(in map[int64]*accountgrpc.Card) NotifyTunnelLargeV1Builder {
	b.authorCard = in
	return b
}

func (b notifyTunnelLargeV1Builder) SetInline(in *jsonlargecover.Inline) NotifyTunnelLargeV1Builder {
	b.inline = in
	return b
}

func (b notifyTunnelLargeV1Builder) bigTunnelObject() (*ai.BigTunnelObject, error) {
	bigTunnelObject := &ai.BigTunnelObject{}
	if err := json.Unmarshal([]byte(b.rcmd.BigTunnelObject), &bigTunnelObject); err != nil {
		log.Error("Failed to unmarshal big tunnel object: %+v", errors.WithStack(err))
		return nil, errors.Errorf("Failed to unmarshal big tunnel object: %+v", err)
	}
	inlineTunnelType := sets.NewString("ugc", "pgc", "live")
	if inlineTunnelType.Has(bigTunnelObject.Type) && !b.BuilderContext.VersionControl().Can("feed.enableInlineTunnel") {
		return nil, errors.Errorf("Unsupport inline tunnel")
	}
	return bigTunnelObject, nil
}

func (b notifyTunnelLargeV1Builder) constructThreePointV2() []*jsoncard.ThreePointV2 {
	enableSwitchColumn := b.BuilderContext.FeatureGates().FeatureEnabled(cardschema.FeatureSwitchColumnThreePoint)
	if b.BuilderContext.VersionControl().Can("feed.usingNewThreePointV2") {
		return b.threePoint.ConstructDefaultThreePointV2(b.BuilderContext, enableSwitchColumn)
	}
	return b.threePoint.ConstructDefaultThreePointV2Legacy(b.BuilderContext, enableSwitchColumn)
}

func (b notifyTunnelLargeV1Builder) setInlineMeta(in *jsoncard.LargeCoverInline) {
	in.HideDanmuSwitch = model.TunnelHideDanmuSwitch
	in.DisableDanmu = model.TunnelDisableDanmu
	uri, err := card.RawExtraURI(b.tunnel)
	if err != nil {
		log.Warn("Failed to raw extra uri: %+v", err)
		return
	}
	in.ExtraURI = uri
}

func (b notifyTunnelLargeV1Builder) Build() (*jsoncard.UniversalNotifyTunnelLargeV1, error) {
	if b.rcmd == nil {
		return nil, errors.Errorf("empty `rcmd` field")
	}
	item := &jsoncard.NotifyTunnelLargeItemV1{}
	item.FromTunnelCard(b.tunnel)

	if err := jsonbuilder.NewBaseUpdater(b.BuilderContext, b.base).
		UpdateArgs(jsoncard.Args{}).
		UpdateThreePoint(b.threePoint.ConstructDefaultThreePoint()).
		UpdateThreePointV2(b.constructThreePointV2()).Update(); err != nil {
		return nil, err
	}
	out := &jsoncard.UniversalNotifyTunnelLargeV1{
		Base: b.base,
		Item: item,
	}
	if b.rcmd.BigTunnelObject == "" {
		return out, nil
	}
	bto, err := b.bigTunnelObject()
	if err != nil {
		return nil, err
	}
	resourceID, _ := strconv.ParseInt(bto.Resource, 10, 64)
	switch bto.Type {
	case "image":
		return out, nil
	case "ugc":
		inlineAv, err := b.buildInlineAv(resourceID)
		if err != nil {
			return nil, err
		}
		out.Item.LargeCover = inlineAv.Cover
		b.setInlineMeta(inlineAv)
		embed := &jsoncard.EmbedInline{LargeCoverInline: *inlineAv}
		card.TunnelHide(embed)
		out.Item.Type = "inline_av"
		out.Item.InlineAv = embed
	case "pgc":
		inlinePGC, err := b.buildInlinePGC(resourceID)
		if err != nil {
			return nil, err
		}
		out.Item.LargeCover = inlinePGC.Cover
		b.setInlineMeta(inlinePGC)
		embed := &jsoncard.EmbedInline{LargeCoverInline: *inlinePGC}
		card.TunnelHide(embed)
		out.Item.Type = "inline_pgc"
		out.Item.InlinePGC = embed
	case "live":
		inlineLive, err := b.buildInlineLive(resourceID)
		if err != nil {
			return nil, err
		}
		out.Item.LargeCover = inlineLive.Cover
		b.setInlineMeta(inlineLive)
		embed := &jsoncard.EmbedInline{LargeCoverInline: *inlineLive}
		card.TunnelHide(embed)
		out.Item.Type = "inline_live"
		out.Item.InlineLive = embed
	default:
		return nil, errors.Errorf("unexpected big tunnel object type %s", bto.Type)
	}
	return out, nil
}

func (b notifyTunnelLargeV1Builder) buildInlineAv(id int64) (*jsoncard.LargeCoverInline, error) {
	archive, ok := b.arcPlayer[id]
	if !ok {
		return nil, errors.Errorf("archvie not exist")
	}
	base, err := jsonbuilder.NewBaseBuilder(b.BuilderContext).SetParam(strconv.FormatInt(id, 10)).
		SetMetricRcmd(b.rcmd).SetTrackID(b.rcmd.TrackID).Build()
	if err != nil {
		return nil, err
	}

	factory := jsonlargecover.NewLargeCoverInlineBuilder(b.BuilderContext)
	card, err := factory.DeriveArcPlayerBuilder().SetBase(base).SetRcmd(b.rcmd).SetArcPlayer(archive).
		SetAuthorCard(b.authorCard[archive.Arc.Author.Mid]).SetInline(b.inline).
		WithAfter(jsonlargecover.DbClickLike(b.BuilderContext, b.rcmd)).Build()
	if err != nil {
		return nil, err
	}
	return card, nil
}

func (b notifyTunnelLargeV1Builder) buildInlinePGC(id int64) (*jsoncard.LargeCoverInline, error) {
	epCard, ok := b.episode[int32(id)]
	if !ok {
		return nil, errors.Errorf("ep card not exist")
	}
	base, err := jsonbuilder.NewBaseBuilder(b.BuilderContext).SetParam(strconv.FormatInt(id, 10)).
		SetMetricRcmd(b.rcmd).SetTrackID(b.rcmd.TrackID).Build()
	if err != nil {
		return nil, err
	}
	factory := jsonlargecover.NewLargeCoverInlineBuilder(b.BuilderContext)
	card, err := factory.DerivePgcBuilder().SetBase(base).SetRcmd(b.rcmd).SetEpisode(epCard).SetInline(b.inline).
		WithAfter(jsonlargecover.DbClickLike(b.BuilderContext, b.rcmd)).Build()
	if err != nil {
		return nil, err
	}
	return card, nil
}

func (b notifyTunnelLargeV1Builder) buildInlineLive(id int64) (*jsoncard.LargeCoverInline, error) {
	live, ok := b.liveRoom[id]
	if !ok {
		return nil, errors.Errorf("live room not exist")
	}
	base, err := jsonbuilder.NewBaseBuilder(b.BuilderContext).SetParam(strconv.FormatInt(id, 10)).
		SetMetricRcmd(b.rcmd).SetTrackID(b.rcmd.TrackID).Build()
	if err != nil {
		return nil, err
	}
	factory := jsonlargecover.NewLargeCoverInlineBuilder(b.BuilderContext)
	card, err := factory.DeriveLiveRoomBuilder().SetBase(base).SetRcmd(b.rcmd).SetLiveRoom(live).SetInline(b.inline).
		SetAuthorCard(b.authorCard[live.UID]).Build()
	if err != nil {
		return nil, err
	}
	return card, nil
}
