package control

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"net/url"
	"strings"
	"sync"
	"time"

	serGRPC "git.bilibili.co/bapis/bapis-go/content-flow-control/service"
	"go-common/library/log"
	"go-common/library/sync/errgroup.v2"
	"go-common/library/utils/collection"
	"go-gateway/app/app-svr/app-show/interface/conf"
	"go-gateway/app/app-svr/app-show/interface/model/rank"
)

var (
	_businessID = int32(1)
	_source     = "app-show"
	_maxAids    = 30
)

// Dao is dao.
type Dao struct {
	c          *conf.Config
	flowClient serGRPC.FlowControlClient
}

func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		c: c,
	}
	var err error
	if d.flowClient, err = serGRPC.NewClient(c.FlowRPC); err != nil {
		panic(fmt.Sprintf("flow control newClient panic(%+v)", err))
	}
	return
}

func (d *Dao) GetInternalAttr(c context.Context, aids []int64) (map[int64]*rank.InnerAttr, error) {
	ts := time.Now().Unix()
	req := &serGRPC.FlowCtlInfosReq{
		Oids:       aids,
		BusinessId: _businessID,
		Source:     _source,
		Ts:         ts,
	}
	req.Sign = getSign(aids, ts, d.c.Custom.FlowSecret)
	rly, err := d.flowClient.InfosV2(c, req)
	if err != nil {
		return nil, err
	}
	innerAttr := make(map[int64]*rank.InnerAttr)
	if rly == nil || rly.ItemsMap == nil { //稿件不禁止，不返回err
		return innerAttr, nil
	}
	for k, v := range rly.ItemsMap {
		if v == nil {
			continue
		}
		innerAttr[k] = rank.ChangeInnerAttr(v.Items)
	}
	return innerAttr, nil
}

func getSign(aids []int64, ts int64, secret string) string {
	params := url.Values{}
	params.Add("oids", collection.JoinSliceInt(aids, ","))
	params.Add("business_id", fmt.Sprintf("%d", _businessID))
	params.Add("source", _source)
	params.Add("ts", fmt.Sprintf("%d", ts))
	tmp := params.Encode()
	if strings.IndexByte(tmp, '+') > -1 {
		tmp = strings.Replace(tmp, "+", "%20", -1)
	}
	var buf bytes.Buffer
	buf.WriteString(tmp)
	buf.WriteString(secret)
	mh := md5.Sum(buf.Bytes())
	return hex.EncodeToString(mh[:])
}

func (d *Dao) CircleReqInternalAttr(ctx context.Context, aids []int64) (aidMap map[int64]*rank.InnerAttr) {
	var (
		aidsLen = len(aids)
		mutex   = sync.Mutex{}
	)
	aidMap = make(map[int64]*rank.InnerAttr, aidsLen)
	gp := errgroup.WithContext(ctx)
	for i := 0; i < aidsLen; i += _maxAids {
		var partAids []int64
		if i+_maxAids > aidsLen {
			partAids = aids[i:]
		} else {
			partAids = aids[i : i+_maxAids]
		}
		//获取失败，降级处理
		gp.Go(func(ctx context.Context) error {
			tmpRes, err := d.GetInternalAttr(ctx, partAids)
			if err != nil {
				log.Error("d.GetInternalAttr(%v) error(%v)", partAids, err)
				return nil
			}
			if tmpRes == nil {
				log.Error("CircleReqInternalAttr is nil(%v)", partAids)
				return nil
			}
			if len(tmpRes) > 0 {
				mutex.Lock()
				for aid, arc := range tmpRes {
					if arc == nil {
						continue
					}
					aidMap[aid] = arc
				}
				mutex.Unlock()
			}
			return nil
		})
	}
	_ = gp.Wait()
	return
}
