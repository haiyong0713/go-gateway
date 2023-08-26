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

	"go-common/library/ecode"
	"go-common/library/xstr"
	"go-gateway/app/app-svr/app-car/interface/conf"

	api "git.bilibili.co/bapis/bapis-go/content-flow-control/service"
	"github.com/pkg/errors"
)

// FlowControlInfoV2 单个获取管控信息
func (d *Dao) FlowControlInfoV2(ctx context.Context, aid int64, cfg *conf.FlowControl) ([]*api.InfoItem, error) {
	if cfg == nil {
		return nil, errors.Wrap(ecode.RequestErr, "FlowControlInfoV2 no cfg!")
	}
	ts := time.Now().Unix()
	params := url.Values{}
	params.Set("source", cfg.Source)
	params.Set("oid", strconv.FormatInt(aid, 10))
	params.Set("business_id", strconv.Itoa(cfg.BusinessID))
	params.Set("ts", strconv.FormatInt(ts, 10))
	req := &api.FlowCtlInfoReq{
		Oid:        aid,
		BusinessId: int32(cfg.BusinessID),
		Source:     cfg.Source,
		Sign:       flowControlSign(params, cfg.Secret),
		Ts:         ts,
	}
	resp, err := d.flowControlClient.InfoV2(ctx, req)
	if err != nil {
		return nil, errors.Wrapf(err, "FlowControlInfoV2 req=%+v", req)
	}
	return resp.Items, nil
}

// FlowControlInfosV2 批量获取管控信息
func (d *Dao) FlowControlInfosV2(ctx context.Context, aids []int64, cfg *conf.FlowControl) (map[int64]*api.FlowCtlInfoV2Reply, error) {
	if cfg == nil {
		return nil, errors.Wrap(ecode.RequestErr, "FlowControlInfosV2 no cfg!")
	}
	ts := time.Now().Unix()
	params := url.Values{}
	params.Set("source", cfg.Source)
	params.Set("oids", xstr.JoinInts(aids))
	params.Set("business_id", strconv.Itoa(cfg.BusinessID))
	params.Set("ts", strconv.FormatInt(ts, 10))
	req := &api.FlowCtlInfosReq{
		Oids:       aids,
		BusinessId: int32(cfg.BusinessID),
		Source:     cfg.Source,
		Sign:       flowControlSign(params, cfg.Secret),
		Ts:         ts,
	}

	resp, err := d.flowControlClient.InfosV2(ctx, req)
	if err != nil {
		return nil, errors.Wrapf(err, "FlowControlInfosV2 req=%+v, params=%s", req, params.Encode())
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
