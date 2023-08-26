package show

import (
	"context"
	"encoding/json"

	"go-common/library/cache/redis"
	"go-gateway/app/app-svr/app-show/interface/model/show"

	"github.com/pkg/errors"
)

// LargeCards .
func (d *Dao) LargeCards(ctx context.Context) ([]*show.LargeCard, error) {
	conn := d.redis.Get(ctx)
	defer conn.Close()
	reply, err := redis.Bytes(conn.Do("GET", showActionKey("loadLargeCards", "largeCard")))
	if err != nil {
		return nil, err
	}
	var res []*show.LargeCard
	if err = json.Unmarshal(reply, &res); err != nil {
		return nil, errors.WithStack(err)
	}
	return res, nil
}
