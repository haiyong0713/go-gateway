package dao

import (
	"context"
	"net/url"
	"strconv"

	xecode "go-common/library/ecode"
	"go-common/library/log"
)

const (
	_replyState   = "0" // 0: open, 1: close
	_gameOfficial = 32708316
)

// RegReply opens eports's reply.
func (d *Dao) RegReply(c context.Context, maid, adid int64, replyType string) (err error) {
	params := url.Values{}
	params.Set("adid", strconv.FormatInt(adid, 10))
	params.Set("mid", strconv.FormatInt(_gameOfficial, 10))
	params.Set("oid", strconv.FormatInt(maid, 10))
	params.Set("type", replyType)
	params.Set("state", _replyState)
	var res struct {
		Code int `json:"code"`
	}
	if err = d.replyClient.Post(c, d.replyURL, "", params, &res); err != nil {
		log.Error("d.replyClient.Post(%s) error(%v)", d.replyURL+"?"+params.Encode(), err)
		return
	}
	if res.Code != xecode.OK.Code() {
		log.Error("d.replyClient.Post(%s) error(%v)", d.replyURL+"?"+params.Encode(), err)
		err = xecode.Int(res.Code)
	}
	return
}

func (d *Dao) FixMatchUseJob(matchID int64) (err error) {
	params := url.Values{}
	params.Set("match_id", strconv.FormatInt(matchID, 10))
	params.Set("tp", "game")
	var res struct {
		Code int `json:"code"`
	}
	log.Info("FixMatchUseJob jobURL(%s) params(%s)", d.jobURL, params.Encode())
	if err = d.jobClient.Get(context.Background(), d.jobURL, "", params, &res); err != nil {
		log.Error("FixMatchUseJob match fix url(%s) error(%v)", d.jobURL+"?"+params.Encode(), err)
		return
	}
	if res.Code != xecode.OK.Code() {
		log.Error("FixMatchUseJob match fix url(%s) error(%v)", d.jobURL+"?"+params.Encode(), err)
		err = xecode.Int(res.Code)
	}
	return
}

func (d *Dao) FixBigUseJob(tp, sid int64) (err error) {
	params := url.Values{}
	params.Set("tp", strconv.FormatInt(tp, 10))
	params.Set("sid", strconv.FormatInt(sid, 10))
	var res struct {
		Code int `json:"code"`
	}
	log.Info("FixBigUseJob jobURL(%s) params(%s)", d.jobBigURL, params.Encode())
	if err = d.jobClient.Get(context.Background(), d.jobBigURL, "", params, &res); err != nil {
		log.Error("FixBigUseJob match fix url(%s) error(%v)", d.jobBigURL+"?"+params.Encode(), err)
		return
	}
	if res.Code != xecode.OK.Code() {
		log.Error("FixBigUseJob match fix url(%s) error(%v)", d.jobBigURL+"?"+params.Encode(), err)
		err = xecode.Int(res.Code)
	}
	return
}
