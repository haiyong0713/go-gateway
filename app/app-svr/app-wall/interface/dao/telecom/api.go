package telecom

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-gateway/app/app-svr/app-wall/interface/model/telecom"

	"github.com/pkg/errors"
	"github.com/xxtea/xxtea-go/xxtea"
)

const (
	_telecomFormat     = "json"
	_telecomClientType = "3"
	_payInfo           = "/api/v1/0/getPayInfo.do"
	_cancelRepeatOrder = "/api/v1/0/cancelRepeatOrder.do"
	_sucOrderList      = "/api/v1/0/getSucOrderList.do"
	_phoneArea         = "/api/v1/0/queryOperatorAndProvince.do"
	_orderState        = "/api/v1/0/queryOrderStatus.do"
	_acviteState       = "/serviceAgent/rest/polaris/biliDataPlanVerify"
	// card
	_phoneAuth = "/ic/auth"
)

var _activeResultMsg = map[string]string{
	"400": "无效的请求",
	"401": "Unauthorized - 用户没有登录，没有认证",
	"403": "Forbidden - 用户没有被授权使用这个服务",
	"404": "Not found - 资源没有被找到",
	"405": "Method Not Allowed - 请求不被支撑",
	"406": "Not Acceptable - 资源允许被创建的Content类型和请求中要求的可接受的类型不一致",
	"408": "Requst Timeout，服务端在请求资源的时候超时",
	"409": "Conflict - 因为冲突，请求没有被完成",
	"410": "Gone - 请求不再存在，比如一个请求已经被删除",
	"412": "Precondition Failed - 前置条件失败。在带有条件的请求的时候返回，例如当使用IFMatch时条件失败。用于乐观锁.",
	"415": "Unsupported Media Type - 不支持的媒体类型。例如客户端发送的请求里面没有Content-Type",
	"423": "Locked - 悲观锁",
	"428": "Precondition Required - 服务端需要有判断条件",
	"429": "Too many requests - 过多的服务请求，参见速率限制",
	"500": "Internal Server Error - 意外的服务器执行错误的通用错误信息（客户端也许可以重试）",
	"501": "Not Implemented - 服务端无法完成请求（通常用于描述未来的可用性，例如新的功能",
	"503": "Service Unavailable - 服务暂时无法使用 (例如因为过载) —客户端也许可以重试",
}

// PayInfo
// nolint:gomnd
func (d *Dao) PayInfo(c context.Context, requestNo int64, phone, isRepeatOrder, payChannel, payAction int, orderID int64, ipStr string,
	beginTime, firstOrderEndtime time.Time) (data *telecom.Pay, err error, msg string) {
	var payChannelStr string
	switch payChannel {
	case 1:
		payChannelStr = "31"
		ipStr = ""
	case 2:
		payChannelStr = "29"
	}
	params := url.Values{}
	params.Set("requestNo", strconv.FormatInt(requestNo, 10))
	params.Set("flowPackageId", strconv.Itoa(d.c.Telecom.PackageID))
	params.Set("contractId", "100174")
	params.Set("activityId", "101043")
	params.Set("phoneId", strconv.Itoa(phone))
	params.Set("bindApps", "tv.danmaku.bilianime|tv.danmaku.bili")
	params.Set("bindAppNames", "哔哩哔哩|哔哩哔哩")
	params.Set("isRepeatOrder", strconv.Itoa(isRepeatOrder))
	params.Set("payChannel", payChannelStr)
	if ipStr != "" {
		params.Set("userIp", ipStr)
	}
	params.Set("payPageType", "1")
	if d.telecomReturnURL != "" {
		params.Set("returnUrl", d.telecomReturnURL)
	}
	if d.telecomCancelPayURL != "" {
		params.Set("cancelPayUrl", d.telecomCancelPayURL)
	}
	params.Set("payAction", strconv.Itoa(payAction))
	// if startTime := beginTime.Format("20060102"); startTime != "19700101" && !beginTime.IsZero() {
	// 	params.Set("beginTime", startTime)
	// }
	// if endTime := firstOrderEndtime.Format("20060102"); endTime != "19700101" && !firstOrderEndtime.IsZero() {
	// 	params.Set("firstOrderEndtime", endTime)
	// }
	if orderID > 0 {
		params.Set("orderId", strconv.FormatInt(orderID, 10))
	}
	var res struct {
		Code   int `json:"resCode"`
		Detail struct {
			OrderID int64 `json:"orderId"`
			PayInfo struct {
				PayURL string `json:"payUrl"`
			} `json:"payInfo"`
		} `json:"detail"`
		Msg string `json:"resMsg"`
	}
	if err = d.wallHTTPPost(c, d.payInfoURL, params, &res); err != nil {
		log.Error("telecom_payInfoURL url(%s) error(%v)", d.payInfoURL+"?"+params.Encode(), err)
		return
	}
	if res.Code != 10000 {
		err = ecode.Int(res.Code)
		log.Error("telecom_url(%s) res code(%d) or res.data(%v)", d.payInfoURL+"?"+params.Encode(), res.Code, res.Detail)
		msg = res.Msg
		return
	}
	data = &telecom.Pay{
		RequestNo: requestNo,
		OrderID:   res.Detail.OrderID,
		PayURL:    res.Detail.PayInfo.PayURL,
	}
	return
}

// wallHTTPPost
func (d *Dao) wallHTTPPost(c context.Context, urlStr string, params url.Values, res interface{}) (err error) {
	newParams := url.Values{}
	encryptData := xxtea.Encrypt([]byte(params.Encode()), []byte(d.c.Telecom.AppSecret))
	hexStr := hex.EncodeToString(encryptData)
	newParams.Set("paras", hexStr)
	mh := md5.Sum([]byte(d.c.Telecom.AppID + _telecomClientType + _telecomFormat + hexStr + d.c.Telecom.AppSecret))
	newParams.Set("sign", hex.EncodeToString(mh[:]))
	newParams.Set("appId", d.c.Telecom.AppID)
	newParams.Set("clientType", _telecomClientType)
	newParams.Set("format", _telecomFormat)
	req, err := http.NewRequest("POST", urlStr, strings.NewReader(newParams.Encode()))
	if err != nil {
		log.Error("telecom_http.NewRequest url(%s) error(%v)", urlStr+"?"+newParams.Encode(), err)
		return
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("X-BACKEND-BILI-REAL-IP", "")
	return d.client.Do(c, req, &res)
}

// CancelRepeatOrder
// nolint:gomnd
func (d *Dao) CancelRepeatOrder(c context.Context, phone int, signNo string) (msg string, err error) {
	params := url.Values{}
	params.Set("phoneId", strconv.Itoa(phone))
	params.Set("signNo", signNo)
	var res struct {
		Code int    `json:"resCode"`
		Msg  string `json:"resMsg"`
	}
	if err = d.wallHTTPPost(c, d.cancelRepeatOrderURL, params, &res); err != nil {
		log.Error("telecom_payInfoURL url(%s) error(%v)", d.cancelRepeatOrderURL+"?"+params.Encode(), err)
		return
	}
	if res.Code != 10000 {
		err = ecode.Int(res.Code)
		log.Error("telecom_url(%s) res code(%d)", d.cancelRepeatOrderURL+"?"+params.Encode(), res.Code)
		msg = res.Msg
		return
	}
	msg = res.Msg
	return
}

// SucOrderList user order list
// nolint:gomnd
func (d *Dao) SucOrderList(c context.Context, phone int) (res *telecom.SucOrder, err error, msg string) {
	params := url.Values{}
	params.Set("phoneId", strconv.Itoa(phone))
	var resData struct {
		Code   int `json:"resCode"`
		Detail struct {
			AccessToken string              `json:"accessToken"`
			Orders      []*telecom.SucOrder `json:"orders"`
		} `json:"detail"`
		Msg string `json:"resMsg"`
	}
	if err = d.wallHTTPPost(c, d.sucOrderListURL, params, &resData); err != nil {
		log.Error("telecom_sucOrderListURL url(%s) error(%v)", d.sucOrderListURL+"?"+params.Encode(), err)
		return
	}
	if resData.Code != 10000 {
		err = ecode.Int(resData.Code)
		log.Error("telecom_url(%s) res code(%d)", d.sucOrderListURL+"?"+params.Encode(), resData.Code)
		msg = resData.Msg
		return
	}
	if len(resData.Detail.Orders) == 0 {
		err = ecode.NothingFound
		msg = "订单不存在"
		log.Error("telecom_order list phone(%v) len 0", phone)
		return
	}
	for _, r := range resData.Detail.Orders {
		if r.FlowPackageID == strconv.Itoa(d.c.Telecom.PackageID) {
			r.OrderID, _ = strconv.ParseInt(r.OrderIDStr, 10, 64)
			r.OrderIDStr = ""
			r.PortInt, _ = strconv.Atoi(r.Port)
			r.Port = ""
			res = r
			res.AccessToken = resData.Detail.AccessToken
			break
		}
	}
	if res == nil {
		log.Error("telecom_order bili phone(%v) is null", phone)
		msg = "订单不存在"
		err = ecode.NothingFound
		return
	}
	msg = resData.Msg
	return
}

// PhoneArea phone by area
// nolint:gomnd
func (d *Dao) PhoneArea(c context.Context, phone int) (area string, err error, msg string) {
	params := url.Values{}
	params.Set("phoneId", strconv.Itoa(phone))
	var resData struct {
		Code   int `json:"resCode"`
		Detail struct {
			RegionCode string `json:"regionCode"`
			AreaName   string `json:"areaName"`
		} `json:"detail"`
		Msg string `json:"resMsg"`
	}
	if err = d.wallHTTPPost(c, d.phoneAreaURL, params, &resData); err != nil {
		log.Error("telecom_phoneAreaURL url(%s) error(%v)", d.phoneAreaURL+"?"+params.Encode(), err)
		return
	}
	if resData.Code != 10000 {
		err = ecode.Int(resData.Code)
		log.Error("telecom_url(%s) res code(%d)", d.phoneAreaURL+"?"+params.Encode(), resData.Code)
		msg = resData.Msg
		return
	}
	area = resData.Detail.RegionCode
	return
}

// OrderState
// nolint:gomnd
func (d *Dao) OrderState(c context.Context, orderid int64) (res *telecom.OrderPhoneState, err error) {
	params := url.Values{}
	params.Set("orderId", strconv.FormatInt(orderid, 10))
	var resData struct {
		Code   int                      `json:"resCode"`
		Detail *telecom.OrderPhoneState `json:"detail"`
		Msg    string                   `json:"resMsg"`
	}
	if err = d.wallHTTPPost(c, d.orderStateURL, params, &resData); err != nil {
		log.Error("telecom_orderStateURL url(%s) error(%v)", d.orderStateURL+"?"+params.Encode(), err)
		return
	}
	if resData.Code != 10000 && resData.Code != 10013 {
		err = ecode.Int(resData.Code)
		log.Error("telecom_url(%s) res code(%d)", d.orderStateURL+"?"+params.Encode(), resData.Code)
		return
	}
	if resData.Code == 10013 {
		res = &telecom.OrderPhoneState{
			OrderState: 6,
		}
		return
	}
	res = resData.Detail
	if res.FlowPackageID != d.c.Telecom.PackageID {
		res.OrderState = 5
		return
	}
	switch resData.Detail.OrderState {
	case 5:
		res.OrderState = 6
	}
	return
}

func (d *Dao) telecomHTTPPost(c context.Context, requestNo int64, urlStr, telecomURL string, bytesData []byte, res interface{}) (err error) {
	var (
		req *http.Request
	)
	req, err = http.NewRequest(http.MethodPost, urlStr, bytes.NewReader(bytesData))
	if err != nil {
		log.Error("telecomHTTPPost NewRequest url(%s) error(%v)", urlStr+"?"+string(bytesData), err)
		return
	}
	req.Header.Set("Content-Type", "application/json;charset=UTF-8")
	mh := md5.Sum([]byte(d.c.Telecom.CardSpid + strconv.FormatInt(requestNo, 10) + d.c.Telecom.CardPass + telecomURL + string(bytesData)))
	xauth := d.c.Telecom.CardSpid + ";" + strconv.FormatInt(requestNo, 10) + ";" + hex.EncodeToString(mh[:])
	req.Header.Set("x-auth", xauth)
	req.Header.Set("User-Agent", "bilibili")
	return d.client.Do(c, req, &res)
}

func (d *Dao) PhoneAuth(c context.Context, requestNo int64, phone int) (*telecom.CardAuth, error) {
	param := map[string]interface{}{
		"phoneNumber": strconv.Itoa(phone),
	}
	var data struct {
		Code    int               `json:"code"`
		Message string            `json:"err"`
		Biz     *telecom.CardAuth `json:"biz"`
	}
	bs, err := json.Marshal(param)
	if err != nil {
		return nil, err
	}
	if err = d.telecomHTTPPost(c, requestNo, d.phoneAuthURL, _phoneAuth, bs, &data); err != nil {
		return nil, errors.Wrapf(ecode.Error(ecode.ServerErr, "电信接口维护中,请稍后再试"), "requestNo:%v, err:%v", requestNo, err)
	}
	if data.Code != 0 {
		return nil, errors.Wrapf(ecode.Error(ecode.Int(data.Code), "电信接口失败,请稍后再试"), "requestNo:%v, url:%v", requestNo, d.phoneAuthURL+"?"+string(bs))
	}
	return data.Biz, nil
}

func (d *Dao) ActiveState(ctx context.Context, accNbr string) (bool, error) {
	var res struct {
		Code   string `json:"code"`
		Status string `json:"status"`
	}
	data := map[string]interface{}{
		"accNbr": accNbr,
	}
	bs, err := json.Marshal(data)
	if err != nil {
		return false, err
	}
	req, err := http.NewRequest(http.MethodPost, d.activeStateURL, bytes.NewReader(bs))
	if err != nil {
		return false, err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("X-APP-ID", "ccc67c2ecf0e14f108172ad20f4e7634")
	req.Header.Add("X-APP-KEY", "78404b6a6288027581ca4356acb7a3a9")
	if err := d.client.Do(ctx, req, &res); err != nil {
		return false, errors.Wrapf(ecode.Error(ecode.ServerErr, "电信接口维护中,请稍后再试"), "%v", err)
	}
	if res.Code != "200" {
		return false, errors.Wrapf(ecode.Error(ecode.String(res.Code), "电信接口失败,请稍后再试"), "接口请求:%s,响应:%+v,错误信息:%s", d.activeStateURL+"?"+string(bs), res, _activeResultMsg[res.Code])
	}
	if res.Status == "valid" {
		return true, nil
	}
	return false, nil
}
