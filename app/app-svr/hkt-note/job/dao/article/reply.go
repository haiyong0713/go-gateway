package article

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"go-common/library/log"
	"net/url"
	"strconv"
	"strings"
	"time"

	"go-common/library/ecode"

	"github.com/pkg/errors"
	reply "go-gateway/app/app-svr/hkt-note/job/model/note"
)

const (
	_tpReply      = "1"
	_sceneNote    = "note"
	_pathReplyAdd = "/x/internal/v2/reply/root/add"
	_pathReplyDel = "/x/internal/v2/reply/del"
)

func (d *Dao) ReplyAdd(c context.Context, mid, oid int64, replyCont string) error {
	params := url.Values{}
	params.Set("mid", strconv.FormatInt(mid, 10))
	params.Set("oid", strconv.FormatInt(oid, 10))
	params.Set("type", _tpReply)
	params.Set("biz_scene", _sceneNote)
	params.Set("message", replyCont)
	params.Set("appkey", d.c.HTTPClient.Key)
	params.Set("ts", strconv.FormatInt(time.Now().Unix(), 10))
	sign := signature(params, d.c.HTTPClient.Secret)
	params.Set("sign", sign)
	res := new(struct {
		Code int
	})
	log.Warn("artTest replyAdd params(%+v)", params)
	if err := d.client.Post(c, d.c.NoteCfg.Host.ReplyHost+_pathReplyAdd, "", params, &res); err != nil {
		return errors.Wrapf(err, "ReplyAdd params(%+v)", params)
	}
	if res.Code != ecode.OK.Code() {
		return errors.Wrapf(ecode.Int(res.Code), "ReplyAdd params(%+v)", params)
	}
	return nil
}
func (d *Dao) ReplyDel(c context.Context, mid, replyId int64) error {
	params := url.Values{}
	params.Set("mid", strconv.FormatInt(mid, 10))
	params.Set("rpids", strconv.FormatInt(replyId, 10))
	params.Set("type", _tpReply)
	res := new(struct {
		Code    int
		Message string
		Ttl     int
	})
	log.Warn("artTest replyDel params(%+v)", params)
	if err := d.client.Post(c, d.c.NoteCfg.Host.ReplyHost+_pathReplyDel, "", params, &res); err != nil {
		return errors.Wrapf(err, "ReplyDel params(%+v)", params)
	}
	if res.Code != ecode.OK.Code() {
		return errors.Wrapf(ecode.Int(res.Code), "ReplyDel params(%+v)", params)
	}
	return nil
}

func (d *Dao) ReplyAddWithRes(ctx context.Context, mid, oid int64, replyCont string) (res *reply.ReplyAddRes, err error) {
	params := url.Values{}
	params.Set("mid", strconv.FormatInt(mid, 10))
	params.Set("oid", strconv.FormatInt(oid, 10))
	params.Set("type", _tpReply)
	params.Set("biz_scene", _sceneNote)
	params.Set("message", replyCont)
	params.Set("appkey", d.c.HTTPClient.Key)
	params.Set("ts", strconv.FormatInt(time.Now().Unix(), 10))
	sign := signature(params, d.c.HTTPClient.Secret)
	params.Set("sign", sign)
	if err := d.client.Post(ctx, d.c.NoteCfg.Host.ReplyHost+_pathReplyAdd, "", params, &res); err != nil {
		log.Warnc(ctx, "ReplyAddWithRes res params(%+v)  mid %v oid %v res %v", params, mid, oid, res)
		return nil, errors.Wrapf(err, "ReplyAddWithRes params(%+v)  mid %v oid %v", params, mid, oid)
	}
	log.Warnc(ctx, "ReplyAddWithRes res params(%+v)  mid %v oid %v res %v", params, mid, oid, res)
	if res.Code != int64(ecode.OK.Code()) {
		return nil, errors.Wrapf(err, "ReplyAddWithRes params(%+v)  mid %v oid %v", params, mid, oid)
	}
	return res, nil
}

func signature(params url.Values, secret string) string {
	data := params.Encode()
	if strings.IndexByte(data, '+') > -1 {
		data = strings.Replace(data, "+", "%20", -1)
	}
	data = strings.ToLower(data)
	digest := md5.Sum([]byte(data + secret))
	return hex.EncodeToString(digest[:])
}
