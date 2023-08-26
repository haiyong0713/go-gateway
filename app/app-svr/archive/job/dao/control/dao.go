package control

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"net/url"
	"strings"
	"time"

	serGRPC "git.bilibili.co/bapis/bapis-go/content-flow-control/service"
	"go-gateway/app/app-svr/archive/job/conf"
)

var (
	_businessID = int32(1)
	_source     = "archive-service"
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
	if d.flowClient, err = serGRPC.NewClient(c.FlowClient); err != nil {
		panic(fmt.Sprintf("flow control newClient panic(%+v)", err))
	}
	return
}

func (d *Dao) GetInternalAttr(c context.Context, aid int64) ([]*serGRPC.InfoItem, error) {
	ts := time.Now().Unix()
	req := &serGRPC.FlowCtlInfoReq{
		Oid:        aid,
		BusinessId: _businessID,
		Source:     _source,
		Ts:         ts,
	}
	req.Sign = getSign(aid, ts, d.c.Custom.FlowSecret)
	rly, err := d.flowClient.InfoV2(c, req)
	if err != nil {
		return nil, err
	}
	if rly == nil { //稿件不禁止，不返回err
		return make([]*serGRPC.InfoItem, 0), nil
	}
	return rly.Items, nil
}

func getSign(aid, ts int64, secret string) string {
	params := url.Values{}
	params.Add("oid", fmt.Sprintf("%d", aid))
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

func (d *Dao) AIDsCursor(c context.Context, aid, ps int64) (*serGRPC.AIDsCursorReply, error) {
	req := &serGRPC.AIDsCursorReq{
		Source:   "archive-service",
		FlowId:   54,
		LastAid:  aid,
		PageSize: ps,
	}
	rly, err := d.flowClient.AIDsCursor(c, req)
	if err != nil {
		return nil, err
	}
	return rly, nil
}
