package archive

import (
	"context"
	"encoding/json"

	"go-common/library/cache/redis"
	"go-gateway/app/app-svr/playurl/service/model/archive"
)

const (
	_playerRulesKey = "chronos_player"
)

// PlayerRules .
func (d *Dao) PlayerRules(c context.Context) ([]*archive.PlayerRule, error) {
	conn := d.arcRedis.Conn(c)
	defer conn.Close()
	res, err := redis.Bytes(conn.Do("GET", _playerRulesKey))
	if err != nil {
		if err == redis.ErrNil {
			return make([]*archive.PlayerRule, 0), nil
		}
		return nil, err
	}
	rules := make([]*archive.PlayerRule, 0)
	if err = json.Unmarshal(res, &rules); err != nil {
		return nil, err
	}
	return rules, nil

}
