package dynamic

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/pkg/errors"
	"go-common/library/ecode"
	"go-common/library/log"
	httpx "go-common/library/net/http/blademaster"
	"go-common/library/net/metadata"

	"go-gateway/app/web-svr/native-page/interface/conf"
)

const (
	_createDynURI = "/dynamic_svr/v0/dynamic_svr/icreate"
)

type Dao struct {
	c      *conf.Config
	client *httpx.Client
	//动态相关http接口
	feedDynamicURL string
	briefDynURL    string
	createDynURL   string
}

func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		c:              c,
		client:         httpx.NewClient(c.HTTPDynamic),
		feedDynamicURL: c.Host.Dynamic + _feedDynamicURI,
		briefDynURL:    c.Host.Dynamic + _briefDynURI,
		createDynURL:   c.Host.Dynamic + _createDynURI,
	}
	return
}

func (d *Dao) CreateDynamic(c context.Context, content string, mid, pageID int64) (int64, error) {
	ip := metadata.String(c, metadata.RemoteIP)
	params := url.Values{}
	params.Set("uid", strconv.FormatInt(mid, 10))
	params.Set("content", content)
	params.Set("extension", dynamicExtension(pageID))
	params.Set("from", "create.up_activity")
	params.Set("type", "4")
	params.Set("audit_level", "100")
	params.Set("user_ip", ip)
	params.Set("user_port", metadata.String(c, metadata.RemotePort))
	params.Set("ts", strconv.FormatInt(time.Now().UnixNano()/1e6, 10))
	var res struct {
		Code int `json:"code"`
		Data struct {
			DynamicID int64 `json:"dynamic_id"`
		} `json:"data"`
	}
	if err := d.client.Post(c, d.createDynURL, ip, params, &res); err != nil {
		log.Error("Fail to request CreateDynamic, req=%+v error=%+v", params.Encode(), err)
		return 0, err
	}
	if res.Code != ecode.OK.Code() {
		err := errors.Wrap(ecode.Int(res.Code), d.createDynURL+"?"+params.Encode())
		log.Error("Fail to request CreateDynamic, req=%+v error=%+v", params.Encode(), err)
		return 0, err
	}
	return res.Data.DynamicID, nil
}

func dynamicExtension(pageID int64) string {
	return fmt.Sprintf(`{"flag_cfg":{"up_activity":{"up_activity_id":%d}},"activity":{"activity_id":%d,"activity_state":1}}`, pageID, pageID)
}
