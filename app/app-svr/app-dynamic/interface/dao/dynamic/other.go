package dynamic

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"go-common/library/ecode"
	"go-common/library/log"
	dynmdl "go-gateway/app/app-svr/app-dynamic/interface/model/dynamic"

	"github.com/pkg/errors"
)

const (
	_emojiURL    = "/x/internal/emote/by/text"
	_userLikeURL = "/x/internal/thumbup/item_has_like_recent"
)

func (d *Dao) GetEmoji(c context.Context, emojis []string) (map[string]*dynmdl.EmojiItem, error) {
	params := url.Values{}
	params.Set("texts", strings.Join(emojis, ","))
	params.Set("business", "dynamic")
	emojiURL := d.emoji
	var ret struct {
		Code int           `json:"code"`
		Msg  string        `json:"msg"`
		Data *dynmdl.Emoji `json:"data"`
	}
	if err := d.client.Get(c, emojiURL, "", params, &ret); err != nil {
		log.Errorc(c, "getEmoji http GET(%s) failed, params:(%s), error(%+v)", emojiURL, params.Encode(), err)
		return nil, err
	}
	if ret.Code != 0 {
		log.Errorc(c, "getEmoji http GET(%s) failed, params:(%s), code: %v, msg: %v", emojiURL, params.Encode(), ret.Code, ret.Msg)
		err := errors.Wrapf(ecode.Int(ret.Code), "getEmoji url(%v) code(%v) msg(%v)", emojiURL, ret.Code, ret.Msg)
		return nil, err
	}
	return ret.Data.Emote, nil
}

func (d *Dao) UserLike(c context.Context, mids []int64, business map[string][]*dynmdl.LikeBusiItem) (map[string][]*dynmdl.UserLikeItem, error) {
	params := &dynmdl.LikeBusiReq{
		Mids:       mids,
		Businesses: business,
	}
	b := &bytes.Buffer{}
	if err := json.NewEncoder(b).Encode(params); err != nil {
		log.Errorc(c, "UserLike json.NewEncoder() error(%+v)", err)
		return nil, err
	}
	fmt.Printf("UserLike params: %v\n", b.String())
	userLikeURL := d.userLike
	req, err := http.NewRequest(http.MethodPost, userLikeURL, b)
	if err != nil {
		log.Errorc(c, "UserLike error(%v)", err)
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	var ret struct {
		Code int                               `json:"code"`
		Msg  string                            `json:"msg"`
		Data map[string][]*dynmdl.UserLikeItem `json:"data"`
	}
	if err := d.client.Do(c, req, &ret); err != nil {
		log.Errorc(c, "UserLike http POST(%s) failed, params:(%s), error(%+v)", userLikeURL, b.String(), err)
		return nil, err
	}
	if ret.Code != 0 {
		log.Errorc(c, "UserLike http POST(%s) failed, params:(%s), code: %v, msg: %v", userLikeURL, b.String(), ret.Code, ret.Msg)
		err := errors.Wrapf(ecode.Int(ret.Code), "UserLike url(%v) code(%v) msg(%v)", userLikeURL, ret.Code, ret.Msg)
		return nil, err
	}
	return ret.Data, nil
}
