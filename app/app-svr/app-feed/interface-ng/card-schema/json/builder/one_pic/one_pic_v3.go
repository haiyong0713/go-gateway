package jsononepic

import (
	appcardmodel "go-gateway/app/app-svr/app-card/interface/model"
	"go-gateway/app/app-svr/app-card/interface/model/bplus"
	"go-gateway/app/app-svr/app-card/interface/model/card/ai"
	jsoncard "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json"
	jsonbuilder "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder"
	jsoncommon "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder/common"
	jsonreasonstyle "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder/reason_style"

	tunnelV2 "git.bilibili.co/bapis/bapis-go/ai/feed/mgr/service"
	"github.com/pkg/errors"
)

type OnePicV3Builder interface {
	ReplaceContext(jsonbuilder.BuilderContext) OnePicV3Builder
	SetBase(*jsoncard.Base) OnePicV3Builder
	SetPicture(*bplus.Picture) OnePicV3Builder
	SetRcmd(*ai.Item) OnePicV3Builder

	Build() (*jsoncard.OnePicV3, error)
	WithAfter(req ...func(*jsoncard.OnePicV3)) OnePicV3Builder
}

type onePicV3Builder struct {
	jsoncommon.PicCommon
	jsonbuilder.BuilderContext
	rcmd    *ai.Item
	base    *jsoncard.Base
	picture *bplus.Picture
	afterFn []func(*jsoncard.OnePicV3)
}

func NewOnePicV3Builder(ctx jsonbuilder.BuilderContext) OnePicV3Builder {
	return onePicV3Builder{BuilderContext: ctx}
}

func (b onePicV3Builder) ReplaceContext(ctx jsonbuilder.BuilderContext) OnePicV3Builder {
	b.BuilderContext = ctx
	return b
}

func (b onePicV3Builder) SetBase(base *jsoncard.Base) OnePicV3Builder {
	b.base = base
	return b
}

func (b onePicV3Builder) SetPicture(picture *bplus.Picture) OnePicV3Builder {
	b.picture = picture
	return b
}

func (b onePicV3Builder) SetRcmd(rcmd *ai.Item) OnePicV3Builder {
	b.rcmd = rcmd
	return b
}

func (b onePicV3Builder) constructURI() string {
	return b.ConstructPictureURI(b.picture.DynamicID)
}

func (b onePicV3Builder) activitySupported() bool {
	return b.VersionControl().Can("pic.activitySupported")
}

func (b onePicV3Builder) usingLikeText() bool {
	return b.BuilderContext.VersionControl().Can("pic.usingLikeText")
}

func (b onePicV3Builder) constructArgs() jsoncard.Args {
	return b.ConstructArgsFromPicture(b.picture)
}

func (b onePicV3Builder) constructDescButton() *jsoncard.Button {
	return b.ConstructDescButtonFromPicture(b.picture)
}

func (b onePicV3Builder) constructThreePoint() *jsoncard.ThreePoint {
	return b.ConstructThreePointFromPicture(b.picture)
}

func (b onePicV3Builder) constructThreePointV2() []*jsoncard.ThreePointV2 {
	return b.ConstructThreePointV2FromPicture(b.picture)
}

func (b onePicV3Builder) Build() (*jsoncard.OnePicV3, error) {
	if b.base == nil {
		return nil, errors.Errorf("empty `base` field")
	}
	if b.rcmd == nil {
		return nil, errors.Errorf("empty `rcmd` field")
	}
	if b.picture == nil {
		return nil, errors.Errorf("empty `picture` field")
	}
	if len(b.picture.Imgs) <= 0 {
		return nil, errors.Errorf("empty `img` field in picture: %+v", b.picture)
	}
	if err := jsonbuilder.NewBaseUpdater(b.BuilderContext, b.base).
		UpdateCover(b.picture.Imgs[0]).
		UpdateTitle(b.picture.DynamicText).
		UpdateGoto(appcardmodel.GotoPicture).
		UpdateURI(b.constructURI()).
		UpdateArgs(b.constructArgs()).
		UpdateBaseInnerDescButton(b.constructDescButton()).
		UpdateThreePoint(b.constructThreePoint()).
		UpdateThreePointV2(b.constructThreePointV2()).
		Update(); err != nil {
		return nil, err
	}
	out := &jsoncard.OnePicV3{
		Base:  b.base,
		Badge: "动态",
	}
	if b.picture.JumpUrl != "" && b.activitySupported() {
		out.URI = b.picture.JumpUrl
		out.Badge = "活动"
	}
	out.CoverLeftText1 = appcardmodel.StatString(int32(b.picture.ViewCount), "")
	out.CoverLeftIcon1 = appcardmodel.IconRead
	if b.usingLikeText() {
		out.CoverLeftText1 = appcardmodel.StatString(b.picture.LikeCount, "")
		out.CoverLeftIcon1 = appcardmodel.IconLike
	}
	out.CoverRightText = b.picture.NickName
	if b.rcmd.RcmdReason != nil {
		if err := jsonbuilder.NewBaseUpdater(b.BuilderContext, b.base).
			UpdateBaseInnerDescButton(nil).Update(); err != nil {
			return nil, err
		}
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

func (b onePicV3Builder) WithAfter(req ...func(*jsoncard.OnePicV3)) OnePicV3Builder {
	b.afterFn = append(b.afterFn, req...)
	return b
}

func OnePicV3ByMultiMaterials(arg *tunnelV2.Material) func(*jsoncard.OnePicV3) {
	return func(card *jsoncard.OnePicV3) {
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
