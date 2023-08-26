package jsonthreeitemhv3

import (
	"fmt"
	"strconv"

	accountgrpc "git.bilibili.co/bapis/bapis-go/account/service"
	"go-common/library/log"
	appcardmodel "go-gateway/app/app-svr/app-card/interface/model"
	appcard "go-gateway/app/app-svr/app-card/interface/model/card"
	"go-gateway/app/app-svr/app-card/interface/model/card/ai"
	jsoncard "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json"
	jsonbuilder "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder"
	jsonavatar "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder/avatar"
	jsoncommon "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder/common"
	jsonreasonstyle "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder/reason_style"

	article "git.bilibili.co/bapis/bapis-go/article/model"

	"github.com/pkg/errors"
)

type ThreeItemHV3Builder interface {
	ReplaceContext(jsonbuilder.BuilderContext) ThreeItemHV3Builder
	SetBase(*jsoncard.Base) ThreeItemHV3Builder
	SetRcmd(rcmd *ai.Item) ThreeItemHV3Builder
	SetArticle(*article.Meta) ThreeItemHV3Builder
	SetAuthorCard(*accountgrpc.Card) ThreeItemHV3Builder
	Build() (*jsoncard.ThreeItemHV3, error)
}

type threeItemHV3Builder struct {
	jsonbuilder.BuilderContext
	jsoncommon.ArticleCommon
	base       *jsoncard.Base
	rcmd       *ai.Item
	article    *article.Meta
	authorCard *accountgrpc.Card
}

func NewThreeItemHV3BuilderBuilder(ctx jsonbuilder.BuilderContext) ThreeItemHV3Builder {
	return threeItemHV3Builder{BuilderContext: ctx}
}

func (b threeItemHV3Builder) ReplaceContext(ctx jsonbuilder.BuilderContext) ThreeItemHV3Builder {
	b.BuilderContext = ctx
	return b
}

func (b threeItemHV3Builder) SetBase(in *jsoncard.Base) ThreeItemHV3Builder {
	b.base = in
	return b
}

func (b threeItemHV3Builder) SetArticle(in *article.Meta) ThreeItemHV3Builder {
	b.article = in
	return b
}

func (b threeItemHV3Builder) SetRcmd(in *ai.Item) ThreeItemHV3Builder {
	b.rcmd = in
	return b
}

func (b threeItemHV3Builder) SetAuthorCard(in *accountgrpc.Card) ThreeItemHV3Builder {
	b.authorCard = in
	return b
}

func (b threeItemHV3Builder) constructArgs() jsoncard.Args {
	return b.ConstructArgsFromArticle(b.article)
}

func (b threeItemHV3Builder) constructThreePoint() *jsoncard.ThreePoint {
	return b.ConstructThreePointFromArticle(b.article)
}

func (b threeItemHV3Builder) constructThreePointV2() []*jsoncard.ThreePointV2 {
	return b.ConstructThreePointV2FromArticle(b.article)
}

func (b threeItemHV3Builder) constructURI() string {
	device := b.BuilderContext.Device()
	return appcardmodel.FillURI(appcardmodel.GotoArticle, device.Plat(), int(device.Build()), strconv.FormatInt(b.article.ID, 10), nil)
}

func (b threeItemHV3Builder) constructAvatar() *jsoncard.Avatar {
	avatar, err := jsonavatar.NewAvatarBuilder(b.BuilderContext).
		SetAvatarStatus(&jsoncard.AvatarStatus{
			Cover: b.article.Author.Face,
			Text:  fmt.Sprintf("%s·%s", b.article.Author.Name, appcardmodel.PubDataByRequestAt(b.article.PublishTime.Time(), b.rcmd.RequestAt())),
			Goto:  appcardmodel.GotoMid,
			Param: strconv.FormatInt(b.article.Author.Mid, 10),
			Type:  appcardmodel.AvatarRound,
		}).Build()
	if err != nil {
		log.Warn("Failed to build avatar: %+v", err)
	}
	return avatar
}

func (b threeItemHV3Builder) Build() (*jsoncard.ThreeItemHV3, error) {
	if b.base == nil {
		return nil, errors.Errorf("empty `base` field")
	}
	if b.rcmd == nil {
		return nil, errors.Errorf("empty `rcmd` field")
	}
	if b.article == nil {
		return nil, errors.Errorf("empty `article` field")
	}
	if appcard.CheckMidMaxInt32(b.article.Author.Mid) && b.BuilderContext.VersionControl().Can("feed.disableInt64Mid") {
		return nil, errors.Errorf("ignore on maxint32 mid: %d", b.article.Author.Mid)
	}

	if err := jsonbuilder.NewBaseUpdater(b.BuilderContext, b.base).
		UpdateCover("").
		UpdateTitle(b.article.Title).
		UpdateArgs(b.constructArgs()).
		UpdateURI(b.constructURI()).
		UpdateThreePoint(b.constructThreePoint()).
		UpdateThreePointV2(b.constructThreePointV2()).
		Update(); err != nil {
		return nil, err
	}

	out := &jsoncard.ThreeItemHV3{
		Covers:        b.article.ImageURLs,
		CoverTopText1: appcardmodel.ArticleViewString(b.article.Stats.View),
		CoverTopText2: appcardmodel.ArticleReplyString(b.article.Stats.Reply),
		Desc:          b.article.Summary,
		Avatar:        b.constructAvatar(),
	}
	if b.rcmd.RcmdReason != nil {
		out.RcmdReasonStyle = jsonreasonstyle.ConstructTopReasonStyle(
			b.rcmd.RcmdReason.Content,
			jsonreasonstyle.CornerMarkFromAI(b.rcmd),
			jsonreasonstyle.CorverMarkFromContext(b.BuilderContext),
		)
	}
	out.OfficialIcon = appcardmodel.OfficialIcon(b.authorCard)
	out.Base = b.base

	return out, nil
}
