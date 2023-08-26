package content_flow

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"go-common/library/ecode"

	"go-gateway/app/app-svr/app-view/interface/conf"

	api "git.bilibili.co/bapis/bapis-go/content-flow-control/service"

	"github.com/pkg/errors"
)

type Dao struct {
	conf                 *conf.Config
	flowControllerClient api.FlowControlClient
}

const (
	BusinessIdArchive = 1                //稿件
	Source            = "app_view_video" //来源
	SourceAppView     = "app-view"       //来源
)

func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		conf: c,
	}
	var err error
	if d.flowControllerClient, err = api.NewClient(c.FlowControllerClient); err != nil {
		panic(fmt.Sprintf("flowController NewClient not found err(%v)", err))
	}
	return
}

func (d *Dao) GetCtlInfo(c context.Context, aid int64) (*api.FlowCtlInfoReply, error) {
	now := time.Now().Unix()
	req := api.FlowCtlInfoReq{
		Oid:        aid,
		BusinessId: BusinessIdArchive,
		Source:     Source,
		Ts:         now,
	}
	res, err := d.flowControllerClient.Info(c, &req)
	if err != nil {
		err = errors.Wrapf(err, "%d", aid)
		return nil, err
	}
	if res == nil {
		return nil, ecode.NothingFound
	}
	return res, nil
}

func (d *Dao) GetCtlInfoV2(c context.Context, aid int64) (*api.FlowCtlInfoV2Reply, error) {
	now := time.Now().Unix()
	req := &api.FlowCtlInfoReq{
		Oid:        aid,
		BusinessId: BusinessIdArchive,
		Source:     SourceAppView,
		Ts:         now,
	}
	req.Sign = getSign(aid, now, d.conf.Custom.FlowControllerSecret, SourceAppView, BusinessIdArchive)
	res, err := d.flowControllerClient.InfoV2(c, req)
	if err != nil {
		err = errors.Wrapf(err, "%d", aid)
		return nil, err
	}
	if res == nil {
		return nil, ecode.NothingFound
	}
	return res, nil
}

func getSign(aid, ts int64, secret, source string, businessID int) string {
	params := url.Values{}
	params.Set("source", source)
	params.Set("oid", strconv.FormatInt(aid, 10)) //单个
	params.Set("business_id", strconv.Itoa(businessID))
	params.Set("ts", strconv.FormatInt(ts, 10))

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
