package jsonsmallcover

import (
	appcardmodel "go-gateway/app/app-svr/app-card/interface/model"
	"go-gateway/app/app-svr/app-card/interface/model/bplus"
	"go-gateway/app/app-svr/app-card/interface/model/card/ai"
	jsoncard "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json"
	jsonbuilder "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder"
	jsoncommon "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder/common"
	jsonreasonstyle "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder/reason_style"

	"github.com/pkg/errors"
)

//go:generate python3 ../../../../contrib/desc-button-overlapped.py

type V2PictureBuilder interface {
	Parent() SmallCoverV2BuilderFactory
	SetBase(*jsoncard.Base) V2PictureBuilder
	SetPicture(*bplus.Picture) V2PictureBuilder
	SetRcmd(*ai.Item) V2PictureBuilder

	Build() (*jsoncard.SmallCoverV2, error)
	WithAfter(req ...func(*jsoncard.SmallCoverV2)) V2PictureBuilder
}

type v2PictureBuilder struct {
	parent *smallCoverV2BuilderFactory
	jsoncommon.PicCommon
	rcmd    *ai.Item
	base    *jsoncard.Base
	picture *bplus.Picture
	afterFn []func(*jsoncard.SmallCoverV2)
}

func (b v2PictureBuilder) Parent() SmallCoverV2BuilderFactory {
	return b.parent
}

func (b v2PictureBuilder) SetBase(base *jsoncard.Base) V2PictureBuilder {
	b.base = base
	return b
}

func (b v2PictureBuilder) SetPicture(picture *bplus.Picture) V2PictureBuilder {
	b.picture = picture
	return b
}

func (b v2PictureBuilder) SetRcmd(item *ai.Item) V2PictureBuilder {
	b.rcmd = item
	return b
}

func (b v2PictureBuilder) constructURI() string {
	return b.ConstructPictureURI(b.picture.DynamicID)
}

func (b v2PictureBuilder) constructArgs() jsoncard.Args {
	return b.ConstructArgsFromPicture(b.picture)
}

func (b v2PictureBuilder) constructDescButton() *jsoncard.Button {
	return b.ConstructDescButtonFromPicture(b.picture)
}

func (b v2PictureBuilder) constructThreePoint() *jsoncard.ThreePoint {
	return b.ConstructThreePointFromPicture(b.picture)
}

func (b v2PictureBuilder) constructThreePointV2() []*jsoncard.ThreePointV2 {
	return b.ConstructThreePointV2FromPicture(b.picture)
}

func (b v2PictureBuilder) usingLikeText() bool {
	return b.parent.BuilderContext.VersionControl().Can("pic.usingLikeText")
}

func (b v2PictureBuilder) Build() (*jsoncard.SmallCoverV2, error) {
	if b.base == nil {
		return nil, errors.Errorf("empty `base` field")
	}
	if b.picture == nil {
		return nil, errors.Errorf("empty `picture` field")
	}
	if b.rcmd == nil {
		return nil, errors.Errorf("empty `rcmd` field")
	}
	if len(b.picture.Imgs) <= 0 {
		return nil, errors.Errorf("empty `img` field in picture: %+v", b.picture)
	}
	if b.picture.ViewCount <= 0 {
		return nil, errors.Errorf("empty viewCount: %+v", b.picture)
	}
	if err := jsonbuilder.NewBaseUpdater(b.parent.BuilderContext, b.base).
		UpdateCover(b.picture.Imgs[0]).
		UpdateTitle(b.picture.DynamicText).
		UpdateGoto(appcardmodel.GotoPicture).
		UpdateURI(b.constructURI()).
		UpdateArgs(b.constructArgs()).
		UpdateThreePoint(b.constructThreePoint()).
		UpdateThreePointV2(b.constructThreePointV2()).
		Update(); err != nil {
		return nil, err
	}
	out := &jsoncard.SmallCoverV2{
		Base:       b.base,
		Badge:      "动态",
		BadgeStyle: jsonreasonstyle.ConstructReasonStyle(appcardmodel.BgColorTransparentRed, "动态"),
		DescButton: b.constructDescButton(),
	}
	out.CoverLeftText1 = appcardmodel.StatString(int32(b.picture.ViewCount), "")
	out.CoverLeftIcon1 = appcardmodel.IconRead
	if b.usingLikeText() {
		out.CoverLeftText1 = appcardmodel.StatString(b.picture.LikeCount, "")
		out.CoverLeftIcon1 = appcardmodel.IconLike
	}
	out.CoverLeft1ContentDescription = appcardmodel.CoverIconContentDescription(out.CoverLeftIcon1,
		out.CoverLeftText1)
	if b.picture.ImgCount > 1 {
		out.CoverRightText = appcardmodel.PictureCountString(b.picture.ImgCount)
		out.CoverRightContentDescription = appcardmodel.PictureCountString(b.picture.ImgCount)
		out.CoverRightBackgroundColor = "#66666666"
	}
	if b.rcmd != nil && b.rcmd.RcmdReason != nil {
		out.Badge = ""
		out.BadgeStyle = nil
		out.DescButton = nil
		reasonText, _ := jsonreasonstyle.BuildRecommendReasonText(
			b.parent.BuilderContext,
			b.rcmd.RcmdReason,
			b.rcmd.Goto,
			b.picture.NickName,
			b.parent.BuilderContext.IsAttentionTo(b.picture.Mid),
		)
		out.RcmdReasonStyle = jsonreasonstyle.ConstructTopReasonStyle(reasonText,
			jsonreasonstyle.CornerMarkFromAI(b.rcmd),
			jsonreasonstyle.CorverMarkFromContext(b.parent.BuilderContext))
		out.RcmdReason = reasonText
	}
	for _, fn := range b.afterFn {
		fn(out)
	}
	return out, nil
}

func (b v2PictureBuilder) WithAfter(req ...func(*jsoncard.SmallCoverV2)) V2PictureBuilder {
	b.afterFn = append(b.afterFn, req...)
	return b
}
