package unicom

import (
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"encoding/hex"
	// nolint:gosec
	"crypto/des"
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-gateway/app/app-svr/app-wall/interface/model/unicom"

	"github.com/pkg/errors"
)

const (
	_cpid          = "bilibl"
	_spid          = "979"
	_apptype       = "2"
	_broadbandPass = "9ed226d9"
	// url
	_orderURL       = "/videoif/order.do"
	_ordercancelURL = "/videoif/cancelOrder.do"
	_sendsmscodeURL = "/videoif/sendSmsCode.do"
	_smsNumberURL   = "/videoif/smsNumber.do"
	// unicom
	_unicomFlowExchangeURL = "/openservlet"
	_unicomVerifyURL       = "/hzzxrightsaccquirenew/BilibiliCard/verifyCardKindEncrypt"
	_unicomFlowTryoutURL   = "/orderFlow/flowTryout"
)

// Order unicom order
func (d *Dao) Order(c context.Context, usermob, channel string, ordertype int) (data *unicom.BroadbandOrder, msg string, err error) {
	params := url.Values{}
	params.Set("cpid", _cpid)
	params.Set("spid", _spid)
	params.Set("ordertype", strconv.Itoa(ordertype))
	params.Set("userid", usermob)
	params.Set("apptype", _apptype)
	if channel != "" {
		params.Set("channel", channel)
	}
	var res struct {
		Code string `json:"resultcode"`
		Msg  string `json:"errorinfo"`
		*unicom.BroadbandOrder
	}
	if err = d.broadbandHTTPGet(c, d.orderURL, params, &res); err != nil {
		log.Error("unicom order url(%v) error(%v)", d.orderURL+"?"+params.Encode(), err)
		return
	}
	b, _ := json.Marshal(&res)
	log.Info("unicom order url(%v) response(%s)", d.orderURL+"?"+params.Encode(), b)
	if res.Code != "0" {
		err = ecode.String(res.Code)
		msg = res.Msg
		log.Error("unicom order url(%v) code(%s) Msg(%s)", d.orderURL+"?"+params.Encode(), res.Code, res.Msg)
		return
	}
	data = res.BroadbandOrder
	return
}

// CancelOrder unicom cancel order
func (d *Dao) CancelOrder(c context.Context, usermob string, spid int) (data *unicom.BroadbandOrder, msg string, err error) {
	params := url.Values{}
	params.Set("cpid", _cpid)
	params.Set("spid", strconv.Itoa(spid))
	params.Set("userid", usermob)
	params.Set("apptype", _apptype)
	var res struct {
		Code string `json:"resultcode"`
		Msg  string `json:"errorinfo"`
		*unicom.BroadbandOrder
	}
	if err = d.broadbandHTTPGet(c, d.ordercancelURL, params, &res); err != nil {
		log.Error("unicom cancel order url(%s) error(%v)", d.ordercancelURL+"?"+params.Encode(), err)
		return
	}
	b, _ := json.Marshal(&res)
	log.Info("unicom cancel order url(%s) response(%s)", d.ordercancelURL+"?"+params.Encode(), b)
	if res.Code != "0" {
		err = ecode.String(res.Code)
		msg = res.Msg
		log.Error("unicom cancel order url(%v) code(%s) Msg(%s)", d.orderURL+"?"+params.Encode(), res.Code, res.Msg)
		return
	}
	data = res.BroadbandOrder
	return
}

// SendSmsCode unicom sms code
func (d *Dao) SendSmsCode(c context.Context, phone string) (msg string, err error) {
	var (
		key       = []byte(_broadbandPass)
		phoneByte = []byte(phone)
		userid    string
	)
	userid, err = d.desEncrypt(phoneByte, key)
	if err != nil {
		log.Error("d.desEncrypt error(%v)", err)
		return
	}
	params := url.Values{}
	params.Set("cpid", _cpid)
	params.Set("userid", string(userid))
	params.Set("apptype", _apptype)
	var res struct {
		Code string `json:"resultcode"`
		Msg  string `json:"errorinfo"`
	}
	if err = d.unicomHTTPGet(c, d.sendsmscodeURL, params, &res); err != nil {
		log.Error("unicom sendsmscode url(%v) error(%v)", d.sendsmscodeURL+"?"+params.Encode(), err)
		return
	}
	b, _ := json.Marshal(&res)
	log.Info("unicom sendsmscode url(%v) response(%s)", d.sendsmscodeURL+"?"+params.Encode(), b)
	if res.Code != "0" {
		err = ecode.String(res.Code)
		msg = res.Msg
		log.Error("unicom sendsmscode url(%v) code(%s) Msg(%s)", d.sendsmscodeURL+"?"+params.Encode(), res.Code, res.Msg)
		return
	}
	return
}

// SmsNumber unicom sms usermob
func (d *Dao) SmsNumber(c context.Context, phone string, code int) (usermob string, msg string, err error) {
	var (
		key       = []byte(_broadbandPass)
		phoneByte = []byte(phone)
		userid    string
	)
	userid, err = d.desEncrypt(phoneByte, key)
	if err != nil {
		log.Error("d.desEncrypt error(%v)", err)
		return
	}
	params := url.Values{}
	params.Set("cpid", _cpid)
	params.Set("userid", userid)
	params.Set("vcode", strconv.Itoa(code))
	params.Set("apptype", _apptype)
	var res struct {
		Code    string `json:"resultcode"`
		Usermob string `json:"userid"`
		Msg     string `json:"errorinfo"`
	}
	if err = d.unicomHTTPGet(c, d.smsNumberURL, params, &res); err != nil {
		log.Error("unicom smsNumberURL url(%v) error(%v)", d.smsNumberURL+"?"+params.Encode(), err)
		return
	}
	b, _ := json.Marshal(&res)
	log.Info("unicom sendsmsnumber url(%v) response(%s)", d.smsNumberURL+"?"+params.Encode(), b)
	if res.Code != "0" {
		err = ecode.String(res.Code)
		msg = res.Msg
		log.Error("unicom sendsmsnumber url(%v) code(%s) Msg(%s)", d.smsNumberURL+"?"+params.Encode(), res.Code, res.Msg)
		return
	}
	usermob = res.Usermob
	return
}

// FlowExchange flow exchange
// nolint:gomnd
func (d *Dao) FlowExchange(c context.Context, phone int, flowcode string, requestNo int64, ts time.Time) (orderID, outorderID, msg string, err error) {
	outorderIDStr := "bili" + ts.Format("20060102") + strconv.FormatInt(requestNo%10000000, 10)
	if len(outorderIDStr) > 22 {
		outorderIDStr = outorderIDStr[:22]
	}
	param := url.Values{}
	param.Set("appkey", d.c.Unicom.UnicomAppKey)
	param.Set("apptx", strconv.FormatInt(requestNo, 10))
	param.Set("flowexchangecode", flowcode)
	param.Set("method", d.c.Unicom.UnicomAppMethodFlow)
	param.Set("outorderid", outorderIDStr)
	param.Set("timestamp", ts.Format("2006-01-02 15:04:05"))
	param.Set("usernumber", strconv.Itoa(phone))
	urlVal := param.Encode() + "&" + d.sign(d.urlParams(param))
	var res struct {
		Code       string `json:"respcode"`
		Msg        string `json:"respdesc"`
		OrderID    string `json:"orderid"`
		OutorderID string `json:"outorderid"`
	}
	if err = d.unicomHTTPGet(c, d.unicomFlowExchangeURL+"?"+urlVal, nil, &res); err != nil {
		err = errors.Wrapf(ecode.Error(ecode.ServerErr, "联通接口维护中,请稍后再试"), "%v", err)
		return "", "", "", err
	}
	if res.Code != "0000" {
		err = errors.Wrap(ecode.Error(ecode.String(res.Code), res.Msg), d.unicomFlowExchangeURL+"?"+urlVal)
		return "", "", res.Msg, err
	}
	return res.OrderID, res.OutorderID, res.Msg, nil
}

// FlowPre unicom phone flow pre
func (d *Dao) FlowPre(c context.Context, phone int, requestNo int64, ts time.Time) (msg string, err error) {
	param := url.Values{}
	param.Set("appkey", d.c.Unicom.UnicomAppKey)
	param.Set("apptx", strconv.FormatInt(requestNo, 10))
	param.Set("method", d.c.Unicom.UnicomMethodFlowPre)
	param.Set("timestamp", ts.Format("2006-01-02 15:04:05"))
	param.Set("usernumber", strconv.Itoa(phone))
	urlVal := param.Encode() + "&" + d.sign(d.urlParams(param))
	var res struct {
		Code   string `json:"respcode"`
		Notice string `json:"noticecontent"`
		Msg    string `json:"respdesc"`
	}
	if err := d.unicomHTTPGet(c, d.unicomFlowExchangeURL+"?"+urlVal, nil, &res); err != nil {
		return "", errors.Wrapf(ecode.Error(ecode.ServerErr, "联通接口维护中,请稍后再试"), "%v", err)
	}
	if res.Code != "0000" {
		if res.Code == "0001" {
			return res.Notice, errors.Wrap(ecode.Error(ecode.String(res.Code), res.Notice), d.unicomFlowExchangeURL+"?"+urlVal)
		}
		return res.Msg, errors.Wrap(ecode.Error(ecode.String(res.Code), res.Msg), d.unicomFlowExchangeURL+"?"+urlVal)
	}
	return res.Msg, nil
}

// broadbandHTTPGet http get
func (d *Dao) broadbandHTTPGet(c context.Context, urlStr string, params url.Values, res interface{}) (err error) {
	return d.wallHTTP(c, http.MethodGet, urlStr, params, res)
}

// unicomHTTPGet http get
func (d *Dao) unicomHTTPGet(c context.Context, urlStr string, params url.Values, res interface{}) (err error) {
	return d.wallHTTP(c, http.MethodGet, urlStr, params, res)
}

// wallHTTP http
func (d *Dao) wallHTTP(c context.Context, method, urlStr string, params url.Values, res interface{}) (err error) {
	var (
		req *http.Request
	)
	ru := urlStr
	if params != nil {
		ru = urlStr + "?" + params.Encode()
	}
	switch method {
	case http.MethodGet:
		req, err = http.NewRequest(http.MethodGet, ru, nil)
	default:
		req, err = http.NewRequest(http.MethodPost, urlStr, strings.NewReader(params.Encode()))
	}
	if err != nil {
		log.Error("unicom_http.NewRequest url(%s) error(%v)", urlStr+"?"+params.Encode(), err)
		return
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("X-BACKEND-BILI-REAL-IP", "")
	return d.client.Do(c, req, &res)
}

func (d *Dao) desEncrypt(src, key []byte) (string, error) {
	block, err := des.NewCipher(key)
	if err != nil {
		return "", err
	}
	bs := block.BlockSize()
	src = d.pkcs5Padding(src, bs)
	if len(src)%bs != 0 {
		return "", errors.New("Need a multiple of the blocksize")
	}
	out := make([]byte, len(src))
	dst := out
	for len(src) > 0 {
		block.Encrypt(dst, src[:bs])
		src = src[bs:]
		dst = dst[bs:]
	}
	encodeString := base64.StdEncoding.EncodeToString(out)
	return encodeString, nil
}

func (d *Dao) pkcs5Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

func (d *Dao) urlParams(v url.Values) string {
	if v == nil {
		return ""
	}
	var buf bytes.Buffer
	keys := make([]string, 0, len(v))
	for k := range v {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		vs := v[k]
		prefix := k + "="
		for _, v := range vs {
			if buf.Len() > 0 {
				buf.WriteByte('&')
			}
			buf.WriteString(prefix)
			buf.WriteString(v)
		}
	}
	return buf.String()
}

func (d *Dao) sign(params string) string {
	str := strings.Replace(params, "&", "$", -1)
	str2 := strings.Replace(str, "=", "$", -1)
	mh := md5.Sum([]byte(d.c.Unicom.UnicomSecurity + "$" + str2 + "$" + d.c.Unicom.UnicomSecurity))
	signparam := url.Values{}
	signparam.Set("sign", base64.StdEncoding.EncodeToString(mh[:]))
	return signparam.Encode()
}

// 通过联通接口验证用户是否为哔哩哔哩卡套餐
func (d *Dao) VerifyBiliBiliCardByUnicom(ctx context.Context, phoneNumber string) (*unicom.VerifyBiliBiliCardResponse, error) {
	user := d.c.Unicom.Verify.User
	pwd := d.c.Unicom.Verify.Password
	tick := strconv.FormatInt(time.Now().Unix(), 10)
	params := url.Values{}
	params.Add("mobile", phoneNumber)
	params.Add("user", user)
	params.Add("tick", tick)
	params.Add("key", verifySign(user, pwd, tick))
	resp := &unicom.VerifyBiliBiliCardResponse{}
	err := d.uclient.Get(ctx, d.unicomVerifyURL, "", params, &resp)
	if err != nil {
		log.Error("[dao.VerifyBiliBiliCardByUnicom]请求联通接口错误, error:%+v", err)
		return nil, err
	}
	return resp, nil
}

func verifySign(user, pwd, tick string) string {
	text := user + pwd + tick
	h := md5.New()
	_, _ = h.Write([]byte(text))
	return hex.EncodeToString(h.Sum(nil))[:16]
}

// 生成免流订购关系
func (d *Dao) UnicomFlowTryout(ctx context.Context, fakeID string) error {
	channel := d.c.Unicom.FlowTryout.Channel
	pwd := d.c.Unicom.FlowTryout.Password
	userid, _ := d.aesEncrypt([]byte(fakeID), []byte(pwd))
	tick := strconv.FormatInt(time.Now().Unix(), 10)
	params := struct {
		UserID    string `json:"userid"`
		Channel   string `json:"channel"`
		Timestamp string `json:"timestamp"`
		Signature string `json:"signature"`
	}{
		UserID:    userid,
		Channel:   channel,
		Timestamp: tick,
		Signature: flowTryoutVerifySign(userid, channel, tick, pwd),
	}
	var (
		data []byte
		req  *http.Request
		err  error
	)
	if data, err = json.Marshal(params); err != nil {
		return err
	}
	if req, err = http.NewRequest(http.MethodPost, d.unicomFlowTryout, bytes.NewBuffer(data)); err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	var resp struct {
		Code string `json:"resultCode"`
		Msg  string `json:"resultInfo"`
	}
	if err = d.uclient.Do(ctx, req, &resp); err != nil {
		log.Error("[dao.UnicomFlowTryout] 请求联通接口错误 url:%s, param:%+v, error:%+v ", d.unicomFlowTryout, params, err)
		return err
	}
	if resp.Code != "0" {
		if resp.Code == "8012" {
			// 已存在免流试看数据
			log.Warn("d.UnicomFlowTryout 已存在免流试看数据 fakeID:%s", fakeID)
			return nil
		}
		code, _ := strconv.Atoi(resp.Code)
		return errors.WithMessage(ecode.Int(code), resp.Msg)
	}
	return nil
}

func (d *Dao) aesEncrypt(src, key []byte) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	bs := block.BlockSize()
	src = d.pkcs5Padding(src, bs)
	if len(src)%bs != 0 {
		return "", errors.New("Need a multiple of the blocksize")
	}
	dst := make([]byte, len(src))
	CBCMode := cipher.NewCBCEncrypter(block, dst[:bs])
	CBCMode.CryptBlocks(dst, src)
	encodeString := strings.ToUpper(hex.EncodeToString(dst))
	return encodeString, nil
}

func flowTryoutVerifySign(userID, channel, tick, pwd string) string {
	text := userID + channel + tick + pwd
	h := md5.New()
	_, _ = h.Write([]byte(text))
	return hex.EncodeToString(h.Sum(nil))
}
