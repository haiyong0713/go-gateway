package http

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"go-common/library/ecode"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/web-svr/web-goblin/interface/model/wechat"
)

func qrcode(c *bm.Context) {
	v := new(struct {
		JSON string `form:"json" validate:"required"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	qrcode, err := srvWechat.Qrcode(c, v.JSON)
	if err != nil {
		c.JSON(nil, err)
		return
	}
	c.Bytes(http.StatusOK, "", qrcode)
}

func push(c *bm.Context) {
	v := new(wechat.PushArg)
	if err := c.Bind(v); err != nil {
		return
	}
	bytes, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		log.Error("push ioutil.ReadAll() error(%v)", err)
		c.JSON(nil, ecode.RequestErr)
		return
	}
	userMsg := new(wechat.Msg)
	if err := json.Unmarshal(bytes, &userMsg); err != nil {
		log.Error("push json.Unmarshal(%s) error(%v)", bytes, err)
		c.JSON(nil, ecode.RequestErr)
		return
	}
	if err := srvWechat.Push(c, v, userMsg); err != nil {
		log.Error("%+v", err)
	}
	c.Bytes(http.StatusOK, "", []byte("success"))
}
