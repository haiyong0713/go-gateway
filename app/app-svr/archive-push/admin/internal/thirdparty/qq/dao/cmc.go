package dao

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"go-common/library/log"

	qqModel "go-gateway/app/app-svr/archive-push/admin/internal/thirdparty/qq/model"
	"go-gateway/app/app-svr/archive-push/admin/internal/thirdparty/qq/util"
	"go-gateway/app/app-svr/archive-push/ecode"
)

const (
	PushPGCAdminURL         = "/cmc/pushPgcAdmin"
	ModifyPGCAdminURL       = "/cmc/modifyPgcAdmin"
	DetailAdminURL          = "/cmc/detailAdmin"
	UserContentListAdminURL = "/cmc/userContentListAdmin"
)

func (d *Dao) PushPGCAdmin(req *qqModel.PushPGCAdminReq) (reply *qqModel.PushPGCAdminReply, err error) {
	if d.httpClient == nil {
		d.httpClient = http.DefaultClient
	}
	if req == nil {
		return nil, ecode.PushRequestError
	}
	req.SExt9 = d.Cfg.CMC.SExt9
	var reqJSON []byte
	if reqJSON, err = json.Marshal(req); err != nil {
		log.Error("qq cmc: PushPGCAdmin Marshal Error (%v)", err)
		return
	}
	toRequestUrl := d.Cfg.CMC.Host + PushPGCAdminURL
	queryParams := url.Values{}
	queryParams.Set("ibiz", d.Cfg.CMC.IBIZ)
	queryParams.Set("source", d.Cfg.CMC.Source)
	now := time.Now().Unix()
	queryParams.Set("t", strconv.FormatInt(now, 10))
	sign := util.Sign(d.Cfg.CMC.Secret, d.Cfg.CMC.Source, d.Cfg.CMC.IBIZ, now)
	queryParams.Set("sign", sign)
	queryParams.Set("ctype", string(qqModel.CTypeVideo))
	toRequestUrl = toRequestUrl + "?" + queryParams.Encode()

	httpReq, _ := http.NewRequest("POST", toRequestUrl, strings.NewReader(string(reqJSON)))
	httpReq.Header.Set(qqModel.ContentTypeKey, qqModel.ContentTypeJSON)

	log.Info("qq cmc: PushPGCAdmin posting Data=%s", string(reqJSON))
	log.Info("qq cmc: PushPGCAdmin posting URL=%s", httpReq.URL.String())
	var resRaw *http.Response
	if resRaw, err = d.httpClient.Do(httpReq); err != nil {
		log.Error("qq cmc: PushPGCAdmin Do Request Error (%v)", err)
		return
	} else if resRaw.Close {
		log.Error("qq cmc: PushPGCAdmin response closed")
	} else if resRaw.Body != nil {
		defer resRaw.Body.Close()
		resBytes, _ := ioutil.ReadAll(resRaw.Body)
		log.Info("qq cmc: PushPGCAdmin Response: %+v", string(resBytes))
		reply = &qqModel.PushPGCAdminReply{}
		if err = json.Unmarshal(resBytes, reply); err != nil {
			log.Error("qq cmc: PushPGCAdmin Unmarshal response Error (%v)", err)
		}
	}
	return
}

func (d *Dao) ModifyPGCAdmin(docid string, mode qqModel.ModifyPGCAdminMode, req *qqModel.ModifyPGCAdminReq) (reply *qqModel.ModifyPGCAdminReply, err error) {
	if d.httpClient == nil {
		d.httpClient = http.DefaultClient
	}
	toRequestUrl := d.Cfg.CMC.Host + ModifyPGCAdminURL
	reqJSONStr := "{}"
	if req != nil {
		if reqJSON, _err := json.Marshal(req); _err != nil {
			log.Error("qq cmc: ModifyPGCAdmin Marshal Error (%v)", err)
			return
		} else {
			reqJSONStr = string(reqJSON)
		}
	}
	queryParams := url.Values{}
	queryParams.Set("ibiz", d.Cfg.CMC.IBIZ)
	queryParams.Set("source", d.Cfg.CMC.Source)
	now := time.Now().Unix()
	queryParams.Set("t", strconv.FormatInt(now, 10))
	sign := util.Sign(d.Cfg.CMC.Secret, d.Cfg.CMC.Source, d.Cfg.CMC.IBIZ, now)
	queryParams.Set("sign", sign)
	queryParams.Set("ctype", string(qqModel.CTypeVideo))
	queryParams.Set("id", docid)
	queryParams.Set("mode", strconv.FormatInt(int64(mode), 10))
	toRequestUrl = toRequestUrl + "?" + queryParams.Encode()

	httpReq, _ := http.NewRequest("POST", toRequestUrl, strings.NewReader(reqJSONStr))
	httpReq.Header.Set(qqModel.ContentTypeKey, qqModel.ContentTypeJSON)

	fmt.Printf("qq cmc: ModifyPGCAdmin posting URL=(%s)", httpReq.URL.String())
	log.Info("qq cmc: ModifyPGCAdmin posting URL=(%s)", httpReq.URL.String())
	var resRaw *http.Response
	if resRaw, err = d.httpClient.Do(httpReq); err != nil {
		log.Error("qq cmc: ModifyPGCAdmin Do Request Error (%v)", err)
		return
	} else if resRaw.Close {
		log.Error("qq cmc: ModifyPGCAdmin response closed")
	} else if resRaw.Body != nil {
		defer resRaw.Body.Close()
		resBytes, _ := ioutil.ReadAll(resRaw.Body)
		fmt.Printf("qq cmc: ModifyPGCAdmin Response: %+v", string(resBytes))
		log.Info("qq cmc: ModifyPGCAdmin Response: %+v", string(resBytes))
		reply = &qqModel.ModifyPGCAdminReply{}
		if err = json.Unmarshal(resBytes, reply); err != nil {
			log.Error("qq cmc: ModifyPGCAdmin Unmarshal response Error (%v)", err)
		}
	}
	return
}

func (d *Dao) DetailAdmin(query *qqModel.DetailAdminQuery) (reply *qqModel.DetailAdminReply, err error) {
	if d.httpClient == nil {
		d.httpClient = http.DefaultClient
	}
	toRequestUrl := d.Cfg.CMC.Host + DetailAdminURL
	queryParams := url.Values{}
	queryParams.Set("id", query.ID)
	queryParams.Set("ibiz", d.Cfg.CMC.IBIZ)
	queryParams.Set("source", d.Cfg.CMC.Source)
	now := time.Now().Unix()
	queryParams.Set("t", strconv.FormatInt(now, 10))
	sign := util.Sign(d.Cfg.CMC.Secret, d.Cfg.CMC.Source, d.Cfg.CMC.IBIZ, now)
	queryParams.Set("sign", sign)
	queryParams.Set("ctype", query.CType)
	toRequestUrl = toRequestUrl + "?" + queryParams.Encode()

	httpReq, _ := http.NewRequest("GET", fmt.Sprintf(toRequestUrl), strings.NewReader(""))
	httpReq.Header.Set(qqModel.ContentTypeKey, qqModel.ContentTypeJSON)

	fmt.Printf("qq cmc: DetailAdmin getting URL=(%s)", httpReq.URL.String())
	log.Info("qq cmc: DetailAdmin getting URL=%s", httpReq.URL.String())
	var resRaw *http.Response
	if resRaw, err = d.httpClient.Do(httpReq); err != nil {
		log.Error("qq cmc: DetailAdmin Do Request Error (%v)", err)
		return
	} else if resRaw.Close {
		log.Error("qq cmc: DetailAdmin response closed")
	} else if resRaw.Body != nil {
		defer resRaw.Body.Close()
		resBytes, _ := ioutil.ReadAll(resRaw.Body)
		fmt.Printf("qq cmc: DetailAdmin Response: %+v", string(resBytes))
		log.Info("qq cmc: DetailAdmin Response: %+v", string(resBytes))
		reply = &qqModel.DetailAdminReply{}
		if err = json.Unmarshal(resBytes, reply); err != nil {
			log.Error("qq cmc: DetailAdmin Unmarshal response Error (%v)", err)
		}
	}
	return
}

func (d *Dao) UserContentListAdmin(query *qqModel.UserContentListAdminQuery) (reply *qqModel.UserContentListAdminReply, err error) {
	if d.httpClient == nil {
		d.httpClient = http.DefaultClient
	}
	toRequestUrl := d.Cfg.CMC.Host + UserContentListAdminURL
	queryParams := url.Values{}
	queryParams.Set("ibiz", d.Cfg.CMC.IBIZ)
	queryParams.Set("source", d.Cfg.CMC.Source)
	now := time.Now().Unix()
	queryParams.Set("t", strconv.FormatInt(now, 10))
	sign := util.Sign(d.Cfg.CMC.Secret, d.Cfg.CMC.Source, d.Cfg.CMC.IBIZ, now)
	queryParams.Set("sign", sign)
	queryParams.Set("ctype", query.CType)
	queryParams.Set("creater", query.Creater)
	queryParams.Set("page", strconv.FormatInt(int64(query.Page), 10))
	queryParams.Set("pagesize", strconv.FormatInt(int64(query.PageSize), 10))
	toRequestUrl = toRequestUrl + "?" + queryParams.Encode()

	httpReq, _ := http.NewRequest("GET", toRequestUrl, strings.NewReader(""))
	httpReq.Header.Set(qqModel.ContentTypeKey, qqModel.ContentTypeJSON)

	fmt.Printf("qq cmc: UserContentListAdmin getting URL=(%s)", httpReq.URL.String())
	log.Info("qq cmc: UserContentListAdmin getting URL=%s", httpReq.URL.String())
	var resRaw *http.Response
	if resRaw, err = d.httpClient.Do(httpReq); err != nil {
		log.Error("qq cmc: UserContentListAdmin Do Request Error (%v)", err)
		return
	} else if resRaw.Close {
		log.Error("qq cmc: UserContentListAdmin response closed")
	} else if resRaw.Body != nil {
		defer resRaw.Body.Close()
		resBytes, _ := ioutil.ReadAll(resRaw.Body)
		fmt.Printf("qq cmc: UserContentListAdmin Response: %+v", string(resBytes))
		log.Info("qq cmc: UserContentListAdmin Response: %+v", string(resBytes))
		reply = &qqModel.UserContentListAdminReply{}
		if err = json.Unmarshal(resBytes, reply); err != nil {
			log.Error("qq cmc: UserContentListAdmin Unmarshal response Error (%v)", err)
		}
	}
	return
}
