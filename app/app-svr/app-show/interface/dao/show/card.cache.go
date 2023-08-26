package show

import (
	"context"
	"encoding/json"
	"fmt"

	"go-gateway/app/app-svr/app-show/interface/model/card"

	"go-common/library/cache/redis"
	"go-common/library/log"
)

const _hotTenprefix = "%d_hchashmap"

func getHotKey(i int) string {
	return fmt.Sprintf(_hotTenprefix, i)
}

// nolint:gomnd
func (d *Dao) PopularCardTenCache(c context.Context, i, index, ps int) (cards []*card.PopularCard, err error) {
	var (
		key  = getHotKey(i)
		conn = d.redis.Get(c)
		bss  [][]byte
	)
	ps = ps * 2 // 拿2倍数据
	defer conn.Close()
	arg := redis.Args{}.Add(key)
	for id := 0; id < ps; id++ {
		arg = arg.Add(index + id)
	}
	if bss, err = redis.ByteSlices(conn.Do("HMGET", arg...)); err != nil {
		log.Error("conn.Do(HGET,%s) error(%v)", key, err)
		return
	}
	for _, bs := range bss {
		if len(bs) == 0 { // 如果拿不到，认为后面没有了，直接返回
			return
		}
		card := new(card.PopularCard)
		if err = json.Unmarshal(bs, card); err != nil {
			log.Error("json.Unmarshal(%s) error(%v)", string(bs), err)
			err = nil
			continue
		}
		cards = append(cards, card)
	}
	return
}
