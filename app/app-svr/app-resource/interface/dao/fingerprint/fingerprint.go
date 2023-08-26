package fingerprint

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"strconv"

	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/metadata"
	"go-gateway/app/app-svr/app-resource/interface/conf"
	fpmdl "go-gateway/app/app-svr/app-resource/interface/model/fingerprint"
)

// Dao is notice dao.
type Dao struct {
	fingerprint string
	client      *bm.Client
}

// New new a notice dao.
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		client:      bm.NewClient(c.HTTPClient),
		fingerprint: c.Host.DP + "/hakase/v1/profile",
	}
	return
}

func (d *Dao) Fingerprint(c context.Context, platfrom, buvid string, mid int64, body []byte) (f *fpmdl.Fingerprint, err error) {
	ip := metadata.String(c, metadata.RemoteIP)
	req, err := http.NewRequest("POST", d.fingerprint, bytes.NewReader(body))
	if err != nil {
		log.Error("Fingerprint http.NewRequest() error(%v)", err)
		return
	}
	req.Header.Set("Mid", strconv.FormatInt(mid, 10))
	req.Header.Set("IP", ip)
	req.Header.Set("Platform", platfrom)
	req.Header.Set("Buvid", buvid)
	req.Header.Add("Content-Type", "application/json")
	var res struct {
		Code         int    `json:"code"`
		Message      string `json:"message"`
		BiliDeviceID string `json:"biliDeviceId"`
	}
	if err = d.client.Do(c, req, &res); err != nil {
		log.Error("httpCli.Do(%s) error(%v)", d.fingerprint, err)
		return
	}
	if res.Code != 0 {
		err = fmt.Errorf("Fingerprint api failed(%d) msg(%v)", res.Code, res.Message)
		log.Error("Fingerprint(%s) res code(%d) msg(%v)", d.fingerprint, res.Code, res.Message)
		return
	}
	f = &fpmdl.Fingerprint{BiliDeviceID: res.BiliDeviceID}
	return
}
