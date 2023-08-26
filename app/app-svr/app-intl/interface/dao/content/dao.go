package content

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
	"go-gateway/app/app-svr/app-intl/interface/conf"

	cfcgrpc "git.bilibili.co/bapis/bapis-go/content-flow-control/service"

	"github.com/pkg/errors"
)

// Dao is content-flow-control dao
type Dao struct {
	flowControlClient cfcgrpc.FlowControlClient
	conf              *conf.Config
}

// New initial content-flow-control dao
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		conf: c,
	}
	var err error
	if d.flowControlClient, err = cfcgrpc.NewClient(c.FlowControlClient); err != nil {
		panic(err)
	}
	return
}

func (d *Dao) ContentFlowControlInfosV2(ctx context.Context, oids []int64) (*cfcgrpc.FlowCtlInfosV2Reply, error) {
	ts := time.Now().Unix()
	params := url.Values{}
	params.Set("source", d.conf.CfcSvrConfig.Source)
	params.Set("business_id", strconv.FormatInt(d.conf.CfcSvrConfig.BusinessID, 10))
	params.Set("ts", strconv.FormatInt(ts, 10))
	params.Set("oids", collection.JoinSliceInt(oids, ","))
	req := &cfcgrpc.FlowCtlInfosReq{
		Oids:       oids,
		BusinessId: int32(d.conf.CfcSvrConfig.BusinessID),
		Source:     d.conf.CfcSvrConfig.Source,
		Sign:       getSign(params, d.conf.CfcSvrConfig.Secret),
		Ts:         ts,
	}
	reply, err := d.flowControlClient.InfosV2(ctx, req)
	if err != nil {
		return nil, errors.WithStack(err)
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
