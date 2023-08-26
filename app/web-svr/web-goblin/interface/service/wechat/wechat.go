package wechat

import (
	"context"
	// nolint:gosec
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"sort"
	"strconv"
	"strings"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-gateway/app/web-svr/web-goblin/interface/model/wechat"
)

const (
	_linkType    = "link"
	_msgTypeText = "text"
	_msgTypeMini = "miniprogrampage"
)

// Qrcode get qrcode from wechat.
func (s *Service) Qrcode(c context.Context, arg string) (qrcode []byte, err error) {
	var accessToken *wechat.AccessToken
	if accessToken, err = s.dao.AccessToken(c); err != nil {
		log.Error("Qrcode s.dao.AccessToken error(%v) arg(%s)", err, arg)
		return
	}
	qrcode, err = s.dao.Qrcode(c, accessToken.AccessToken, arg)
	return
}

// Push push wechat service msg.
func (s *Service) Push(c context.Context, param *wechat.PushArg, userMsg *wechat.Msg) (err error) {
	if !checkWechatSignature(param.Timestamp, s.c.Wechat.PushToken, param.Nonce, param.Signature) {
		log.Warn("Push checkWechatSignature fail param(%+v),userMsg(%+v)", param, userMsg)
		err = ecode.RequestErr
		return
	}
	var (
		accessToken *wechat.AccessToken
		sendBytes   []byte
		sendReply   bool
	)
	switch userMsg.MsgType {
	case _msgTypeText:
		if userMsg.Content != "" {
			if param.Openid == "" {
				log.Warn("Push param.Openid empty param(%+v),userMsg(%+v)", param, userMsg)
				err = ecode.RequestErr
				return
			}
			sendReply = true
		}
	case _msgTypeMini:
		sendReply = true
	}
	if !sendReply {
		return
	}
	if accessToken, err = s.dao.AccessToken(c); err != nil {
		log.Error("Push s.dao.AccessToken error(%v)", err)
		return
	}
	sendArg := &wechat.SendMsg{
		Touser:  param.Openid,
		Msgtype: _linkType,
		Link:    s.c.Wechat.LinkMsg,
	}
	if sendBytes, err = json.Marshal(sendArg); err != nil {
		return
	}
	return s.dao.SendMessage(c, accessToken.AccessToken, sendBytes)
}

func checkWechatSignature(ts int64, token, nonce, sign string) bool {
	tsStr := strconv.FormatInt(ts, 10)
	shStr := wechatSign(tsStr, token, nonce)
	return shStr == sign
}

func wechatSign(tsStr, token, nonce string) string {
	tmp := []string{token, tsStr, nonce}
	sort.Strings(tmp)
	tmpStr := strings.Join(tmp, "")
	sh := sha1.Sum([]byte(tmpStr))
	return hex.EncodeToString(sh[:])
}
