package show

import (
	"context"
	"encoding/json"
	"fmt"

	"go-common/library/cache/redis"
	"go-common/library/log"
	"go-gateway/app/app-svr/app-car/interface/model/popular"
)

const _hotTenprefix = "%d_hchashmap_car"

func getHotKey(i int) string {
	return fmt.Sprintf(_hotTenprefix, i)
}

func (d *Dao) PopularCardTenCache(c context.Context, i, index, ps int) ([]*popular.PopularCard, error) {
	var (
		key  = getHotKey(i)
		conn = d.redis.Get(c)
		bss  [][]byte
	)
	defer conn.Close()
	arg := redis.Args{}.Add(key)
	for id := 0; id < ps; id++ {
		arg = arg.Add(index + id)
	}
	bss, err := redis.ByteSlices(conn.Do("HMGET", arg...))
	if err != nil {
		log.Error("PopularCardTenCache conn.Do(HGET,%s) error(%v)", key, err)
		return nil, err
	}
	cards := []*popular.PopularCard{}
	for _, bs := range bss {
		if len(bs) == 0 {
			continue
		}
		card := new(popular.PopularCard)
		if err = json.Unmarshal(bs, card); err != nil {
			log.Error("PopularCardTenCache json.Unmarshal(%s) error(%v)", string(bs), err)
			continue
		}
		cards = append(cards, card)
	}
	return cards, nil
}
