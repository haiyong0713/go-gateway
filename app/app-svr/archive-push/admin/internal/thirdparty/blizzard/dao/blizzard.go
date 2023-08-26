package dao

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"go-common/library/log"

	bzModel "go-gateway/app/app-svr/archive-push/admin/internal/thirdparty/blizzard/model"
	"go-gateway/app/app-svr/archive-push/admin/internal/thirdparty/blizzard/util"
	"go-gateway/app/app-svr/archive-push/ecode"
)

const (
	VodAddURL              = "/action/external/bilibili/vod/add"
	ContentTypeKey         = "content-type"
	ContentTypeFormEncoded = "application/x-www-form-urlencoded"
)

func (d *Dao) VodAdd(req bzModel.VodAddReq) (reply *bzModel.VodAddReply, err error) {
	if d.httpClient == nil {
		d.httpClient = http.DefaultClient
	}
	if req.Title == "" {
		return nil, ecode.PushRequestError
	}
	req.Timestamp = time.Now().UnixNano() / 1e6 // 毫秒
	req.Sign = util.Sign(req, d.Cfg.Secret.Key)
	var reqForm bytes.Buffer
	params := url.Values{}
	params.Add("bvId", req.BVID)
	params.Add("category", req.Category)
	params.Add("description", req.Description)
	params.Add("duration", strconv.FormatInt(req.Duration, 10))
	params.Add("page", strconv.FormatInt(int64(req.Page), 10))
	params.Add("stage", req.Stage)
	params.Add("status", strconv.FormatInt(int64(req.Status), 10))
	params.Add("thumbnail", req.Thumbnail)
	params.Add("title", req.Title)
	params.Add("ts", strconv.FormatInt(req.Timestamp, 10))
	params.Add("sign", req.Sign)
	reqForm.WriteString(params.Encode())
	toRequestUrl := d.Cfg.Host.Host + VodAddURL

	httpReq, _ := http.NewRequest("POST", toRequestUrl, strings.NewReader(reqForm.String()))
	httpReq.Header.Set(ContentTypeKey, ContentTypeFormEncoded)

	log.Info("Blizzard Dao: VodAdd URL = %s posting Data=%s", httpReq.URL.String(), reqForm.String())
	var resRaw *http.Response
	if resRaw, err = d.httpClient.Do(httpReq); err != nil {
		log.Error("Blizzard Dao: VodAdd Do Request Error (%v)", err)
		return
	} else if resRaw.Close {
		log.Error("Blizzard Dao: VodAdd response closed")
	} else if resRaw.Body != nil {
		defer resRaw.Body.Close()
		resBytes, _ := ioutil.ReadAll(resRaw.Body)
		log.Info("Blizzard Dao: VodAdd Response: %s", string(resBytes))
		reply = &bzModel.VodAddReply{}
		if err = json.Unmarshal(resBytes, reply); err != nil {
			log.Error("Blizzard Dao: VodAdd Unmarshal response Error (%v)", err)
		}
	}
	return
}
