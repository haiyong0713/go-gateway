package archive

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"net/url"
	"strconv"
	"strings"
	"time"

	"go-common/library/utils/collection"

	api "git.bilibili.co/bapis/bapis-go/content-flow-control/service"
	"github.com/pkg/errors"
)

func (d *Dao) FlowControlInfoV2(ctx context.Context, aid int64) ([]*api.InfoItem, error) {
	ts := time.Now().Unix()
	params := url.Values{}
	params.Set("source", d.c.FlowControl.Source)
	params.Set("oid", strconv.FormatInt(aid, 10))
	params.Set("business_id", strconv.Itoa(d.c.FlowControl.BusinessID))
	params.Set("ts", strconv.FormatInt(ts, 10))
	req := &api.FlowCtlInfoReq{
		Oid:        aid,
		BusinessId: int32(d.c.FlowControl.BusinessID),
		Source:     d.c.FlowControl.Source,
		Sign:       flowControlSign(params, d.c.FlowControl.Secret),
		Ts:         ts,
	}
	resp, err := d.flowControlClient.InfoV2(ctx, req)
	if err != nil {
		return nil, errors.Wrapf(err, "FlowControlInfoV2 req=%+v", req)
	}
	return resp.Items, nil
}

func (d *Dao) FlowControlInfosV2(ctx context.Context, aids []int64) (map[int64]*api.FlowCtlInfoV2Reply, error) {
	ts := time.Now().Unix()
	params := url.Values{}
	params.Set("source", d.c.FlowControl.Source)
	params.Set("oids", collection.JoinSliceInt(aids, ","))
	params.Set("business_id", strconv.Itoa(d.c.FlowControl.BusinessID))
	params.Set("ts", strconv.FormatInt(ts, 10))
	req := &api.FlowCtlInfosReq{
		Oids:       aids,
		BusinessId: int32(d.c.FlowControl.BusinessID),
		Source:     d.c.FlowControl.Source,
		Sign:       flowControlSign(params, d.c.FlowControl.Secret),
		Ts:         ts,
	}
	resp, err := d.flowControlClient.InfosV2(ctx, req)
	if err != nil {
		return nil, errors.Wrapf(err, "FlowControlInfoV2 req=%+v", req)
	}
	return resp.ItemsMap, nil
}

func flowControlSign(params url.Values, secret string) string {
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
