package business

import (
	httpx "go-common/library/net/http/blademaster"

	"go-gateway/app/web-svr/native-page/interface/conf"
)

type Dao struct {
	c      *conf.Config
	client *httpx.Client
	//企业号-商单相关http接口
	businessSourceURL  string
	businessProduceURL string
}

func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		c:                  c,
		client:             httpx.NewClient(c.HTTPBusiness),
		businessSourceURL:  c.Host.Business + _sourceURI,
		businessProduceURL: c.Host.Business + _productURI,
	}
	return
}
