package like

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"go-common/library/cache/redis"
	"go-common/library/log"
	"go-gateway/app/web-svr/activity/job/component"
	likemdl "go-gateway/app/web-svr/activity/job/model/like"

	"github.com/pkg/errors"
)

// userContriKey .
func userContriKey(mid int64) string {
	return fmt.Sprintf("contri_mid_%d", mid)
}

func (d *Dao) AddCacheContributionUser(c context.Context, data []*likemdl.ContributionUser) error {
	if len(data) == 0 {
		return errors.New("contribution user nil")
	}
	var (
		key  string
		keys []string
		args = redis.Args{}
	)
	conn := component.GlobalRedis.Conn(c)
	defer conn.Close()
	for _, userContri := range data {
		bs, err := json.Marshal(userContri)
		if err != nil {
			log.Error("json.Marshal err(%v)", err)
			continue
		}
		key = userContriKey(userContri.Mid)
		keys = append(keys, key)
		args = args.Add(key).Add(string(bs))
	}
	if err := conn.Send("MSET", args...); err != nil {
		return err
	}
	count := 1
	for _, v := range keys {
		count++
		if err := conn.Send("EXPIRE", v, 8640000); err != nil {
			return err
		}
	}
	if err := conn.Flush(); err != nil {
		return err
	}
	for i := 0; i < count; i++ {
		if _, err := conn.Receive(); err != nil {
			return err
		}
	}
	return nil
}

func (d *Dao) AddCacheTotalRank(ctx context.Context, sid int64, topArcs string) error {
	if _, err := component.GlobalRedis.Do(ctx, "SETEX", fmt.Sprintf("lg_contri:%d:%s", sid, time.Now().Format("20060102")), 86400*7, topArcs); err != nil {
		return err
	}
	return nil
}

// UserContribution
func (d *Dao) UserContribution(c context.Context, mid int64) (data *likemdl.ContributionUser, err error) {
	var (
		key = userContriKey(mid)
		bs  []byte
	)
	if bs, err = redis.Bytes(component.GlobalRedis.Do(c, "GET", key)); err != nil {
		if err == redis.ErrNil {
			err = nil
			return
		}
		log.Error("UserContribution redis.String(conn.Do(GET,%s)) error(%v)", key, err)
		return
	}
	if err = json.Unmarshal(bs, &data); err != nil {
		log.Error("UserContribution json.Unmarshal(%s) error(%v)", string(bs), err)
	}
	return
}
