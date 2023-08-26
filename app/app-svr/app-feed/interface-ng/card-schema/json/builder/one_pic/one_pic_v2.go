package jsononepic

import (
	"strconv"

	"go-common/library/log"
	appcardmodel "go-gateway/app/app-svr/app-card/interface/model"
	"go-gateway/app/app-svr/app-card/interface/model/bplus"
	"go-gateway/app/app-svr/app-card/interface/model/card/ai"
	jsoncard "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json"
	jsonbuilder "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder"
	jsonavatar "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder/avatar"
	jsoncommon "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder/common"
	jsonreasonstyle "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder/reason_style"

	tunnelV2 "git.bilibili.co/bapis/bapis-go/ai/feed/mgr/service"
	"github.com/pkg/errors"
)

type OnePicV2Builder interface {
	ReplaceContext(jsonbuilder.BuilderContext) OnePicV2Builder
	SetBase(*jsoncard.Base) OnePicV2Builder
	SetPicture(*bplus.Picture) OnePicV2Builder
	SetRcmd(*ai.Item) OnePicV2Builder

	Build() (*jsoncard.OnePicV2, error)
	WithAfter(req ...func(*jsoncard.OnePicV2)) OnePicV2Builder
}

type onePicV2Builder struct {
	jsoncommon.PicCommon
	jsonbuilder.BuilderContext
	rcmd    *ai.Item
	base    *jsoncard.Base
	picture *bplus.Picture
	afterFn []func(*jsoncard.OnePicV2)
}

func NewOnePicV2Builder(ctx jsonbuilder.BuilderContext) OnePicV2Builder {
	return onePicV2Builder{BuilderContext: ctx}
}

func (b onePicV2Builder) ReplaceContext(ctx jsonbuilder.BuilderContext) OnePicV2Builder {
	b.BuilderContext = ctx
	return b
}

func (b onePicV2Builder) SetBase(base *jsoncard.Base) OnePicV2Builder {
	b.base = base
	return b
}

func (b onePicV2Builder) SetPicture(picture *bplus.Picture) OnePicV2Builder {
	b.picture = picture
	return b
}

func (b onePicV2Builder) SetRcmd(rcmd *ai.Item) OnePicV2Builder {
	b.rcmd = rcmd
	return b
}

func (b onePicV2Builder) constructURI() string {
	return b.ConstructPictureURI(b.picture.DynamicID)
}

func (b onePicV2Builder) constructArgs() jsoncard.Args {
	return b.ConstructArgsFromPicture(b.picture)
}

func (b onePicV2Builder) constructDescButton() *jsoncard.Button {
	return b.ConstructDescButtonFromPicture(b.picture)
}

func (b onePicV2Builder) constructThreePoint() *jsoncard.ThreePoint {
	return b.ConstructThreePointFromPicture(b.picture)
}

func (b onePicV2Builder) constructThreePointV2() []*jsoncard.ThreePointV2 {
	return b.ConstructThreePointV2FromPicture(b.picture)
}

func (b onePicV2Builder) activitySupported() bool {
	return b.BuilderContext.VersionControl().Can("pic.activitySupported")
}

func (b onePicV2Builder) usingLikeText() bool {
	return b.BuilderContext.VersionControl().Can("pic.usingLikeText")
}

func (b onePicV2Builder) Build() (*jsoncard.OnePicV2, error) {
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
		return nil, errors.Errorf("empty `imgs` field in picture: %+v", b.picture)
	}
	if b.picture.ViewCount == 0 {
		return nil, errors.Errorf("invalid `view_count` in picture: %+v", b.picture)
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
	out := &jsoncard.OnePicV2{
		Base:  b.base,
		Badge: "动态",
	}
	if b.picture.JumpUrl != "" && b.activitySupported() {
		out.URI = b.picture.JumpUrl
	}
	out.CoverLeftText1 = appcardmodel.StatString(int32(b.picture.ViewCount), "")
	out.CoverLeftIcon1 = appcardmodel.IconRead
	if b.usingLikeText() {
		out.CoverLeftText1 = appcardmodel.StatString(b.picture.LikeCount, "")
		out.CoverLeftIcon1 = appcardmodel.IconLike
	}
	if b.picture.ImgCount > 1 {
		out.CoverRightText = appcardmodel.PictureCountString(b.picture.ImgCount)
		out.CoverRightBackgroundColor = "#66666666"
	}
	if b.rcmd.RcmdReason != nil {
		out.Badge = ""
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
		out.RcmdReason = reasonText
	}
	avatar, err := jsonavatar.NewAvatarBuilder(b.BuilderContext).
		SetAvatarStatus(&jsoncard.AvatarStatus{
			Cover: b.picture.FaceImg,
			Text:  b.picture.NickName,
			Goto:  appcardmodel.GotoDynamicMid,
			Param: strconv.FormatInt(b.picture.Mid, 10),
			Type:  appcardmodel.AvatarRound,
		}).Build()
	if err != nil {
		log.Error("Failed to build avatar: %+v", err)
	}
	out.Avatar = avatar
	for _, fn := range b.afterFn {
		fn(out)
	}
	return out, nil
}

func (b onePicV2Builder) WithAfter(req ...func(*jsoncard.OnePicV2)) OnePicV2Builder {
	b.afterFn = append(b.afterFn, req...)
	return b
}

func OnePicV2ByMultiMaterials(arg *tunnelV2.Material) func(*jsoncard.OnePicV2) {
	return func(card *jsoncard.OnePicV2) {
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
