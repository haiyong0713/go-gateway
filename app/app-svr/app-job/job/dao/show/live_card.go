package show

import (
	"context"
	"encoding/json"

	"go-gateway/app/app-svr/app-show/interface/model/show"

	"github.com/pkg/errors"
)

const (
	_validLiveCardsSQL = `SELECT id,rid,cover from popular_live_card WHERE state=1`
	_loadLiveCardKey   = "loadLiveCards"
)

func (d *Dao) LiveCards(ctx context.Context) ([]*show.LiveCard, error) {
	rows, err := d.db.Query(ctx, _validLiveCardsSQL)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []*show.LiveCard
	for rows.Next() {
		var a = new(show.LiveCard)
		if err = rows.Scan(&a.ID, &a.RID, &a.Cover); err != nil {
			return nil, errors.Wrapf(err, "SQL %s", _validLiveCardsSQL)
		}
		res = append(res, a)
	}
	if err = rows.Err(); err != nil {
		return nil, errors.Wrapf(err, "SQL %s", _validLiveCardsSQL)
	}
	return res, nil
}

func (d *Dao) AddCacheLiveCards(ctx context.Context, res []*show.LiveCard) error {
	if len(res) == 0 {
		return nil
	}
	conn := d.redis.Get(ctx)
	defer conn.Close()
	bs, err := json.Marshal(res)
	if err != nil {
		return errors.WithStack(err)
	}
	key := showActionKey(_loadLiveCardKey, "liveCard")
	if _, err = conn.Do("SETEX", key, _showExpire, bs); err != nil {
		return err
	}
	return nil
}
