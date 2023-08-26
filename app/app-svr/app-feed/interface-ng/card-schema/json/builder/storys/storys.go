package storys

import (
	"strconv"

	"go-common/library/log"
	appcardmodel "go-gateway/app/app-svr/app-card/interface/model"
	"go-gateway/app/app-svr/app-card/interface/model/card"
	"go-gateway/app/app-svr/app-card/interface/model/card/ai"
	cardschema "go-gateway/app/app-svr/app-feed/interface-ng/card-schema"
	jsoncard "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json"
	jsonbuilder "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder"
	jsonavatar "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder/avatar"
	jsoncommon "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder/common"
	"go-gateway/app/app-svr/app-feed/interface/common"
	arcgrpc "go-gateway/app/app-svr/archive/service/api"

	accountgrpc "git.bilibili.co/bapis/bapis-go/account/service"
	taggrpc "git.bilibili.co/bapis/bapis-go/community/interface/tag"
	"github.com/pkg/errors"
)

type StorysBuilder interface {
	ReplaceContext(jsonbuilder.BuilderContext) StorysBuilder
	SetBase(*jsoncard.Base) StorysBuilder
	SetRcmd(*ai.Item) StorysBuilder
	SetArcPlayer(map[int64]*arcgrpc.ArcPlayer) StorysBuilder
	SetTags(map[int64]*taggrpc.Tag) StorysBuilder
	SetAuthorCard(map[int64]*accountgrpc.Card) StorysBuilder
	Build() (*jsoncard.Storys, error)
}

type storysBuilder struct {
	jsonbuilder.BuilderContext
	archvieCommon jsoncommon.ArchiveCommon
	rcmd          *ai.Item
	base          *jsoncard.Base
	arcPlayer     map[int64]*arcgrpc.ArcPlayer
	tags          map[int64]*taggrpc.Tag
	authorCard    map[int64]*accountgrpc.Card
	threePoint    jsoncommon.ThreePoint
}

func NewStorysBuilder(ctx jsonbuilder.BuilderContext) StorysBuilder {
	return storysBuilder{BuilderContext: ctx}
}

func (b storysBuilder) ReplaceContext(ctx jsonbuilder.BuilderContext) StorysBuilder {
	b.BuilderContext = ctx
	return b
}

func (b storysBuilder) SetBase(base *jsoncard.Base) StorysBuilder {
	b.base = base
	return b
}

func (b storysBuilder) SetRcmd(item *ai.Item) StorysBuilder {
	b.rcmd = item
	return b
}

func (b storysBuilder) SetArcPlayer(in map[int64]*arcgrpc.ArcPlayer) StorysBuilder {
	b.arcPlayer = in
	return b
}

func (b storysBuilder) SetTags(in map[int64]*taggrpc.Tag) StorysBuilder {
	b.tags = in
	return b
}

func (b storysBuilder) SetAuthorCard(in map[int64]*accountgrpc.Card) StorysBuilder {
	b.authorCard = in
	return b
}

func (b storysBuilder) Build() (*jsoncard.Storys, error) {
	if len(b.arcPlayer) == 0 {
		return nil, errors.Errorf("empty `arcPlayer` field")
	}
	if b.rcmd == nil {
		return nil, errors.Errorf("empty `rcmd` field")
	}
	if b.rcmd.StoryInfo == nil {
		return nil, errors.Errorf("empty `StoryInfo` field")
	}
	if len(b.rcmd.StoryInfo.Items) == 0 {
		return nil, errors.Errorf("empty `StoryInfo Items` field")
	}
	if err := jsonbuilder.NewBaseUpdater(b.BuilderContext, b.base).
		UpdateTitle(b.rcmd.StoryInfo.Title).
		UpdateArgs(jsoncard.Args{}).
		Update(); err != nil {
		return nil, err
	}
	out := &jsoncard.Storys{
		Base:  b.base,
		Items: make([]*card.StoryItems, 0, len(b.rcmd.StoryInfo.Items)),
	}
	for _, v := range b.rcmd.StoryInfo.Items {
		arc, ok := b.arcPlayer[v.ID]
		if !ok {
			log.Error("storys arc is not ok: %d", v.ID)
			continue
		}
		if !appcardmodel.AvIsNormalGRPC(arc) {
			log.Error("storys arc is not normal: %+v", arc)
			continue
		}
		item := &card.StoryItems{
			FfCover:        common.Ffcover(arc.Arc.FirstFrame, appcardmodel.FfCoverFromFeed),
			OfficialIcon:   appcardmodel.OfficialIcon(b.authorCard[arc.Arc.Author.Mid]),
			OfficialIconV2: appcardmodel.OfficialIcon(b.authorCard[arc.Arc.Author.Mid]),
			CoverLeftText1: appcardmodel.StatString(arc.Arc.Stat.View, ""),
			CoverLeftIcon1: appcardmodel.IconPlay,
		}
		avatar, err := jsonavatar.NewAvatarBuilder(b.BuilderContext).
			SetAvatarStatus(&jsoncard.AvatarStatus{
				Cover: arc.Arc.Author.Face,
				Text:  arc.Arc.Author.Name,
				Goto:  appcardmodel.GotoMid,
				Param: strconv.FormatInt(arc.Arc.Author.Mid, 10),
				Type:  appcardmodel.AvatarRound,
			}).Build()
		if err != nil {
			log.Error("Failed to build avatar: %+v", err)
		}
		item.Avatar = avatar
		base := &jsoncard.Base{
			CardGoto: appcardmodel.CardGotoVerticalAv,
			Goto:     appcardmodel.GotoVerticalAv,
		}
		args := b.archvieCommon.ConstructArgs(arc, b.tags[v.Tid])
		device := b.BuilderContext.Device()
		enableSwitchColumn := b.BuilderContext.FeatureGates().FeatureEnabled(cardschema.FeatureSwitchColumnThreePoint)
		if err := jsonbuilder.NewBaseUpdater(b.BuilderContext, base).
			UpdateURI(b.archvieCommon.ConstructVerticalArchiveURI(arc.Arc.Aid, device, appcardmodel.ArcPlayHandler(arc.Arc,
				appcardmodel.ArcPlayURL(arc, 0), b.rcmd.TrackID, b.rcmd, int(device.Build()), device.RawMobiApp(), true))).
			UpdateCover(v.Cover).
			UpdateTitle(arc.Arc.Title).
			UpdateParam(strconv.FormatInt(v.ID, 10)).
			UpdateArgs(args).
			UpdatePlayerArgs(b.archvieCommon.ConstructPlayerArgs(arc)).
			UpdateThreePointV2(b.threePoint.ConstructArchvieThreePointV2(b.BuilderContext, &args,
				jsoncommon.WatchLater(false),
				jsoncommon.SwitchColumn(enableSwitchColumn),
				jsoncommon.AvDislikeInfo(0),
				jsoncommon.Item(b.rcmd))).
			UpdateThreePoint(nil).
			Update(); err != nil {
			log.Error("Failed to update base: %+v", err)
			continue
		}
		if b.BuilderContext.IsAttentionTo(arc.Arc.Author.Mid) {
			item.OfficialIcon = appcardmodel.IconIsAttenm
			item.IsAtten = true
		}
		item.Base = *base
		out.Items = append(out.Items, item)
	}
	return out, nil
}
