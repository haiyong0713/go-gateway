package article

import (
	"context"
	"strconv"

	"go-gateway/app/app-svr/app-interface/interface-legacy/model"

	"git.bilibili.co/go-tool/libbdevice/pkg/pd"
)

type ArticleInfo struct {
	Badge string
	Uri   string
}

func GetArticleInfo(ctx context.Context, articleType, articleId, aid int64) *ArticleInfo {
	am := &ArticleInfo{
		Badge: "专栏",
		Uri:   model.FillURI(model.GotoArticle, strconv.FormatInt(articleId, 10), nil),
	}
	var noteType = int64(2)
	if articleType == noteType {
		am.Badge = "笔记"
		if useNewNoteUri(ctx) {
			am.Uri = model.FillURI(model.GotoAv, strconv.FormatInt(aid, 10), model.NoteHandler(articleId))
		}
	}
	return am
}

func useNewNoteUri(ctx context.Context) bool {
	if pd.WithContext(ctx).Where(func(pd *pd.PDContext) {
		pd.IsPlatIPhone().And().Build(">=", int64(66900000))
	}).OrWhere(func(pd *pd.PDContext) {
		pd.IsPlatAndroid().And().Build(">=", int64(6690000))
	}).MustFinish() {
		return true
	}
	return false
}
