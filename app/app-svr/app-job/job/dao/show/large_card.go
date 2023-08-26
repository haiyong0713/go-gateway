package show

import (
	"context"
	"encoding/json"

	"go-gateway/app/app-svr/app-show/interface/model/show"

	"github.com/pkg/errors"
)

const (
	_validLargeCardsSQL = `SELECT popular_large_card.id,popular_large_card.title,popular_large_card.rid,popular_large_card.white_list,popular_large_card.auto,popular_card.sticky  
		FROM popular_large_card LEFT JOIN popular_card ON popular_large_card.id=popular_card.card_value WHERE popular_large_card.deleted=0 AND popular_card.is_delete=0 AND popular_card.card_type="av_largecard"`
	_loadLargeCardKey = "loadLargeCards"
)

func (d *Dao) LargeCards(ctx context.Context) ([]*show.LargeCard, error) {
	rows, err := d.db.Query(ctx, _validLargeCardsSQL)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []*show.LargeCard
	for rows.Next() {
		var a = new(show.LargeCard)
		if err = rows.Scan(&a.ID, &a.Title, &a.RID, &a.WhiteList, &a.Auto, &a.Sticky); err != nil {
			return nil, errors.Wrapf(err, "SQL %s", _validLargeCardsSQL)
		}
		res = append(res, a)
	}
	if err = rows.Err(); err != nil {
		return nil, errors.Wrapf(err, "SQL %s", _validLargeCardsSQL)
	}
	return res, nil
}

func (d *Dao) AddCacheLargeCards(ctx context.Context, cards []*show.LargeCard) error {
	if len(cards) == 0 {
		return nil
	}
	conn := d.redis.Get(ctx)
	defer conn.Close()
	bs, err := json.Marshal(cards)
	if err != nil {
		return errors.WithStack(err)
	}
	key := showActionKey(_loadLargeCardKey, "largeCard")
	if _, err = conn.Do("SETEX", key, _showExpire, bs); err != nil {
		return err
	}
	return nil
}
