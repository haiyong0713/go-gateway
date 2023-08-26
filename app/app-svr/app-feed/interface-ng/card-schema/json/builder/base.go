package jsonbuilder

import (
	appcardmodel "go-gateway/app/app-svr/app-card/interface/model"
	"go-gateway/app/app-svr/app-card/interface/model/card/ai"
	"go-gateway/app/app-svr/app-card/interface/model/card/cm"
	"go-gateway/app/app-svr/app-card/interface/model/card/report"
	"go-gateway/app/app-svr/app-card/interface/model/card/threePointMeta"
	jsoncard "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json"

	"github.com/pkg/errors"
)

type BaseBuilder interface {
	ReplaceContext(BuilderContext) BaseBuilder
	SetIndex(int64) BaseBuilder
	SetCardType(appcardmodel.CardType) BaseBuilder
	SetGoto(appcardmodel.Gt) BaseBuilder
	SetCardGoto(appcardmodel.CardGt) BaseBuilder
	SetCardLen(int64) BaseBuilder
	SetParam(string) BaseBuilder
	SetMetricRcmd(*ai.Item) BaseBuilder
	SetTrackID(string) BaseBuilder
	SetPosRecUniqueID(string) BaseBuilder
	SetAdInfo(*cm.AdInfo) BaseBuilder
	SetCreativeId(int64) BaseBuilder
	Build() (*jsoncard.Base, error)
}

type BaseUpdater interface {
	ReplaceContext(BuilderContext) BaseUpdater
	UpdateIndex(int64) BaseUpdater
	UpdateParam(string) BaseUpdater
	UpdateCover(string) BaseUpdater
	UpdateTitle(string) BaseUpdater
	UpdateGoto(appcardmodel.Gt) BaseUpdater
	UpdateCardGoto(appcardmodel.CardGt) BaseUpdater
	UpdateURI(string) BaseUpdater
	UpdateArgs(jsoncard.Args) BaseUpdater
	UpdatePlayerArgs(*jsoncard.PlayerArgs) BaseUpdater
	UpdateBaseInnerDescButton(*jsoncard.Button) BaseUpdater
	UpdateAdInfo(*cm.AdInfo) BaseUpdater
	UpdateThreePoint(*jsoncard.ThreePoint) BaseUpdater
	UpdateThreePointV2([]*jsoncard.ThreePointV2) BaseUpdater
	UpdateUpArgs(*jsoncard.UpArgs) BaseUpdater
	UpdateMask(*jsoncard.Mask) BaseUpdater
	UpdateBvid(string) BaseUpdater
	UpdateThreePointMeta(*threePointMeta.PanelMeta) BaseUpdater
	UpdateCmInfo(*jsoncard.CmInfo) BaseUpdater

	Update() error
}

type baseBuilder struct {
	BuilderContext
	cardType       appcardmodel.CardType
	index          int64
	param          string
	cover          string
	title          string
	goto_          appcardmodel.Gt
	cardGoto       appcardmodel.CardGt
	uri            string
	args           jsoncard.Args
	descButton     *jsoncard.Button
	cardLen        int64
	rcmd           *ai.Item
	trackID        string
	posRecUniqueID string
	adInfo         *cm.AdInfo
	creativeId     int64
}

func NewBaseBuilder(ctx BuilderContext) BaseBuilder {
	return baseBuilder{BuilderContext: ctx}
}

func (b baseBuilder) ReplaceContext(ctx BuilderContext) BaseBuilder {
	b.BuilderContext = ctx
	return b
}
func (b baseBuilder) SetCardType(in appcardmodel.CardType) BaseBuilder {
	b.cardType = in
	return b
}
func (b baseBuilder) SetIndex(in int64) BaseBuilder {
	b.index = in
	return b
}
func (b baseBuilder) SetParam(in string) BaseBuilder {
	b.param = in
	return b
}
func (b baseBuilder) SetCover(in string) BaseBuilder {
	b.cover = in
	return b
}
func (b baseBuilder) SetTitle(in string) BaseBuilder {
	b.title = in
	return b
}
func (b baseBuilder) SetGoto(in appcardmodel.Gt) BaseBuilder {
	b.goto_ = in
	return b
}
func (b baseBuilder) SetCardGoto(in appcardmodel.CardGt) BaseBuilder {
	b.cardGoto = in
	return b
}
func (b baseBuilder) SetURI(in string) BaseBuilder {
	b.uri = in
	return b
}
func (b baseBuilder) SetArgs(in jsoncard.Args) BaseBuilder {
	b.args = in
	return b
}
func (b baseBuilder) SetDescButton(in *jsoncard.Button) BaseBuilder {
	b.descButton = in
	return b
}

func (b baseBuilder) SetCardLen(in int64) BaseBuilder {
	b.cardLen = in
	return b
}

func (b baseBuilder) SetMetricRcmd(in *ai.Item) BaseBuilder {
	b.rcmd = in
	return b
}

func (b baseBuilder) SetTrackID(in string) BaseBuilder {
	b.trackID = in
	return b
}

func (b baseBuilder) SetPosRecUniqueID(in string) BaseBuilder {
	b.posRecUniqueID = in
	return b
}

func (b baseBuilder) SetAdInfo(in *cm.AdInfo) BaseBuilder {
	b.adInfo = in
	return b
}

func (b baseBuilder) SetCreativeId(in int64) BaseBuilder {
	b.creativeId = in
	return b
}

func (b baseBuilder) Build() (*jsoncard.Base, error) {
	out := &jsoncard.Base{}
	out.Idx = b.index
	out.CardType = b.cardType
	out.URI = b.uri
	out.Cover = b.cover
	out.Title = b.title
	out.CardGoto = b.cardGoto
	if b.goto_ != "" {
		out.Goto = b.goto_
	}
	out.Param = b.param
	out.Args = b.args
	out.DescButton = b.descButton
	out.CardLen = int(b.cardLen)
	if b.rcmd == nil {
		return nil, errors.Errorf("Failed to build base, the rcmd is nil")
	}
	out.Rcmd = b.rcmd
	out.TrackID = b.trackID
	out.PosRecUniqueID = b.posRecUniqueID
	out.AdInfo = b.adInfo
	out.OgvCreativeId = b.creativeId
	out.MaterialId = b.creativeId
	out.DislikeReportData = report.BuildDislikeReportData(b.creativeId, b.posRecUniqueID)
	return out, nil
}

type baseUpdater struct {
	BuilderContext
	base *jsoncard.Base

	param          *string
	index          *int64
	cover          *string
	title          *string
	goto_          *appcardmodel.Gt
	cardGoto       *appcardmodel.CardGt
	uri            *string
	args           *jsoncard.Args
	playerArgs     **jsoncard.PlayerArgs
	descButton     **jsoncard.Button
	adInfo         **cm.AdInfo
	threePoint     **jsoncard.ThreePoint
	threePointV2   *[]*jsoncard.ThreePointV2
	upArgs         **jsoncard.UpArgs
	mask           **jsoncard.Mask
	bvid           *string
	threePointMeta **threePointMeta.PanelMeta
	cmInfo         **jsoncard.CmInfo
}

func NewBaseUpdater(ctx BuilderContext, base *jsoncard.Base) BaseUpdater {
	return baseUpdater{BuilderContext: ctx, base: base}
}
func (b baseUpdater) ReplaceContext(ctx BuilderContext) BaseUpdater {
	b.BuilderContext = ctx
	return b
}
func (b baseUpdater) UpdateIndex(in int64) BaseUpdater {
	b.index = &in
	return b
}
func (b baseUpdater) UpdateParam(in string) BaseUpdater {
	b.param = &in
	return b
}
func (b baseUpdater) UpdateCover(in string) BaseUpdater {
	b.cover = &in
	return b
}
func (b baseUpdater) UpdateTitle(in string) BaseUpdater {
	b.title = &in
	return b
}
func (b baseUpdater) UpdateGoto(in appcardmodel.Gt) BaseUpdater {
	b.goto_ = &in
	return b
}
func (b baseUpdater) UpdateCardGoto(in appcardmodel.CardGt) BaseUpdater {
	b.cardGoto = &in
	return b
}
func (b baseUpdater) UpdateURI(in string) BaseUpdater {
	b.uri = &in
	return b
}
func (b baseUpdater) UpdateArgs(in jsoncard.Args) BaseUpdater {
	b.args = &in
	return b
}
func (b baseUpdater) UpdatePlayerArgs(in *jsoncard.PlayerArgs) BaseUpdater {
	b.playerArgs = &in
	return b
}
func (b baseUpdater) UpdateBaseInnerDescButton(in *jsoncard.Button) BaseUpdater {
	b.descButton = &in
	return b
}
func (b baseUpdater) UpdateAdInfo(in *cm.AdInfo) BaseUpdater {
	b.adInfo = &in
	return b
}
func (b baseUpdater) UpdateThreePoint(in *jsoncard.ThreePoint) BaseUpdater {
	b.threePoint = &in
	return b
}
func (b baseUpdater) UpdateThreePointV2(in []*jsoncard.ThreePointV2) BaseUpdater {
	b.threePointV2 = &in
	return b
}

func (b baseUpdater) UpdateUpArgs(in *jsoncard.UpArgs) BaseUpdater {
	b.upArgs = &in
	return b
}

func (b baseUpdater) UpdateMask(in *jsoncard.Mask) BaseUpdater {
	b.mask = &in
	return b
}

func (b baseUpdater) UpdateBvid(in string) BaseUpdater {
	b.bvid = &in
	return b
}

func (b baseUpdater) UpdateThreePointMeta(in *threePointMeta.PanelMeta) BaseUpdater {
	b.threePointMeta = &in
	return b
}

func (b baseUpdater) UpdateCmInfo(in *jsoncard.CmInfo) BaseUpdater {
	b.cmInfo = &in
	return b
}

func (b baseUpdater) Update() error {
	if b.base == nil {
		return errors.Errorf("empty `base` field")
	}

	if b.param != nil {
		b.base.Param = *b.param
	}
	if b.index != nil {
		b.base.Idx = *b.index
	}
	if b.cover != nil {
		b.base.Cover = *b.cover
	}
	if b.title != nil {
		b.base.Title = *b.title
	}
	if b.cardGoto != nil {
		b.base.CardGoto = appcardmodel.CardGt(*b.cardGoto)
	}
	if _goto := gtValue(b.goto_); string(_goto) != "" {
		b.base.Goto = _goto
	}
	if b.uri != nil {
		b.base.URI = *b.uri
	}
	if b.args != nil {
		b.base.Args = *b.args
	}
	if b.playerArgs != nil {
		b.base.PlayerArgs = *b.playerArgs
	}
	if b.descButton != nil {
		b.base.DescButton = *b.descButton
	}
	if b.adInfo != nil {
		b.base.AdInfo = *b.adInfo
	}
	if b.threePoint != nil {
		b.base.ThreePoint = *b.threePoint
	}
	if b.threePointV2 != nil {
		b.base.ThreePointV2 = *b.threePointV2
	}
	if b.upArgs != nil {
		b.base.UpArgs = *b.upArgs
	}
	if b.mask != nil {
		b.base.Mask = *b.mask
	}
	if b.bvid != nil {
		b.base.Bvid = *b.bvid
	}
	if b.threePointMeta != nil {
		b.base.ThreePointMeta = *b.threePointMeta
	}
	if b.cmInfo != nil {
		b.base.CmInfo = *b.cmInfo
	}
	return nil
}

func gtValue(v *appcardmodel.Gt) appcardmodel.Gt {
	if v != nil {
		return *v
	}
	return appcardmodel.Gt("")
}
