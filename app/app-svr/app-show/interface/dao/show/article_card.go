package show

import (
	"context"

	"go-common/library/cache/redis"

	jobApi "go-gateway/app/app-svr/app-job/job/api"
	"go-gateway/app/app-svr/app-show/interface/model/show"
)

func (d *Dao) ArticleCard(ctx context.Context) (map[int64]*show.ArticleCard, error) {
	conn := d.redis.Get(ctx)
	defer conn.Close()
	reply, err := redis.Bytes(conn.Do("GET", showActionKey("loadArticleCardsCache", "ArticleCardMapReply")))
	if err != nil {
		return nil, err
	}
	var raw jobApi.ArticleCardMapReply
	if err = raw.Unmarshal(reply); err != nil {
		return nil, err
	}
	res := map[int64]*show.ArticleCard{}
	for _, v := range raw.Cardm {
		itm := show.ArticleCard{}
		itm.FromJobPBArticleCard(v.Card)
		res[v.Key] = &itm
	}
	return res, nil
}
