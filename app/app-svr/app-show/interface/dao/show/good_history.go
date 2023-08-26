package show

import (
	"context"
	"encoding/json"

	"go-common/library/cache/redis"
	"go-gateway/app/app-svr/app-show/interface/model/show"

	"github.com/pkg/errors"
)

// RawGoodHisRes def.
func (d *Dao) RawGoodHisRes(ctx context.Context) ([]*show.GoodHisRes, error) {
	conn := d.redis.Get(ctx)
	defer conn.Close()
	reply, err := redis.Bytes(conn.Do("GET", showActionKey("loadGoodHistory", "goodHisRes")))
	if err != nil {
		return nil, err
	}
	var res []*show.GoodHisRes
	if err := json.Unmarshal(reply, &res); err != nil {
		return nil, errors.WithStack(err)
	}
	return res, nil
}
