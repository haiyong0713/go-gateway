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

type ThreePicV1Builder interface {
	ReplaceContext(jsonbuilder.BuilderContext) ThreePicV1Builder
	SetBase(*jsoncard.Base) ThreePicV1Builder
	SetPicture(*bplus.Picture) ThreePicV1Builder
	SetRcmd(rcmd *ai.Item) ThreePicV1Builder
	Build() (*jsoncard.ThreePicV1, error)
	WithAfter(req ...func(*jsoncard.ThreePicV1)) ThreePicV1Builder
}

type threePicV1Builder struct {
	jsoncommon.PicCommon
	jsonbuilder.BuilderContext
	base    *jsoncard.Base
	picture *bplus.Picture
	rcmd    *ai.Item
	afterFn []func(*jsoncard.ThreePicV1)
}

func NewThreePicV1Builder(ctx jsonbuilder.BuilderContext) ThreePicV1Builder {
	return threePicV1Builder{BuilderContext: ctx}
}

func (b threePicV1Builder) ReplaceContext(ctx jsonbuilder.BuilderContext) ThreePicV1Builder {
	b.BuilderContext = ctx
	return b
}

func (b threePicV1Builder) SetBase(base *jsoncard.Base) ThreePicV1Builder {
	b.base = base
	return b
}

func (b threePicV1Builder) SetPicture(picture *bplus.Picture) ThreePicV1Builder {
	b.picture = picture
	return b
}

func (b threePicV1Builder) SetRcmd(rcmd *ai.Item) ThreePicV1Builder {
	b.rcmd = rcmd
	return b
}

func (b threePicV1Builder) constructArgs() jsoncard.Args {
	return b.ConstructArgsFromPicture(b.picture)
}

func (b threePicV1Builder) constructURI() string {
	return b.ConstructPictureURI(b.picture.DynamicID)
}

func (b threePicV1Builder) constructDescButton() *jsoncard.Button {
	return b.ConstructDescButtonFromPicture(b.picture)
}

func (b threePicV1Builder) activitySupported() bool {
	return b.VersionControl().Can("pic.activitySupported")
}

func (b threePicV1Builder) usingLikeText() bool {
	return b.BuilderContext.VersionControl().Can("pic.usingLikeText")
}

func (b threePicV1Builder) constructThreePoint() *jsoncard.ThreePoint {
	return b.ConstructThreePointFromPicture(b.picture)
}

func (b threePicV1Builder) constructThreePointV2() []*jsoncard.ThreePointV2 {
	return b.ConstructThreePointV2FromPicture(b.picture)
}

func (b threePicV1Builder) constructAvatar() *jsoncard.Avatar {
	avatar, err := jsonavatar.NewAvatarBuilder(b.BuilderContext).
		SetAvatarStatus(&jsoncard.AvatarStatus{
			Cover: b.picture.FaceImg,
			Goto:  appcardmodel.GotoDynamicMid,
			Param: strconv.FormatInt(b.picture.Mid, 10),
			Type:  appcardmodel.AvatarRound,
		}).Build()
	if err != nil {
		log.Error("Failed to build avatar: %+v", err)
	}
	return avatar
}

func (b threePicV1Builder) Build() (*jsoncard.ThreePicV1, error) {
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
		UpdateBaseInnerDescButton(b.constructDescButton()).
		Update(); err != nil {
		return nil, err
	}
	out := &jsoncard.ThreePicV1{
		Base:       b.base,
		CoverBadge: "动态",
		Covers:     b.picture.Imgs[:3],
	}
	if b.picture.JumpUrl != "" && b.activitySupported() {
		out.URI = b.picture.JumpUrl
		out.CoverBadge = "活动"
	}
	out.TitleLeftText1 = appcardmodel.PictureViewString(b.picture.ViewCount)
	if b.usingLikeText() {
		out.TitleLeftText1 = appcardmodel.LikeString(b.picture.LikeCount)
	}
	out.TitleLeftText2 = appcardmodel.ArticleReplyString(b.picture.CommentCount)
	//nolint:gomnd
	if b.picture.ImgCount > 3 {
		out.CoverRightText = appcardmodel.PictureCountString(b.picture.ImgCount)
		out.CoverRightBackgroundColor = "#66666666"
	}
	out.Desc1 = b.picture.NickName
	out.Desc2 = appcardmodel.PubDataByRequestAt(b.picture.PublishTime.Time(), b.rcmd.RequestAt())
	out.CoverBadgeStyle = jsonreasonstyle.ConstructReasonStyle(appcardmodel.BgColorRed, out.CoverBadge)
	out.Avatar = b.constructAvatar()

	topRcmdReason, bottomRcmdReason := jsonreasonstyle.BuildTopBottomRecommendReasonText(
		b.BuilderContext,
		b.rcmd.RcmdReason,
		b.rcmd.Goto,
		b.BuilderContext.IsAttentionTo(b.picture.Mid),
	)
	out.TopRcmdReason = topRcmdReason
	out.BottomRcmdReason = bottomRcmdReason
	out.TopRcmdReasonStyle = jsonreasonstyle.ConstructTopReasonStyle(
		topRcmdReason,
		jsonreasonstyle.CornerMarkFromAI(b.rcmd),
		jsonreasonstyle.CorverMarkFromContext(b.BuilderContext),
	)
	out.BottomRcmdReasonStyle = jsonreasonstyle.ConstructBottomReasonStyle(
		bottomRcmdReason,
		jsonreasonstyle.CornerMarkFromAI(b.rcmd),
		jsonreasonstyle.CorverMarkFromContext(b.BuilderContext),
	)
	for _, fn := range b.afterFn {
		fn(out)
	}

	return out, nil
}

func (b threePicV1Builder) WithAfter(req ...func(*jsoncard.ThreePicV1)) ThreePicV1Builder {
	b.afterFn = append(b.afterFn, req...)
	return b
}

func ThreePicV1ByMultiMaterials(arg *tunnelV2.Material) func(*jsoncard.ThreePicV1) {
	return func(card *jsoncard.ThreePicV1) {
		if arg == nil {
			return
		}
		if arg.Title != "" {
			card.Title = arg.Title
		}
	}
}
