package show

import (
	"context"
	"encoding/json"

	"go-common/library/database/sql"
	"go-common/library/log"
	"go-gateway/app/app-svr/app-job/job/model/show"

	"github.com/pkg/errors"
)

const (
	goodHistoryRes = "SELECT aid, achievement, add_date FROM good_history WHERE deleted=0 ORDER BY rank ASC"
)

// RawGoodHisRes def.
func (d *Dao) RawGoodHisRes(ctx context.Context) (res []*show.GoodHisRes, err error) {
	var rows *sql.Rows
	if rows, err = d.db.Query(ctx, goodHistoryRes); err != nil {
		log.Error("[RawGoodHisRes] d.db.Query() error(%v)", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		a := &show.GoodHisRes{}
		if err = rows.Scan(&a.Aid, &a.Achievement, &a.AddDate); err != nil {
			log.Error("[RawGoodHisRes] rows.Scan error(%v)", err)
			return
		}
		res = append(res, a)
	}
	if err = rows.Err(); err != nil {
		log.Error("rows.Err() error(%v)", err)
	}
	return
}

func (d *Dao) AddCacheGoodHistory(ctx context.Context, cards []*show.GoodHisRes) error {
	if len(cards) == 0 {
		return nil
	}
	conn := d.redis.Get(ctx)
	defer conn.Close()
	bs, err := json.Marshal(cards)
	if err != nil {
		return errors.WithStack(err)
	}
	key := showActionKey("loadGoodHistory", "goodHisRes")
	if _, err := conn.Do("SETEX", key, _showExpire, bs); err != nil {
		return err
	}
	return nil
}
