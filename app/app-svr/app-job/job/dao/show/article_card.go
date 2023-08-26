package show

import (
	"context"

	v1 "go-gateway/app/app-svr/app-job/job/api"

	"github.com/pkg/errors"
)

const (
	_articleCardSQL = "SELECT id,article_id,cover FROM popular_article_card WHERE state=1"
)

func (d *Dao) ArticleCard(ctx context.Context) (*v1.ArticleCardMapReply, error) {
	rows, err := d.db.Query(ctx, _articleCardSQL)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	cards := map[int64]*v1.ArticleCard{}
	for rows.Next() {
		a := &v1.ArticleCard{}
		if err = rows.Scan(&a.Id, &a.ArticleId, &a.Cover); err != nil {
			return nil, err
		}
		cards[a.Id] = a
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	res := &v1.ArticleCardMapReply{}
	for k, v := range cards {
		res.Cardm = append(res.Cardm, &v1.ArticleCardMap{
			Key:  k,
			Card: v,
		})
	}
	return res, nil
}

func (d *Dao) AddCacheArticleCards(ctx context.Context, cards *v1.ArticleCardMapReply) error {
	if cards.Size() <= 0 {
		return nil
	}
	conn := d.redis.Get(ctx)
	defer conn.Close()
	val, err := cards.Marshal()
	if err != nil {
		return errors.WithStack(err)
	}
	key := showActionKey("loadArticleCardsCache", "ArticleCardMapReply")
	if _, err = conn.Do("SETEX", key, _showExpire, val); err != nil {
		return err
	}
	return nil
}
