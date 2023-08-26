package show

import (
	"context"
	"encoding/json"

	"go-common/library/cache/redis"
	"go-gateway/app/app-svr/app-show/interface/model/show"

	"github.com/pkg/errors"
)

// Entrances .
func (d *Dao) Entrances(ctx context.Context) ([]*show.EntranceMem, error) {
	conn := d.redis.Get(ctx)
	defer conn.Close()
	reply, err := redis.Bytes(conn.Do("GET", showActionKey("loadPopEntrances", "entranceMem")))
	if err != nil {
		return nil, err
	}
	var res []*show.EntranceMem
	if err = json.Unmarshal(reply, &res); err != nil {
		return nil, errors.WithStack(err)
	}
	return res, nil

}

func (d *Dao) MidTopPhoto(ctx context.Context) (string, error) {
	conn := d.redis.Get(ctx)
	defer conn.Close()
	reply, err := redis.String(conn.Do("GET", showActionKey("loadMiddTopPhoto", "string")))
	if err != nil {
		return reply, err
	}
	return reply, nil
}
