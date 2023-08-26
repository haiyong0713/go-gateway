package jsonthreepic

import (
	tunnelV2 "git.bilibili.co/bapis/bapis-go/ai/feed/mgr/service"
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

type ThreePicV3Builder interface {
	ReplaceContext(jsonbuilder.BuilderContext) ThreePicV3Builder
	SetBase(*jsoncard.Base) ThreePicV3Builder
	SetPicture(*bplus.Picture) ThreePicV3Builder
	SetRcmd(*ai.Item) ThreePicV3Builder

	Build() (*jsoncard.ThreePicV3, error)
	WithAfter(req ...func(*jsoncard.ThreePicV3)) ThreePicV3Builder
}

type threePicV3Builder struct {
	jsoncommon.PicCommon
	jsonbuilder.BuilderContext
	base    *jsoncard.Base
	picture *bplus.Picture
	rcmd    *ai.Item
	afterFn []func(*jsoncard.ThreePicV3)
}

func NewThreePicV3Builder(ctx jsonbuilder.BuilderContext) ThreePicV3Builder {
	return threePicV3Builder{BuilderContext: ctx}
}

func (b threePicV3Builder) ReplaceContext(ctx jsonbuilder.BuilderContext) ThreePicV3Builder {
	b.BuilderContext = ctx
	return b
}

func (b threePicV3Builder) SetBase(base *jsoncard.Base) ThreePicV3Builder {
	b.base = base
	return b
}

func (b threePicV3Builder) SetPicture(picture *bplus.Picture) ThreePicV3Builder {
	b.picture = picture
	return b
}

func (b threePicV3Builder) SetRcmd(rcmd *ai.Item) ThreePicV3Builder {
	b.rcmd = rcmd
	return b
}

func (b threePicV3Builder) constructArgs() jsoncard.Args {
	return b.ConstructArgsFromPicture(b.picture)
}

func (b threePicV3Builder) constructURI() string {
	return b.ConstructPictureURI(b.picture.DynamicID)
}

func (b threePicV3Builder) constructDescButton() *jsoncard.Button {
	return b.ConstructDescButtonFromPicture(b.picture)
}

func (b threePicV3Builder) constructThreePoint() *jsoncard.ThreePoint {
	return b.ConstructThreePointFromPicture(b.picture)
}

func (b threePicV3Builder) constructThreePointV2() []*jsoncard.ThreePointV2 {
	return b.ConstructThreePointV2FromPicture(b.picture)
}

func (b threePicV3Builder) activitySupported() bool {
	return b.VersionControl().Can("pic.activitySupported")
}

func (b threePicV3Builder) usingLikeText() bool {
	return b.BuilderContext.VersionControl().Can("pic.usingLikeText")
}

func (b threePicV3Builder) Build() (*jsoncard.ThreePicV3, error) {
	if b.base == nil {
		return nil, errors.Errorf("empty `base` field")
	}
	if b.rcmd == nil {
		return nil, errors.Errorf("empty `rcmd` field")
	}
	if b.picture == nil {
		return nil, errors.Errorf("empty `picture` field")
	}
	//nolint:gomnd
	if len(b.picture.Imgs) < 3 {
		return nil, errors.Errorf("insufficient `imgs` field in picture: %+v", b.picture)
	}
	if b.picture.ViewCount == 0 {
		return nil, errors.Errorf("invalid `view_count` in picture: %+v", b.picture)
	}

	if err := jsonbuilder.NewBaseUpdater(b.BuilderContext, b.base).
		UpdateCover("").
		UpdateTitle(b.picture.DynamicText).
		UpdateGoto(appcardmodel.GotoPicture).
		UpdateURI(b.constructURI()).
		UpdateArgs(b.constructArgs()).
		UpdateThreePoint(b.constructThreePoint()).
		UpdateThreePointV2(b.constructThreePointV2()).
		Update(); err != nil {
		return nil, err
	}
	out := &jsoncard.ThreePicV3{
		Base:       b.base,
		Badge:      "动态",
		DescButton: b.constructDescButton(),
	}
	if b.picture.JumpUrl != "" && b.activitySupported() {
		out.URI = b.picture.JumpUrl
		out.Badge = "活动"
	}
	out.LeftCover = b.picture.Imgs[0]
	out.RightCover1 = b.picture.Imgs[1]
	out.RightCover2 = b.picture.Imgs[2]
	out.CoverLeftText1 = appcardmodel.StatString(int32(b.picture.ViewCount), "")
	out.CoverLeftIcon1 = appcardmodel.IconRead
	if b.usingLikeText() {
		out.CoverLeftText1 = appcardmodel.StatString(b.picture.LikeCount, "")
		out.CoverLeftIcon1 = appcardmodel.IconLike
	}
	out.CoverRightText = b.picture.NickName
	if b.rcmd.RcmdReason != nil {
		out.DescButton = nil
		reasonText, _ := jsonreasonstyle.BuildRecommendReasonText(
			b.BuilderContext,
			b.rcmd.RcmdReason,
			b.rcmd.Goto,
			b.picture.NickName,
			b.IsAttentionTo(b.picture.Mid),
		)
		out.RcmdReasonStyle = jsonreasonstyle.ConstructTopReasonStyle(reasonText,
			jsonreasonstyle.CornerMarkFromAI(b.rcmd),
			jsonreasonstyle.CorverMarkFromContext(b.BuilderContext))
	}
	out.BadgeStyle = jsonreasonstyle.ConstructReasonStyle(appcardmodel.BgColorTransparentRed, out.Badge)
	for _, fn := range b.afterFn {
		fn(out)
	}
	return out, nil
}

func (b threePicV3Builder) WithAfter(req ...func(*jsoncard.ThreePicV3)) ThreePicV3Builder {
	b.afterFn = append(b.afterFn, req...)
	return b
}

func ThreePicV3ByMultiMaterials(arg *tunnelV2.Material) func(*jsoncard.ThreePicV3) {
	return func(card *jsoncard.ThreePicV3) {
		if arg == nil {
			return
		}
		if arg.Title != "" {
			card.Title = arg.Title
		}
		if arg.Cover != "" {
			card.Cover = arg.Cover
		}
	}
}
