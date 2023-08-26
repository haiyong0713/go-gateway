package http

import (
	"crypto/md5"
	"encoding/hex"
	"net/url"
	"strings"

	"go-common/library/conf/paladin.v2"
	"go-common/library/ecode"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
)

func verify(ctx *bm.Context, appKeys map[string]string, checkSDK bool) error {
	req := ctx.Request
	params := req.Form
	if req.Method == "POST" {
		// Give priority to sign in url query, otherwise check sign in post form.
		q := req.URL.Query()
		if q.Get("sign") != "" {
			params = q
		}
	}

	// check timestamp is not empty (TODO : Check if out of some seconds.., like 100s)
	if params.Get("ts") == "" {
		log.Error("ts is empty")
		return ecode.RequestErr
	}

	sign := params.Get("sign")
	params.Del("sign")
	defer params.Set("sign", sign)
	sappkey := params.Get("appkey")
	identifier := params.Get("platform")
	if checkSDK {
		sdkOdentifier := params.Get("sdk_identifier")
		if sdkOdentifier != "" {
			identifier = sdkOdentifier
		}
	}
	happkey := appkey(identifier)
	if sappkey != happkey {
		log.Error("Get appkey: %s, expect %s", sappkey, happkey)
		return ecode.Unauthorized
	}
	secret, ok := appKeys[identifier]
	if !ok {
		return ecode.Unauthorized
	}
	if hsign := Sign(params, sappkey, secret, true); hsign != sign {
		if hsign1 := Sign(params, sappkey, secret, false); hsign1 != sign {
			log.Error("Get sign: %s, expect %s", sign, hsign)
			return ecode.SignCheckErr
		}
	}
	return nil
}

// Verify will inject into handler func as verify required
func Verify(ac *paladin.Map, checkSDK bool) func(ctx *bm.Context) {
	return func(ctx *bm.Context) {
		var appKeys map[string]string
		if err := ac.Get("appKeys").UnmarshalTOML(&appKeys); err != nil {
			log.Error("Get appKeys error:%+v", err)
			ctx.JSON(nil, ecode.Unauthorized)
			ctx.Abort()
			return
		}
		if err := verify(ctx, appKeys, checkSDK); err != nil {
			ctx.JSON(nil, err)
			ctx.Abort()
			return
		}
	}
}

func appkey(sdkIdentifier string) string {
	const _salt = "bE589Rsj9W9q"
	digest := md5.Sum([]byte(sdkIdentifier + _salt))
	return hex.EncodeToString(digest[:])[8:24]
}

// Sign is used to sign form params by given condition.
func Sign(params url.Values, appkey, secret string, lower bool) string {
	data := params.Encode()
	if strings.IndexByte(data, '+') > -1 {
		data = strings.Replace(data, "+", "%20", -1)
	}
	if lower {
		data = strings.ToLower(data)
	}
	digest := md5.Sum([]byte(data + secret))
	return hex.EncodeToString(digest[:])
}
