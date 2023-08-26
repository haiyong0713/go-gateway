package feed

import (
	"context"
	"fmt"

	"go-common/library/cache/memcache"
	"go-gateway/app/app-svr/app-card/interface/model/card/ai"
	"go-gateway/app/app-svr/app-card/interface/model/card/operate"

	"github.com/pkg/errors"
)

const (
	_prefixRcmd       = "rc3"
	_prefixConvergeAi = "rcai_%d"
)

func keyRcmd() string {
	return _prefixRcmd
}

func keyConvergeAi(i int64) string {
	return fmt.Sprintf(_prefixConvergeAi, i)
}

// AddRcmdCache add ai Item data into cahce.
func (d *Dao) AddRcmdCache(c context.Context, is []*ai.Item) (err error) {
	conn := d.mcRcmd.Get(c)
	key := keyRcmd()
	item := &memcache.Item{Key: key, Object: is, Flags: memcache.FlagJSON, Expiration: d.expireMC}
	if err = conn.Set(item); err != nil {
		err = errors.Wrap(err, key)
	}
	conn.Close()
	return
}

// AddConvergeAiCache add converge ai
func (d *Dao) AddConvergeAiCache(c context.Context, id int64, card *operate.Card) (err error) {
	var (
		key  = keyConvergeAi(id)
		conn = d.mcRcmd.Get(c)
	)
	if err = conn.Set(&memcache.Item{Key: key, Object: card, Flags: memcache.FlagJSON, Expiration: d.expireMC}); err != nil {
		err = errors.Wrap(err, key)
	}
	conn.Close()
	return
}
