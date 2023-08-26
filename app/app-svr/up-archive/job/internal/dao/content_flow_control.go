package dao

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"go-common/library/xstr"
	"net/url"
	"strconv"
	"strings"
	"time"

	cfcgrpc "git.bilibili.co/bapis/bapis-go/content-flow-control/service"
)

const (
	_source      = "up-archive-service"
	_business_id = 1
)

func (d *dao) ContentFlowControlInfo(ctx context.Context, aid int64) ([]*cfcgrpc.ForbiddenItem, error) {
	ts := time.Now().Unix()
	params := url.Values{}
	params.Set("source", _source)
	params.Set("oid", strconv.FormatInt(aid, 10))
	params.Set("business_id", strconv.Itoa(_business_id))
	params.Set("ts", strconv.FormatInt(ts, 10))
	in := &cfcgrpc.FlowCtlInfoReq{
		Oid:        aid,
		BusinessId: _business_id,
		Source:     _source,
		Sign:       d.getSign(params),
		Ts:         ts,
	}
	reply, err := d.cfcGRPC.Info(ctx, in)
	if err != nil {
		return nil, err
	}
	return reply.ForbiddenItems, nil
}

func (d *dao) ContentFlowControlInfos(ctx context.Context, aids []int64) (map[int64][]*cfcgrpc.ForbiddenItem, error) {
	ts := time.Now().Unix()
	params := url.Values{}
	params.Set("source", _source)
	params.Set("oids", xstr.JoinInts(aids)) //批量
	params.Set("business_id", strconv.Itoa(_business_id))
	params.Set("ts", strconv.FormatInt(ts, 10))
	in := &cfcgrpc.FlowCtlInfosReq{
		Oids:       aids,
		BusinessId: _business_id,
		Source:     _source,
		Sign:       d.getSign(params),
		Ts:         ts,
	}
	reply, err := d.cfcGRPC.Infos(ctx, in)
	if err != nil {
		return nil, err
	}
	res := make(map[int64][]*cfcgrpc.ForbiddenItem, len(reply.ForbiddenItemMap))
	for mid, infoReply := range reply.ForbiddenItemMap {
		res[mid] = infoReply.ForbiddenItems
	}
	return res, nil
}

func (d *dao) getSign(params url.Values) string {
	tmp := params.Encode()
	if strings.IndexByte(tmp, '+') > -1 {
		tmp = strings.Replace(tmp, "+", "%20", -1)
	}
	var buf bytes.Buffer
	buf.WriteString(tmp)
	buf.WriteString(d.secret)
	mh := md5.Sum(buf.Bytes())
	return hex.EncodeToString(mh[:])
}
