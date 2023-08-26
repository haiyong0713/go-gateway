package jsonthreepic

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

//go:generate python3 ../../../../contrib/desc-button-overlapped.py

type ThreePicV2Builder interface {
	ReplaceContext(jsonbuilder.BuilderContext) ThreePicV2Builder
	SetBase(*jsoncard.Base) ThreePicV2Builder
	SetPicture(*bplus.Picture) ThreePicV2Builder
	SetRcmd(*ai.Item) ThreePicV2Builder

	Build() (*jsoncard.ThreePicV2, error)
	WithAfter(req ...func(*jsoncard.ThreePicV2)) ThreePicV2Builder
}

type threePicV2Builder struct {
	jsoncommon.PicCommon
	jsonbuilder.BuilderContext
	base    *jsoncard.Base
	picture *bplus.Picture
	rcmd    *ai.Item
	afterFn []func(*jsoncard.ThreePicV2)
}

func NewThreePicV2Builder(ctx jsonbuilder.BuilderContext) ThreePicV2Builder {
	return threePicV2Builder{BuilderContext: ctx}
}

func (b threePicV2Builder) ReplaceContext(ctx jsonbuilder.BuilderContext) ThreePicV2Builder {
	b.BuilderContext = ctx
	return b
}

func (b threePicV2Builder) SetBase(base *jsoncard.Base) ThreePicV2Builder {
	b.base = base
	return b
}

func (b threePicV2Builder) SetPicture(picture *bplus.Picture) ThreePicV2Builder {
	b.picture = picture
	return b
}

func (b threePicV2Builder) SetRcmd(rcmd *ai.Item) ThreePicV2Builder {
	b.rcmd = rcmd
	return b
}

func (b threePicV2Builder) constructArgs() jsoncard.Args {
	return b.ConstructArgsFromPicture(b.picture)
}

func (b threePicV2Builder) constructURI() string {
	return b.ConstructPictureURI(b.picture.DynamicID)
}

func (b threePicV2Builder) constructDescButton() *jsoncard.Button {
	return b.ConstructDescButtonFromPicture(b.picture)
}

func (b threePicV2Builder) activitySupported() bool {
	return b.VersionControl().Can("pic.activitySupported")
}

func (b threePicV2Builder) usingLikeText() bool {
	return b.BuilderContext.VersionControl().Can("pic.usingLikeText")
}

func (b threePicV2Builder) constructThreePoint() *jsoncard.ThreePoint {
	return b.ConstructThreePointFromPicture(b.picture)
}

func (b threePicV2Builder) constructThreePointV2() []*jsoncard.ThreePointV2 {
	return b.ConstructThreePointV2FromPicture(b.picture)
}

func (b threePicV2Builder) Build() (*jsoncard.ThreePicV2, error) {
	if b.base == nil {
		return nil, errors.Errorf("empty `base` field")
	}
	if b.picture == nil {
		return nil, errors.Errorf("empty `picture` field")
	}
	if b.rcmd == nil {
		return nil, errors.Errorf("empty `rcmd` field")
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
	out := &jsoncard.ThreePicV2{
		Base:       b.base,
		Badge:      "动态",
		DescButton: b.constructDescButton(),
	}
	if b.picture.JumpUrl != "" && b.activitySupported() {
		out.URI = b.picture.JumpUrl
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
	//nolint:gomnd
	if b.picture.ImgCount > 3 {
		out.CoverRightText = appcardmodel.PictureCountString(b.picture.ImgCount)
		out.CoverRightBackgroundColor = "#66666666"
	}
	if b.rcmd.RcmdReason != nil {
		out.Badge = ""
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

func (b threePicV2Builder) WithAfter(req ...func(*jsoncard.ThreePicV2)) ThreePicV2Builder {
	b.afterFn = append(b.afterFn, req...)
	return b
}

func ThreePicV2ByMultiMaterials(arg *tunnelV2.Material) func(*jsoncard.ThreePicV2) {
	return func(card *jsoncard.ThreePicV2) {
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
