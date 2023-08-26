package reply

import (
	"context"
	"encoding/json"
	"net/url"
	"strconv"

	"go-common/library/ecode"
	httpx "go-common/library/net/http/blademaster"
	"go-common/library/net/metadata"
	"go-gateway/app/app-svr/app-car/interface/conf"
	"go-gateway/app/app-svr/app-car/interface/model/reply"

	"github.com/pkg/errors"
)

const (
	_replymain  = "/x/internal/v2/reply/main"
	_replychild = "/x/internal/v2/reply/reply"
)

type Dao struct {
	client     *httpx.Client
	replyMain  string
	replyChild string
}

func New(c *conf.Config) *Dao {
	d := &Dao{
		client:     httpx.NewClient(c.HTTPClient),
		replyMain:  c.Host.APICo + _replymain,
		replyChild: c.Host.APICo + _replychild,
	}
	return d
}

func (d *Dao) ReplyMain(ctx context.Context, ps int, oid, otype, mode, next, mid int64, ext *reply.ReplyExtra) (*reply.ReplyList, error) {
	var (
		ip = metadata.String(ctx, metadata.RemoteIP)
	)
	params := url.Values{}
	params.Set("oid", strconv.FormatInt(oid, 10))
	params.Set("mid", strconv.FormatInt(mid, 10))
	params.Set("type", strconv.FormatInt(otype, 10))
	params.Set("mode", strconv.FormatInt(mode, 10))
	params.Set("next", strconv.FormatInt(next, 10))
	params.Set("ps", strconv.Itoa(ps))
	if ext != nil {
		if extraJSON, err := json.Marshal(ext); err == nil {
			params.Set("extra", string(extraJSON))
		}
	}
	var res struct {
		Code int              `json:"code"`
		Data *reply.ReplyList `json:"data"`
	}
	req, err := d.client.NewRequest("GET", d.replyMain, ip, params)
	if err != nil {
		return nil, err
	}
	if err = d.client.Do(ctx, req, &res); err != nil {
		return nil, err
	}
	if res.Code != ecode.OK.Code() {
		return nil, errors.Wrap(ecode.Int(res.Code), d.replyMain+"?"+params.Encode())
	}
	if res.Data == nil {
		return nil, ecode.NothingFound
	}
	return res.Data, nil
}

func (d *Dao) ReplyChild(ctx context.Context, pn, ps int, oid, otype, root, jump, mode, next, mid int64) (*reply.ReplyList, error) {
	var (
		ip = metadata.String(ctx, metadata.RemoteIP)
	)
	params := url.Values{}
	params.Set("oid", strconv.FormatInt(oid, 10))
	params.Set("mid", strconv.FormatInt(mid, 10))
	params.Set("type", strconv.FormatInt(otype, 10))
	params.Set("root", strconv.FormatInt(root, 10))
	if jump > 0 {
		params.Set("jump", strconv.FormatInt(jump, 10))
	}
	params.Set("mode", strconv.FormatInt(mode, 10))
	params.Set("pn", strconv.Itoa(pn))
	params.Set("ps", strconv.Itoa(ps))
	var res struct {
		Code int              `json:"code"`
		Data *reply.ReplyList `json:"data"`
	}
	req, err := d.client.NewRequest("GET", d.replyChild, ip, params)
	if err != nil {
		return nil, err
	}
	if err = d.client.Do(ctx, req, &res); err != nil {
		return nil, err
	}
	if res.Code != ecode.OK.Code() {
		return nil, errors.Wrap(ecode.Int(res.Code), d.replyChild+"?"+params.Encode())
	}
	if res.Data == nil {
		return nil, ecode.NothingFound
	}
	return res.Data, nil
}
