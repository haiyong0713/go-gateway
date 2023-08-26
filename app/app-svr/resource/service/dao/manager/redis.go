package manager

import (
	"context"
	"encoding/json"
	"fmt"

	"go-common/library/cache/redis"
	"go-common/library/log"

	pb2 "go-gateway/app/app-svr/resource/service/api/v2"
)

const (
	_prefixSpecialCard = "special:card:v2:%d"
	_expire            = "604800" //redis缓存超时时间
)

func keySpecialCard(id int64) string {
	return fmt.Sprintf(_prefixSpecialCard, id)
}

func (d *Dao) SetSpecial2Cache(ctx context.Context, special *pb2.AppSpecialCard) (err error) {
	key := keySpecialCard(special.Id)
	bs, err := json.Marshal(special)
	if err != nil {
		log.Error("dao.SetSpecial2Cache json.Marshal special(%+v) err(%+v)", special, err)
		return err
	}

	if _, err = d.redis.Do(ctx, "SETEX", key, _expire, bs); err != nil {
		log.Error("dao:SetSpecial2Cache SETEX KEY(%s) err(%+v)", key, err)
	}

	return
}

func (d *Dao) GetSpecialFromCache(ctx context.Context, id int64) (special *pb2.AppSpecialCard, err error) {
	key := keySpecialCard(id)
	bs, err := redis.Bytes(d.redis.Do(ctx, "GET", key))
	if err != nil {
		if err == redis.ErrNil {
			return nil, nil
		}
		log.Error("dao.GetSpecialFromCache GET key(%s) err(%+v)", key, err)
		return nil, err
	}

	if err = json.Unmarshal(bs, &special); err != nil {
		log.Error("dao:GetSpecialFromCache json.Unmarshal err(%+v)", err)
	}

	return
}
