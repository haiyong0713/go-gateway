package large_cover_v1

import (
	"fmt"
	"strconv"

	"go-common/library/log"
	appcardmodel "go-gateway/app/app-svr/app-card/interface/model"
	"go-gateway/app/app-svr/app-card/interface/model/card"
	"go-gateway/app/app-svr/app-card/interface/model/card/ai"
	jsoncard "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json"
	jsonbuilder "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder"
	jsonavatar "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder/avatar"
	jsoncommon "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder/common"
	jsonreasonstyle "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder/reason_style"

	accountgrpc "git.bilibili.co/bapis/bapis-go/account/service"
	article "git.bilibili.co/bapis/bapis-go/article/model"
	"github.com/pkg/errors"
)

type V1ArticleBuilder interface {
	Parent() LargeCoverV1BuilderFactory
	SetBase(*jsoncard.Base) V1ArticleBuilder
	SetRcmd(*ai.Item) V1ArticleBuilder
	SetArticle(*article.Meta) V1ArticleBuilder
	SetAuthorCard(*accountgrpc.Card) V1ArticleBuilder
	Build() (*jsoncard.LargeCoverV1, error)
}

type v1ArticleBuilder struct {
	parent *largeCoverV1BuilderFactory
	base   *jsoncard.Base
	rcmd   *ai.Item
	jsoncommon.ArticleCommon
	article    *article.Meta
	authorCard *accountgrpc.Card
}

func (b v1ArticleBuilder) Parent() LargeCoverV1BuilderFactory {
	return b.parent
}

func (b v1ArticleBuilder) SetBase(base *jsoncard.Base) V1ArticleBuilder {
	b.base = base
	return b
}

func (b v1ArticleBuilder) SetRcmd(item *ai.Item) V1ArticleBuilder {
	b.rcmd = item
	return b
}

func (b v1ArticleBuilder) SetArticle(in *article.Meta) V1ArticleBuilder {
	b.article = in
	return b
}

func (b v1ArticleBuilder) SetAuthorCard(in *accountgrpc.Card) V1ArticleBuilder {
	b.authorCard = in
	return b
}

func (b v1ArticleBuilder) constructArgs() jsoncard.Args {
	return b.ConstructArgsFromArticle(b.article)
}

func (b v1ArticleBuilder) constructThreePoint() *jsoncard.ThreePoint {
	return b.ConstructThreePointFromArticle(b.article)
}

func (b v1ArticleBuilder) constructThreePointV2() []*jsoncard.ThreePointV2 {
	return b.ConstructThreePointV2FromArticle(b.article)
}

func (b v1ArticleBuilder) constructURI() string {
	device := b.parent.BuilderContext.Device()
	return appcardmodel.FillURI(appcardmodel.GotoArticle, device.Plat(), int(device.Build()), strconv.FormatInt(b.article.ID, 10), nil)
}

func (b v1ArticleBuilder) constructAvatar() *jsoncard.Avatar {
	avatar, err := jsonavatar.NewAvatarBuilder(b.parent.BuilderContext).
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

func (b v1ArticleBuilder) Build() (*jsoncard.LargeCoverV1, error) {
	if b.base == nil {
		return nil, errors.Errorf("empty `base` field")
	}
	if b.rcmd == nil {
		return nil, errors.Errorf("empty `rcmd` field")
	}
	if b.article == nil {
		return nil, errors.Errorf("empty `article` field")
	}
	if b.parent.BuilderContext.VersionControl().Can("feed.disableHDArticleS") {
		return nil, errors.Errorf("disable article_s build")
	}
	if card.CheckMidMaxInt32(b.article.Author.Mid) && b.parent.BuilderContext.VersionControl().Can("feed.disableInt64Mid") {
		return nil, errors.Errorf("ignore on maxint32 mid: %d", b.article.Author.Mid)
	}
	if len(b.article.ImageURLs) == 0 {
		return nil, errors.Errorf("empty `ImageURLs`")
	}
	if err := jsonbuilder.NewBaseUpdater(b.parent.BuilderContext, b.base).
		UpdateURI(b.constructURI()).
		UpdateCover(b.article.ImageURLs[0]).
		UpdateTitle(b.article.Title).
		UpdateArgs(b.constructArgs()).
		UpdateThreePoint(b.constructThreePoint()).
		UpdateThreePointV2(b.constructThreePointV2()).
		Update(); err != nil {
		return nil, err
	}
	out := &jsoncard.LargeCoverV1{
		Avatar:          b.constructAvatar(),
		CoverBadge:      "专栏",
		CoverBadgeStyle: jsonreasonstyle.ConstructReasonStyle(appcardmodel.BgColorRed, "专栏"),
		Desc: fmt.Sprintf("%s · %s", b.article.Author.Name,
			appcardmodel.PubDataByRequestAt(b.article.PublishTime.Time(), b.rcmd.RequestAt())),
	}
	if b.article.Stats != nil {
		out.CoverLeftText1 = appcardmodel.StatString(int32(b.article.Stats.View), "")
		out.CoverLeftIcon1 = appcardmodel.IconRead
		out.CoverLeftText2 = appcardmodel.StatString(int32(b.article.Stats.Reply), "")
		out.CoverLeftIcon2 = appcardmodel.IconComment
	}
	topRcmdReason, bottomRcmdReason := jsonreasonstyle.BuildTopBottomRecommendReasonText(
		b.parent.BuilderContext,
		b.rcmd.RcmdReason,
		b.rcmd.Goto,
		false,
	)
	out.TopRcmdReason = topRcmdReason
	out.BottomRcmdReason = bottomRcmdReason
	out.TopRcmdReasonStyle = jsonreasonstyle.ConstructTopReasonStyle(
		topRcmdReason,
		jsonreasonstyle.CornerMarkFromAI(b.rcmd),
		jsonreasonstyle.CorverMarkFromContext(b.parent.BuilderContext),
	)
	out.BottomRcmdReasonStyle = jsonreasonstyle.ConstructBottomReasonStyle(
		bottomRcmdReason,
		jsonreasonstyle.CornerMarkFromAI(b.rcmd),
		jsonreasonstyle.CorverMarkFromContext(b.parent.BuilderContext),
	)
	out.Base = b.base
	return out, nil
}
