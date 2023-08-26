package duertv

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-gateway/app/app-svr/app-car/job/model/duertv"

	"github.com/pkg/errors"
)

const (
	_duertvPush    = "/duertv/data/pushjson"
	_duertvPushUGC = "/duertv/data/pushugc"
)

func (d *Dao) Push(c context.Context, data []*duertv.DuertvPush, now time.Time) error {
	params := url.Values{}
	params.Set("code", d.pushCode(now))
	params.Set("t", strconv.FormatInt(now.Unix(), 10))
	duertvURL, err := url.Parse(d.duertvPush)
	if err != nil {
		return err
	}
	duertvURL.RawQuery = params.Encode()
	var res struct {
		Code int    `json:"status"`
		Msg  string `json:"msg"`
	}
	bytesData, err := json.Marshal(data)
	if err != nil {
		log.Error("json.Marshal error(%v)", err)
		return err
	}
	req, err := http.NewRequest("POST", duertvURL.String(), bytes.NewReader(bytesData))
	if err != nil {
		return errors.Wrap(err, duertvURL.String()+"&"+string(bytesData))
	}
	req.Header.Set("Content-Type", "application/json;charset=UTF-8")
	req.Header.Set("X-BACKEND-BILI-REAL-IP", "")
	if err = d.client.Do(c, req, &res); err != nil {
		return errors.Wrap(err, duertvURL.String()+"&"+string(bytesData))
	}
	if res.Code != ecode.OK.Code() {
		return errors.Wrap(ecode.Int(res.Code), "msg: "+res.Msg+" url: "+duertvURL.String()+"&"+string(bytesData))
	}
	log.Infoc(c, "push pgc : %s", JsonFlatStructToString(data))
	return nil
}

func (d *Dao) PushUGC(c context.Context, data *duertv.DuertvPushUGC, now time.Time) error {
	params := url.Values{}
	params.Set("code", d.pushCode(now))
	params.Set("t", strconv.FormatInt(now.Unix(), 10))
	duertvURL, err := url.Parse(d.duertvPushUGC)
	if err != nil {
		return err
	}
	duertvURL.RawQuery = params.Encode()
	var res struct {
		Code int    `json:"status"`
		Msg  string `json:"msg"`
	}
	bytesData, err := json.Marshal(data)
	if err != nil {
		log.Error("json.Marshal error(%v)", err)
		return err
	}
	req, err := http.NewRequest("POST", duertvURL.String(), bytes.NewReader(bytesData))
	if err != nil {
		return errors.Wrap(err, duertvURL.String()+"&"+string(bytesData))
	}
	req.Header.Set("Content-Type", "application/json;charset=UTF-8")
	req.Header.Set("X-BACKEND-BILI-REAL-IP", "")
	if err = d.client.Do(c, req, &res); err != nil {
		return errors.Wrap(err, duertvURL.String()+"&"+string(bytesData))
	}
	if res.Code != ecode.OK.Code() {
		return errors.Wrap(ecode.Int(res.Code), "msg: "+res.Msg+" url: "+duertvURL.String()+"&"+string(bytesData))
	}
	log.Warnc(c, "push ugc : %s", JsonFlatStructToString(data))
	return nil
}

func (d *Dao) pushCode(now time.Time) string {
	mh := md5.Sum([]byte(d.c.Duertv.Key + "#" + strconv.FormatInt(now.Unix(), 10)))
	return hex.EncodeToString(mh[:])
}

func JsonFlatStructToString(src interface{}) string {
	var builder strings.Builder
	if ss, err := json.Marshal(src); err == nil {
		builder.Write(ss)
	}
	return builder.String()
}
