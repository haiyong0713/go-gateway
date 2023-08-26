package dao

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"net/url"
	"strconv"
	"strings"
	"time"

	cfcgrpc "git.bilibili.co/bapis/bapis-go/content-flow-control/service"
	"go-common/library/utils/collection"
	log "go-gateway/app/app-svr/fawkes/service/tools/logger"

	"github.com/pkg/errors"
)

func (d *Dao) ContentFlowControlInfoV2(ctx context.Context, aid int64) (*cfcgrpc.FlowCtlInfoV2Reply, error) {
	ts := time.Now().Unix()
	params := url.Values{}
	params.Set("source", d.conf.CfcSvrConfig.Source)
	params.Set("business_id", strconv.FormatInt(d.conf.CfcSvrConfig.BusinessID, 10))
	params.Set("ts", strconv.FormatInt(ts, 10))
	params.Set("oid", strconv.FormatInt(aid, 10))
	req := &cfcgrpc.FlowCtlInfoReq{
		Oid:        aid,
		BusinessId: int32(d.conf.CfcSvrConfig.BusinessID),
		Source:     d.conf.CfcSvrConfig.Source,
		Sign:       getSign(params, d.conf.CfcSvrConfig.Secret),
		Ts:         ts,
	}
	reply, err := d.cfcGRPC.InfoV2(ctx, req)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if reply == nil {
		log.Warn("contentFlowControlInfoV2 info is empty aid:%d", aid)
		return nil, nil
	}
	return reply, nil
}

func getSign(params url.Values, secret string) string {
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

func (d *Dao) ContentFlowControlInfosV2(ctx context.Context, aids []int64) (*cfcgrpc.FlowCtlInfosV2Reply, error) {
	ts := time.Now().Unix()
	params := url.Values{}
	params.Set("source", d.conf.CfcSvrConfig.Source)
	params.Set("business_id", strconv.FormatInt(d.conf.CfcSvrConfig.BusinessID, 10))
	params.Set("ts", strconv.FormatInt(ts, 10))
	params.Set("oids", collection.JoinSliceInt(aids, ","))
	req := &cfcgrpc.FlowCtlInfosReq{
		Oids:       aids,
		BusinessId: int32(d.conf.CfcSvrConfig.BusinessID),
		Source:     d.conf.CfcSvrConfig.Source,
		Sign:       getSign(params, d.conf.CfcSvrConfig.Secret),
		Ts:         ts,
	}
	reply, err := d.cfcGRPC.InfosV2(ctx, req)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if reply == nil {
		log.Warn("contentFlowControlInfosV2 info is empty aids:%d", aids)
		return nil, nil
	}
	return reply, nil
}
