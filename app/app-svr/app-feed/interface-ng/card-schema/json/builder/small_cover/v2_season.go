package jsonsmallcover

import (
	"fmt"
	appcardmodel "go-gateway/app/app-svr/app-card/interface/model"
	"go-gateway/app/app-svr/app-card/interface/model/card/bangumi"
	jsoncard "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json"
	jsonbuilder "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder"
	jsoncommon "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder/common"
	jsonreasonstyle "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder/reason_style"

	taggrpc "git.bilibili.co/bapis/bapis-go/community/interface/tag"
	"github.com/pkg/errors"
)

//go:generate python3 ../../../../contrib/desc-button-overlapped.py

type V2BangumiSeasonBuilder interface {
	Parent() SmallCoverV2BuilderFactory
	SetBase(*jsoncard.Base) V2BangumiSeasonBuilder
	SetBangumiSeason(*bangumi.Season) V2BangumiSeasonBuilder
	SetTag(*taggrpc.Tag) V2BangumiSeasonBuilder
	SetTypeName(string) V2BangumiSeasonBuilder

	Build() (*jsoncard.SmallCoverV2, error)
	WithAfter(req ...func(*jsoncard.SmallCoverV2)) V2BangumiSeasonBuilder
}

type v2BangumiSeasonBuilder struct {
	seasonCommon jsoncommon.BangumiSeason
	base         *jsoncard.Base
	parent       *smallCoverV2BuilderFactory
	season       *bangumi.Season
	tag          *taggrpc.Tag
	typeName     string
	afterFn      []func(*jsoncard.SmallCoverV2)
}

func (b v2BangumiSeasonBuilder) Parent() SmallCoverV2BuilderFactory {
	return b.parent
}

func (b v2BangumiSeasonBuilder) SetBase(base *jsoncard.Base) V2BangumiSeasonBuilder {
	b.base = base
	return b
}

func (b v2BangumiSeasonBuilder) SetBangumiSeason(in *bangumi.Season) V2BangumiSeasonBuilder {
	b.season = in
	return b
}

func (b v2BangumiSeasonBuilder) SetTag(in *taggrpc.Tag) V2BangumiSeasonBuilder {
	b.tag = in
	return b
}

func (b v2BangumiSeasonBuilder) SetTypeName(in string) V2BangumiSeasonBuilder {
	b.typeName = in
	return b
}

func (b v2BangumiSeasonBuilder) constructDescButton() *jsoncard.Button {
	if b.tag == nil {
		return nil
	}
	if b.tag.Name == "" || b.typeName == "" {
		return nil
	}
	tagDup := &taggrpc.Tag{}
	*tagDup = *b.tag
	tagDup.Name = fmt.Sprintf("%s . %s", b.typeName, tagDup.Name)
	return b.seasonCommon.ConstructDescButtonFromTag(b.tag)
}

func (b v2BangumiSeasonBuilder) constructURI() string {
	return b.seasonCommon.ConstructSeasonURI(b.season.EpisodeID)
}

func (b v2BangumiSeasonBuilder) Build() (*jsoncard.SmallCoverV2, error) {
	if b.season == nil {
		return nil, errors.Errorf("empty `season` field")
	}
	if b.base.Rcmd == nil {
		return nil, errors.Errorf("empty `rcmd` field in base")
	}

	baseUpdater := jsonbuilder.NewBaseUpdater(b.parent.BuilderContext, b.base).
		UpdateURI(b.constructURI()).
		UpdateParam(b.season.EpisodeID).
		UpdateCover(b.season.Cover).
		UpdateTitle(b.season.Title).
		UpdateGoto(appcardmodel.GotoBangumi)

	out := &jsoncard.SmallCoverV2{
		Base: b.base,
	}
	out.CoverLeftText1 = appcardmodel.StatString(b.season.PlayCount, "")
	out.CoverLeftIcon1 = appcardmodel.IconPlay
	out.CoverLeftText2 = appcardmodel.StatString(b.season.Favorites, "")
	out.CoverLeftIcon2 = appcardmodel.IconFavorite
	out.Badge = b.season.TypeBadge
	out.BadgeStyle = jsonreasonstyle.ConstructReasonStyle(appcardmodel.BgColorTransparentRed, b.season.TypeBadge)
	out.Subtitle = b.season.UpdateDesc
	out.DescButton = b.constructDescButton()

	rcmd := b.base.Rcmd
	if rcmd.RcmdReason != nil {
		out.RcmdReasonStyle = jsonreasonstyle.ConstructTopReasonStyle(rcmd.RcmdReason.Content,
			jsonreasonstyle.CornerMarkFromAI(rcmd))
		return out, nil
	}
	if err := baseUpdater.Update(); err != nil {
		return nil, err
	}
	for _, fn := range b.afterFn {
		fn(out)
	}
	return out, nil
}

func (b v2BangumiSeasonBuilder) WithAfter(req ...func(*jsoncard.SmallCoverV2)) V2BangumiSeasonBuilder {
	b.afterFn = append(b.afterFn, req...)
	return b
}
