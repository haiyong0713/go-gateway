package like

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"time"

	"go-common/library/cache/redis"
	"go-common/library/database/elastic"
	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/net/metadata"
	"go-common/library/xstr"
	"go-gateway/app/web-svr/activity/interface/model/like"

	"github.com/pkg/errors"
)

const _resAuditURI = "/x/internal/resource/res/audit"

func keyLikeMidTotal(mid int64, sids []int64) string {
	return fmt.Sprintf("like_m_t_%d_%s", mid, xstr.JoinInts(sids))
}

// UserMatchCheck user match check.
func (d *Dao) UserMatchCheck(c context.Context, mid int64, sids []int64) (sid int64, err error) {
	actResult := new(struct {
		Result []*struct {
			ID        int64 `json:"id"`
			Wid       int64 `json:"wid"`
			Sid       int64 `json:"sid"`
			Type      int   `json:"type"`
			Mid       int64 `json:"mid"`
			State     int   `json:"state"`
			Copyright int   `json:"copyright"`
		} `json:"result"`
		Page *like.Page `json:"page"`
	})
	req := d.es.NewRequest(_activity).Index(_activity).
		Fields("id", "wid", "sid", "type", "mid", "state", "copyright").
		WhereEq("state", 1).
		WhereEq("mid", mid).
		WhereIn("sid", sids).
		Order("ctime", elastic.OrderDesc).
		Pn(1).
		Ps(1)
	if err = req.Scan(c, actResult); err != nil {
		log.Error("UserMatchCheck req.Scan mid(%d) error(%v)", mid, err)
		return
	}
	if len(actResult.Result) > 0 && actResult.Result[0] != nil {
		sid = actResult.Result[0].Sid
	}
	return
}

// CacheLikeMidTotal .
func (d *Dao) RawLikeMidTotal(c context.Context, mid int64, sids []int64) (total int64, err error) {
	result := new(struct {
		Page *like.Page `json:"page"`
	})
	req := d.es.NewRequest(_activity).Index(_activity).WhereIn("sid", sids).WhereEq("mid", mid).WhereEq("state", 1)
	if err = req.Scan(c, result); err != nil {
		log.Error("UserMatchCheck req.Scan mid(%d) error(%v)", mid, err)
		return
	}
	if result.Page != nil {
		total = result.Page.Total
	}
	return
}

// CacheLikeMidTotal .
func (d *Dao) CacheLikeMidTotal(c context.Context, mid int64, sids []int64) (total int64, err error) {
	key := keyLikeMidTotal(mid, sids)
	conn := d.redis.Get(c)
	defer conn.Close()
	if total, err = redis.Int64(conn.Do("GET", key)); err != nil {
		if err == redis.ErrNil {
			err = nil
		} else {
			err = errors.Wrapf(err, "conn.Do(GET, %s)", key)
		}
	}
	return
}

// AddCacheLikeMidTotal .
func (d *Dao) AddCacheLikeMidTotal(c context.Context, mid, total int64, sids []int64) (err error) {
	key := keyLikeMidTotal(mid, sids)
	conn := d.redis.Get(c)
	defer conn.Close()
	if err = conn.Send("SET", key, total); err != nil {
		err = errors.Wrapf(err, "conn.Send(SET, %s, %d)", key, total)
		return
	}
	if err = conn.Send("EXPIRE", key, d.likeMidTotalExpire); err != nil {
		err = errors.Wrapf(err, "conn.Send(EXPIRE, %s, %d)", key, total)
		return
	}
	if err = conn.Flush(); err != nil {
		err = errors.Wrap(err, "conn.Flush")
		return
	}
	for i := 0; i < 2; i++ {
		if _, err = conn.Receive(); err != nil {
			err = errors.Wrapf(err, "conn.Receive(%d)", i+1)
			return
		}
	}
	return
}

func (d *Dao) ResAudit(c context.Context) (data map[string][]int64, err error) {
	var res struct {
		Code int                `json:"code"`
		Data map[string][]int64 `json:"data"`
	}
	if err = d.client.Get(c, d.resAuditURL, metadata.String(c, metadata.RemoteIP), url.Values{}, &res); err != nil {
		return
	}
	if res.Code != ecode.OK.Code() {
		err = errors.Wrap(ecode.Int(res.Code), d.resAuditURL)
		return
	}
	data = res.Data
	return
}

func (d *Dao) SpecialData(c context.Context, uri string, timeStamp int64) (ret *like.SpecialArcList, err error) {
	params := url.Values{}
	params.Set("_", strconv.FormatInt(timeStamp, 10))
	ret = new(like.SpecialArcList)
	if err = d.singleClient.Get(c, uri, "", params, &ret); err != nil {
		log.Error("SpecialData d.client.Get(%s) error(%+v)", uri+"?"+params.Encode(), err)
		return
	}
	return
}

func (d *Dao) CacheStupidArcs(ctx context.Context, sid int64) ([]*like.StupidVv, error) {
	conn := d.redis.Get(ctx)
	defer conn.Close()
	args := redis.Args{}
	for i := 0; i <= 49; i++ {
		key := fmt.Sprintf("stupid:arc:%d:%d:%s", sid, i, stupidKey())
		args = append(args, key)
	}
	bss, err := redis.ByteSlices(conn.Do("MGET", args...))
	if err != nil {
		return nil, err
	}
	arcs := make([]*like.StupidVv, 0)
	for _, bs := range bss {
		if bs == nil {
			continue
		}
		arc := make([]*like.StupidVv, 0)
		if err = json.Unmarshal(bs, &arc); err != nil {
			log.Error("CacheStupidArcs json.Unmarshal(%s) error(%v)", string(bs), err)
			err = nil
			continue
		}
		arcs = append(arcs, arc...)
	}
	return arcs, nil
}

func stupidKey() string {
	if time.Now().Unix() >= 1592240400 {
		return "2020061600"
	}
	return time.Now().Add(time.Hour * -1).Format("2006010215")
}

func (dao *Dao) CacheStupidTotal(ctx context.Context, sid int64) (int64, error) {
	conn := dao.redis.Get(ctx)
	defer conn.Close()
	key := fmt.Sprintf("stupid:vv:%d:%s", sid, stupidKey())
	total, err := redis.Int64(conn.Do("GET", key))
	if err != nil {
		return 0, err
	}
	return total, nil
}
