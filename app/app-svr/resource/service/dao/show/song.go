package show

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"go-common/library/log"
)

const (
	_songList = "/web/song/upper"
)

// songList https://info.bilibili.co/pages/viewpage.action?pageId=80628892#Music%E2%99%AA%E9%9F%B3%E9%A2%91%E6%8E%A5%E5%8F%A3%E6%96%87%E6%A1%A3%EF%BC%88musicc%EF%BC%89WEB%E7%AB%AF-UP%E4%B8%BB%E6%AD%8C%E6%9B%B2%E7%A8%BF%E4%BB%B6%E5%88%97%E8%A1%A8
// 业务方非go项目，无法用基础库请求，必须使用原生请求
func (d *Dao) IsSongUploader(c context.Context, uid, order int64) (is bool, err error) {
	var ret struct {
		Code int       `json:"code"`
		Data *struct{} `json:"data"`
	}
	baseUrl := d.c.Host.Song + _songList
	AllUrl := fmt.Sprintf("%s?uid=%d&order=%d&pn=0&ps=10", baseUrl, uid, order)
	req, err := http.NewRequestWithContext(c, "GET", AllUrl, nil)
	if err != nil {
		is = true
		log.Error("http.NewRequestWithContext %v", err)
		return
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		is = true
		log.Error("http.DefaultClient.Do %v", err)
		return
	}
	defer resp.Body.Close()
	resp.Header.Add("Content-Type", "application/json;charset=UTF-8")
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		is = true
		log.Error("ioutil.ReadAll %v", err)
		return
	}
	if err = json.Unmarshal(body, &ret); err != nil {
		is = true
		log.Error("json.Unmarshal %s", body)
		return
	}
	if ret.Code != 0 || ret.Data != nil {
		is = true
	}
	return
}
