package archive

import (
	"context"
	"encoding/json"

	"go-common/library/cache/redis"
	"go-gateway/app/app-svr/app-view/interface/model/view"
)

const (
	_playerRulesKey = "chronos_player"
)

// PlayerRules .
func (d *Dao) PlayerRules(c context.Context) ([]*view.ChronosRule, error) {
	conn := d.redis.Get(c)
	defer conn.Close()
	res, err := redis.Bytes(conn.Do("GET", _playerRulesKey))
	if err != nil {
		if err == redis.ErrNil {
			return make([]*view.ChronosRule, 0), nil
		}
		return nil, err
	}
	rules := make([]*view.ChronosRule, 0)
	if err = json.Unmarshal(res, &rules); err != nil {
		return nil, err
	}
	return rules, nil
}

func (d *Dao) ChronosPkgInfo(c context.Context) (map[string][]*view.PackageInfo, error) {
	conn := d.redis.Get(c)
	defer conn.Close()
	res, err := redis.Bytes(conn.Do("GET", "chronosV2"))
	if err != nil {
		return nil, err
	}
	rules := make(map[string][]*view.PackageInfo)
	if err = json.Unmarshal(res, &rules); err != nil {
		return nil, err
	}
	return rules, nil
}
