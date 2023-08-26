package jsoncommon

import (
	"fmt"

	"go-gateway/app/app-svr/app-card/interface/model"
	jsoncard "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json"

	article "git.bilibili.co/bapis/bapis-go/article/model"
)

type ArticleCommon struct{}

func (ArticleCommon) ConstructThreePointFromArticle(data *article.Meta) *jsoncard.ThreePoint {
	return &jsoncard.ThreePoint{DislikeReasons: constructThreePointFromArticle(data)}
}

func constructThreePointFromArticle(data *article.Meta) []*jsoncard.DislikeReason {
	dislikeReasons := []*jsoncard.DislikeReason{}
	if data.Author != nil && data.Author.Name != "" {
		dislikeReasons = append(dislikeReasons, &jsoncard.DislikeReason{
			ID:    _upper,
			Name:  fmt.Sprintf("UP主:%s", data.Author.Name),
			Toast: _dislikeToast,
		})
	}
	dislikeReasons = append(dislikeReasons, &jsoncard.DislikeReason{
		ID:    _noSeason,
		Name:  "不感兴趣",
		Toast: _dislikeToast,
	})
	return dislikeReasons
}

func (ArticleCommon) ConstructThreePointV2FromArticle(data *article.Meta) []*jsoncard.ThreePointV2 {
	out := []*jsoncard.ThreePointV2{}
	out = append(out, &jsoncard.ThreePointV2{
		Title:    "不感兴趣",
		Subtitle: "(选择后将减少相似内容推荐)",
		Reasons:  constructThreePointFromArticle(data),
		Type:     model.ThreePointDislike,
	})
	return out
}

func (ArticleCommon) ConstructArgsFromArticle(in *article.Meta) jsoncard.Args {
	args := jsoncard.Args{}
	if in.Author != nil {
		args.UpID = in.Author.Mid
		args.UpName = in.Author.Name
	}
	if len(in.Categories) != 0 {
		if in.Categories[0] != nil {
			args.Rid = int32(in.Categories[0].ID)
			args.Rname = in.Categories[0].Name
		}
		if len(in.Categories) > 1 {
			if in.Categories[1] != nil {
				args.Tid = in.Categories[1].ID
				args.Tname = in.Categories[1].Name
			}
		}
	}
	return args
}
