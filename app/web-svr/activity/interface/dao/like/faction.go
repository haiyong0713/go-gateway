package like

import (
	"context"
	"encoding/json"

	"go-common/library/cache/redis"
	"go-gateway/app/web-svr/activity/interface/model/like"

	"github.com/pkg/errors"
)

func (d *Dao) FactionRank(c context.Context) (data []*like.Faction, err error) {
	key := "faction_rank"
	conn := d.redis.Get(c)
	defer conn.Close()
	var bytes []byte
	bytes, err = redis.Bytes(conn.Do("GET", key))
	if err != nil {
		if err == redis.ErrNil {
			err = nil
		} else {
			err = errors.Wrap(err, "FactionRank:redis GET")
		}
		return
	}
	if err = json.Unmarshal(bytes, &data); err != nil {
		err = errors.Wrap(err, "FactionRank json.Unmarshal")
		return
	}
	return
}
