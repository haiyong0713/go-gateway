package http

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"go-common/library/ecode"
	log "go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/http/blademaster/render"
	"go-common/library/net/metadata"
	"go-gateway/app/app-svr/app-wall/interface/model/telecom"
)

const (
	_tkey  = "6b7e8b8a"
	_tckey = "9ed226d9"
)

// ordersSync
func telecomOrdersSync(c *bm.Context) {
	data, err := requestJSONTelecom(c.Request)
	if err != nil {
		telecomMessage(c, err.Error())
		return
	}
	switch v := data.(type) {
	case *telecom.TelecomOrderJson:
		if err = telecomSvc.InOrdersSync(c, metadata.String(c, metadata.RemoteIP), v); err != nil {
			log.Error("telecomSvc.InOrdersSync error (%v)", err)
			telecomMessage(c, err.Error())
			return
		}
		telecomMessage(c, "1")
	case *telecom.TelecomRechargeJson:
		if v == nil {
			telecomMessage(c, ecode.NothingFound.Error())
			return
		}
		if err = telecomSvc.InRechargeSync(c, metadata.String(c, metadata.RemoteIP), v.Detail); err != nil {
			log.Error("telecomSvc.InOrdersSync error (%v)", err)
			telecomMessage(c, err.Error())
			return
		}
		telecomMessage(c, "1")
	}
}

// telecomMsgSync
func telecomMsgSync(c *bm.Context) {
	data, err := requestJSONTelecomMsg(c.Request)
	if err != nil {
		telecomMessage(c, err.Error())
		return
	}
	if err = telecomSvc.TelecomMessageSync(c, metadata.String(c, metadata.RemoteIP), data); err != nil {
		log.Error("telecomSvc.TelecomMessageSync error (%v)", err)
		telecomMessage(c, err.Error())
		return
	}
	telecomMessage(c, "1")
}

func telecomPay(c *bm.Context) {
	res := map[string]interface{}{}
	params := c.Request.Form
	phoneDES := params.Get("phone")
	phone, err := phoneDesToInt(phoneDES, _tkey)
	if err != nil {
		log.Error("phoneDesToInt error(%v)", err)
		res["message"] = ""
		returnDataJSON(c, res, ecode.RequestErr)
		return
	}
	isRepeatOrderStr := params.Get("is_repeat_order")
	isRepeatOrder, err := strconv.Atoi(isRepeatOrderStr)
	if err != nil {
		log.Error("isRepeatOrder strconv.Atoi error(%v)", err)
		res["message"] = ""
		returnDataJSON(c, res, ecode.RequestErr)
		return
	}
	payChannelStr := params.Get("pay_channel")
	payChannel, err := strconv.Atoi(payChannelStr)
	if err != nil {
		log.Error("payChannel strconv.Atoi error(%v)", err)
		res["message"] = ""
		returnDataJSON(c, res, ecode.RequestErr)
		return
	}
	payActionStr := params.Get("pay_action")
	payAction, err := strconv.Atoi(payActionStr)
	if err != nil {
		log.Error("payAction strconv.Atoi error(%v)", err)
		res["message"] = ""
		returnDataJSON(c, res, ecode.RequestErr)
		return
	}
	orderIDStr := params.Get("orderid")
	orderID, _ := strconv.ParseInt(orderIDStr, 10, 64)
	data, msg, err := telecomSvc.TelecomPay(c, phone, isRepeatOrder, payChannel, payAction, orderID, metadata.String(c, metadata.RemoteIP))
	if err != nil {
		log.Error("telecomSvc.TelecomPay error(%v)", err)
		res["message"] = msg
		returnDataJSON(c, res, err)
		return
	}
	res["data"] = data
	returnDataJSON(c, res, nil)
}

func cancelRepeatOrder(c *bm.Context) {
	res := map[string]interface{}{}
	params := c.Request.Form
	phoneDES := params.Get("phone")
	phone, err := phoneDesToInt(phoneDES, _tkey)
	if err != nil {
		log.Error("phoneDesToInt error(%v)", err)
		res["message"] = ""
		returnDataJSON(c, res, ecode.RequestErr)
		return
	}
	msg, err := telecomSvc.CancelRepeatOrder(c, phone)
	if err != nil {
		res["message"] = msg
		returnDataJSON(c, res, err)
		return
	}
	res["message"] = msg
	returnDataJSON(c, res, nil)
}

func orderList(c *bm.Context) {
	res := map[string]interface{}{}
	params := c.Request.Form
	phoneDES := params.Get("phone")
	phone, err := phoneDesToInt(phoneDES, _tkey)
	if err != nil {
		log.Error("phoneDesToInt error(%v)", err)
		res["message"] = ""
		returnDataJSON(c, res, ecode.RequestErr)
		return
	}
	orderIDStr := params.Get("orderid")
	orderID, _ := strconv.ParseInt(orderIDStr, 10, 64)
	data, msg, err := telecomSvc.OrderList(c, orderID, phone)
	if err != nil {
		log.Error("telecomSvc.OrderList error(%v)", err)
		res["message"] = msg
		returnDataJSON(c, res, err)
		return
	}
	res["data"] = data
	returnDataJSON(c, res, nil)
}

func phoneFlow(c *bm.Context) {
	res := map[string]interface{}{}
	params := c.Request.Form
	phoneDES := params.Get("phone")
	phone, err := phoneDesToInt(phoneDES, _tkey)
	if err != nil {
		log.Error("phoneDesToInt error(%v)", err)
		res["message"] = ""
		returnDataJSON(c, res, ecode.RequestErr)
		return
	}
	orderIDStr := params.Get("orderid")
	orderID, _ := strconv.ParseInt(orderIDStr, 10, 64)
	data, msg, err := telecomSvc.PhoneFlow(c, orderID, phone)
	if err != nil {
		log.Error("telecomSvc.PhoneFlow error(%v)", err)
		res["message"] = msg
		returnDataJSON(c, res, err)
		return
	}
	res["data"] = data
	returnDataJSON(c, res, nil)
}

func orderConsent(c *bm.Context) {
	res := map[string]interface{}{}
	params := c.Request.Form
	phoneDES := params.Get("phone")
	phone, err := phoneDesToInt(phoneDES, _tkey)
	if err != nil {
		log.Error("phoneDesToInt error(%v)", err)
		res["message"] = ""
		returnDataJSON(c, res, ecode.RequestErr)
		return
	}
	captcha := params.Get("captcha")
	orderIDStr := params.Get("orderid")
	orderID, _ := strconv.ParseInt(orderIDStr, 10, 64)
	data, msg, err := telecomSvc.OrderConsent(c, phone, orderID, captcha)
	if err != nil {
		log.Error("telecomSvc.OrderConsent error(%v)", err)
		res["message"] = msg
		returnDataJSON(c, res, err)
		return
	}
	res["data"] = data
	returnDataJSON(c, res, nil)
}

func phoneSendSMS(c *bm.Context) {
	res := map[string]interface{}{}
	params := c.Request.Form
	phoneDES := params.Get("phone")
	phone, err := phoneDesToInt(phoneDES, _tkey)
	if err != nil {
		log.Error("phoneDesToInt error(%v)", err)
		res["message"] = ""
		returnDataJSON(c, res, ecode.RequestErr)
		return
	}
	err = telecomSvc.PhoneSendSMS(c, phone)
	if err != nil {
		log.Error("telecomSvc.PhoneSendSMS error(%v)", err)
		res["code"] = err
		res["message"] = ""
		returnDataJSON(c, res, err)
		return
	}
	res["message"] = ""
	returnDataJSON(c, res, nil)
}

func phoneVerification(c *bm.Context) {
	res := map[string]interface{}{}
	params := c.Request.Form
	phoneDES := params.Get("phone")
	phone, err := phoneDesToInt(phoneDES, _tkey)
	if err != nil {
		log.Error("phoneDesToInt error(%v)", err)
		res["message"] = ""
		returnDataJSON(c, res, ecode.RequestErr)
		return
	}
	captcha := params.Get("captcha")
	data, err, msg := telecomSvc.PhoneCode(c, phone, captcha, time.Now())
	if err != nil {
		log.Error("telecomSvc.PhoneCode error(%v)", err)
		res["message"] = msg
		returnDataJSON(c, res, err)
		return
	}
	res["data"] = data
	res["message"] = msg
	returnDataJSON(c, res, nil)
}

func orderState(c *bm.Context) {
	res := map[string]interface{}{}
	params := c.Request.Form
	orderIDStr := params.Get("orderid")
	orderID, err := strconv.ParseInt(orderIDStr, 10, 64)
	if err != nil {
		log.Error("orderID strconv.ParseInt error(%v)", err)
		res["message"] = ""
		returnDataJSON(c, res, ecode.RequestErr)
		return
	}
	data, msg, err := telecomSvc.OrderState(c, orderID)
	if err != nil {
		log.Error("telecomSvc.OrderState error(%v)", err)
		res["message"] = msg
		returnDataJSON(c, res, err)
		return
	}
	res["data"] = data
	res["message"] = msg
	returnDataJSON(c, res, nil)
}

// requestJSONTelecom
// nolint:gomnd
func requestJSONTelecom(request *http.Request) (res interface{}, err error) {
	var (
		telecomOrder    *telecom.TelecomOrderJson
		telecomRecharge *telecom.TelecomRechargeJson
	)
	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		log.Error("unicom_ioutil.ReadAll error (%v)", err)
		return
	}
	defer request.Body.Close()
	if len(body) == 0 {
		err = ecode.RequestErr
		return
	}
	log.Info("telecom orders json body(%s)", body)
	if err = json.Unmarshal(body, &telecomOrder); err == nil && telecomOrder != nil {
		if telecomOrder.ResultType != 2 {
			res = telecomOrder
			return
		}
	}
	if err = json.Unmarshal(body, &telecomRecharge); err == nil {
		res = telecomRecharge
		return
	}
	log.Error("telecom json.Unmarshal error (%v)", err)
	return
}

// requestJSONTelecomMsg
func requestJSONTelecomMsg(request *http.Request) (res *telecom.TelecomMessageJSON, err error) {
	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		log.Error("unicom_ioutil.ReadAll error (%v)", err)
		return
	}
	defer request.Body.Close()
	if len(body) == 0 {
		err = ecode.RequestErr
		return
	}
	log.Info("telecom json msg body(%s)", body)
	if err = json.Unmarshal(body, &res); err != nil {
		log.Error("telecom Message json.Unmarshal error (%v)", err)
		return
	}
	return
}

// telecomMessage
func telecomMessage(c *bm.Context, code string) {
	// response header
	c.Writer.Header().Set("Content-Type", "text; charset=UTF-8")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	_, _ = c.Writer.Write([]byte(code))
}

// phoneDesToInt des to int
// nolint:gomnd
func phoneDesToInt(phoneDes, key string) (phoneInt int, err error) {
	var (
		_aesKey = []byte(key)
	)
	bs, err := base64.StdEncoding.DecodeString(phoneDes)
	if err != nil {
		log.Error("base64.StdEncoding.DecodeString(%s) error(%v)", phoneDes, err)
		err = ecode.RequestErr
		return
	}
	if bs, err = telecomSvc.DesDecrypt(bs, _aesKey); err != nil {
		log.Error("phone s.DesDecrypt error(%v)", err)
		return
	}
	var phoneStr string
	if len(bs) > 11 {
		phoneStr = string(bs[:11])
	} else {
		phoneStr = string(bs)
	}
	phoneInt, err = strconv.Atoi(phoneStr)
	if err != nil {
		log.Error("phoneDesToInt phoneStr:%v error(%v)", phoneStr, err)
		err = ecode.RequestErr
		return
	}
	return
}

func requestJSONTelecomCard(request *http.Request) (res *telecom.CardOrderJson, err error) {
	var (
		order *telecom.CardOrderJson
	)
	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		log.Error("telecom_ioutil.ReadAll error (%v)", err)
		return
	}
	defer request.Body.Close()
	log.Info("telecom card orders json body(%s)", body)
	if len(body) == 0 {
		err = ecode.RequestErr
		return
	}
	if err = json.Unmarshal(body, &order); err == nil && order != nil {
		res = order
	}
	return
}

func telecomCardOrdersSync(c *bm.Context) {
	var (
		now = time.Now()
	)
	data, err := requestJSONTelecomCard(c.Request)
	if err != nil {
		returnTelecomDataJSON(c, data, now, err)
		return
	}
	if data == nil || data.Head == nil || data.Biz == nil {
		returnTelecomDataJSON(c, data, now, ecode.RequestErr)
		return
	}
	ipStr := metadata.String(c, metadata.RemoteIP)
	if err = telecomSvc.InCardOrderSync(c, data, ipStr); err != nil {
		returnTelecomDataJSON(c, data, now, err)
		return
	}
	returnTelecomDataJSON(c, data, now, nil)
}

func telecomCardOrder(c *bm.Context) {
	params := c.Request.Form
	usermob := params.Get("usermob")
	phone, err := phoneDesToInt(usermob, _tckey)
	if err != nil {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	data, err := telecomSvc.CardOrder(c, phone)
	log.Warn("telecomCardOrder usermob:%s phone:%d data:%+v error:%+v", usermob, phone, data, err)
	c.JSON(data, err)
}

func telecomCardCodeOrder(c *bm.Context) {
	var (
		params = c.Request.Form
		now    = time.Now()
	)
	usermob := params.Get("usermob")
	captcha := params.Get("captcha")
	phone, err := phoneDesToInt(usermob, _tckey)
	if err != nil {
		log.Error("phoneDesToInt error(%v)", err)
		c.JSON(nil, ecode.RequestErr)
		return
	}
	data, err := telecomSvc.PhoneCardCode(c, phone, captcha, now)
	log.Warn("telecomCardCodeOrder usermob:%s phone:%d data:%+v error:%+v", usermob, phone, data, err)
	c.JSON(data, err)
}

// telecomMessage
func returnTelecomDataJSON(c *bm.Context, t *telecom.CardOrderJson, now time.Time, err error) {
	if t == nil || t.Head == nil {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	bcode := ecode.Cause(err)
	data := map[string]interface{}{
		"head": map[string]interface{}{
			"transactionId": t.Head.SysCode + now.Format("0601021504") + uniqueID(),
			"resTime":       now.Format("2006-01-02 15:04:05"),
			"code":          bcode.Code(),
			"err":           nil,
			"attach":        t.Head.Attach,
		},
		"biz": nil,
	}
	c.Render(http.StatusOK, render.MapJSON(data))
}

func uniqueID() string {
	b := make([]byte, 48)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return ""
	}
	return getMd5String(base64.URLEncoding.EncodeToString(b))
}

func getMd5String(s string) string {
	h := md5.New()
	_, _ = h.Write([]byte(s))
	return hex.EncodeToString(h.Sum(nil))
}

func phoneCardVerification(c *bm.Context) {
	params := c.Request.Form
	usermob := params.Get("usermob")
	phone, err := phoneDesToInt(usermob, _tckey)
	if err != nil {
		log.Error("phoneDesToInt error(%v)", err)
		c.JSON(nil, ecode.RequestErr)
		return
	}
	err = telecomSvc.PhoneCardSendSMS(c, phone)
	if err != nil {
		log.Error("telecomSvc.PhoneCardSendSMS error(%v)", err)
		c.JSON(nil, ecode.RequestErr)
		return
	}
	c.JSON(nil, nil)
}

func vipPacksLog(c *bm.Context) {
	params := c.Request.Form
	starttimeStr := params.Get("starttime")
	pnStr := params.Get("pn")
	pn, err := strconv.Atoi(pnStr)
	if err != nil {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	if pn < 1 {
		pn = 1
	}
	timeLayout := "2006-01"
	loc, _ := time.LoadLocation("Local")
	startTime, err := time.ParseInLocation(timeLayout, starttimeStr, loc)
	if err != nil {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	c.JSON(telecomSvc.VipPackLog(c, startTime, time.Now(), pn, metadata.String(c, metadata.RemoteIP)))
}

func telecomActiveState(ctx *bm.Context) {
	param := &telecom.UserActiveParam{}
	if err := ctx.Bind(param); err != nil {
		return
	}
	phone, err := phoneDesToInt(param.Usermob, _tckey)
	if err != nil {
		ctx.JSON(nil, ecode.RequestErr)
		return
	}
	midInter, ok := ctx.Get("mid")
	if ok {
		param.Mid = midInter.(int64)
	}
	header := ctx.Request.Header
	param.Buvid = header.Get("Buvid")
	data, err := telecomSvc.ActiveState(ctx, phone, param.Captcha)
	telecomSvc.UserActiveLog(param, data, err)
	ctx.JSON(data, err)
}
