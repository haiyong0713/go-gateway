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

	accountgrpc "git.bilibili.co/bapis/bapis-go/account/service"
	tunnelV2 "git.bilibili.co/bapis/bapis-go/ai/feed/mgr/service"
	"github.com/pkg/errors"
)

type OnePicV1Builder interface {
	ReplaceContext(jsonbuilder.BuilderContext) OnePicV1Builder
	SetBase(*jsoncard.Base) OnePicV1Builder
	SetPicture(*bplus.Picture) OnePicV1Builder
	SetRcmd(*ai.Item) OnePicV1Builder
	SetAuthor(*accountgrpc.Card) OnePicV1Builder
	Build() (*jsoncard.OnePicV1, error)
	WithAfter(req ...func(*jsoncard.OnePicV1)) OnePicV1Builder
}

type onePicV1Builder struct {
	jsoncommon.PicCommon
	jsonbuilder.BuilderContext
	rcmd    *ai.Item
	base    *jsoncard.Base
	picture *bplus.Picture
	author  *accountgrpc.Card
	afterFn []func(*jsoncard.OnePicV1)
}

func NewOnePicV1Builder(ctx jsonbuilder.BuilderContext) OnePicV1Builder {
	return onePicV1Builder{BuilderContext: ctx}
}

func (b onePicV1Builder) ReplaceContext(ctx jsonbuilder.BuilderContext) OnePicV1Builder {
	b.BuilderContext = ctx
	return b
}

func (b onePicV1Builder) SetBase(base *jsoncard.Base) OnePicV1Builder {
	b.base = base
	return b
}

func (b onePicV1Builder) SetPicture(picture *bplus.Picture) OnePicV1Builder {
	b.picture = picture
	return b
}

func (b onePicV1Builder) SetRcmd(rcmd *ai.Item) OnePicV1Builder {
	b.rcmd = rcmd
	return b
}

func (b onePicV1Builder) SetAuthor(in *accountgrpc.Card) OnePicV1Builder {
	b.author = in
	return b
}

func (b onePicV1Builder) constructURI() string {
	return appcardmodel.FillURI(appcardmodel.GotoPicture, b.BuilderContext.Device().Plat(), int(b.BuilderContext.Device().Build()), strconv.FormatInt(b.picture.DynamicID, 10), nil)
}

func (b onePicV1Builder) constructArgs() jsoncard.Args {
	return b.ConstructArgsFromPicture(b.picture)
}

func (b onePicV1Builder) constructDescButton() *jsoncard.Button {
	return b.ConstructDescButtonFromPicture(b.picture)
}

func (b onePicV1Builder) constructThreePoint() *jsoncard.ThreePoint {
	return b.ConstructThreePointFromPicture(b.picture)
}

func (b onePicV1Builder) constructThreePointV2() []*jsoncard.ThreePointV2 {
	return b.ConstructThreePointV2FromPicture(b.picture)
}

func (b onePicV1Builder) activitySupported() bool {
	return b.BuilderContext.VersionControl().Can("pic.activitySupported")
}

func (b onePicV1Builder) usingLikeText() bool {
	return b.BuilderContext.VersionControl().Can("pic.usingLikeText")
}

func (b onePicV1Builder) Build() (*jsoncard.OnePicV1, error) {
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
		UpdateTitle(b.picture.DynamicText).
		UpdateGoto(appcardmodel.GotoPicture).
		UpdateCover(b.picture.Imgs[0]).
		UpdateURI(b.constructURI()).
		UpdateArgs(b.constructArgs()).
		UpdateBaseInnerDescButton(b.constructDescButton()).
		UpdateThreePoint(b.constructThreePoint()).
		UpdateThreePointV2(b.constructThreePointV2()).
		Update(); err != nil {
		return nil, err
	}
	out := &jsoncard.OnePicV1{
		Base:       b.base,
		CoverBadge: "动态",
	}
	if b.picture.JumpUrl != "" && b.activitySupported() {
		out.URI = b.picture.JumpUrl
		out.CoverBadge = "活动"
	}
	out.CoverLeftText1 = appcardmodel.PictureViewString(b.picture.ViewCount)
	if b.usingLikeText() {
		out.CoverLeftText1 = appcardmodel.LikeString(b.picture.LikeCount)
	}
	out.CoverLeftText2 = appcardmodel.ArticleReplyString(b.picture.CommentCount)
	if b.picture.ImgCount > 1 {
		out.CoverRightText = appcardmodel.PictureCountString(b.picture.ImgCount)
		out.CoverRightBackgroundColor = "#66666666"
	}
	out.Desc1 = b.picture.NickName
	out.Desc2 = appcardmodel.PubDataByRequestAt(b.picture.PublishTime.Time(), b.rcmd.RequestAt())
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
	out.Avatar = avatar
	out.CoverBadgeStyle = jsonreasonstyle.ConstructReasonStyle(appcardmodel.BgColorRed, out.CoverBadge)

	out.TopRcmdReason, out.BottomRcmdReason = jsonreasonstyle.BuildTopBottomRecommendReasonText(
		b.BuilderContext,
		b.rcmd.RcmdReason,
		b.rcmd.Goto,
		b.BuilderContext.IsAttentionTo(b.picture.Mid),
	)
	out.TopRcmdReasonStyle = jsonreasonstyle.ConstructTopReasonStyle(
		out.TopRcmdReason,
		jsonreasonstyle.CornerMarkFromAI(b.rcmd),
		jsonreasonstyle.CorverMarkFromContext(b.BuilderContext),
	)
	out.BottomRcmdReasonStyle = jsonreasonstyle.ConstructBottomReasonStyle(
		out.BottomRcmdReason,
		jsonreasonstyle.CornerMarkFromAI(b.rcmd),
		jsonreasonstyle.CorverMarkFromContext(b.BuilderContext),
	)
	out.OfficialIcon = appcardmodel.OfficialIcon(b.author)

	for _, fn := range b.afterFn {
		fn(out)
	}
	return out, nil
}

func (b onePicV1Builder) WithAfter(req ...func(*jsoncard.OnePicV1)) OnePicV1Builder {
	b.afterFn = append(b.afterFn, req...)
	return b
}

func OnePicV1ByMultiMaterials(arg *tunnelV2.Material) func(*jsoncard.OnePicV1) {
	return func(card *jsoncard.OnePicV1) {
		if arg == nil {
			return
		}
		if arg.Title != "" {
			card.Title = arg.Title
		}
	}
}
