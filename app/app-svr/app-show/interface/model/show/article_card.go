package show

import jobApi "go-gateway/app/app-svr/app-job/job/api"

type ArticleCard struct {
	ID        int64  `json:"-"`
	ArticleID int64  `json:"-"`
	Cover     string `json:"-"`
}

// FromJobPBArticleCard
func (card *ArticleCard) FromJobPBArticleCard(articleCard *jobApi.ArticleCard) {
	card.ID = articleCard.Id
	card.ArticleID = articleCard.ArticleId
	card.Cover = articleCard.Cover
}
