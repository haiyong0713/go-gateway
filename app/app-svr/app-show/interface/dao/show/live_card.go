package show

import (
	"context"
	"encoding/json"

	"go-common/library/cache/redis"
	"go-gateway/app/app-svr/app-show/interface/model/show"

	"github.com/pkg/errors"
)

func (d *Dao) LiveCards(ctx context.Context) ([]*show.LiveCard, error) {
	conn := d.redis.Get(ctx)
	defer conn.Close()
	reply, err := redis.Bytes(conn.Do("GET", showActionKey("loadLiveCards", "liveCard")))
	if err != nil {
		return nil, err
	}
	var res []*show.LiveCard
	if err := json.Unmarshal(reply, &res); err != nil {
		return nil, errors.WithStack(err)
	}
	return res, nil
}
