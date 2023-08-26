package http

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"go-common/library/ecode"
	log "go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/http/blademaster/binding"
	"go-common/library/net/http/blademaster/render"
	"go-common/library/net/metadata"
	"go-gateway/app/app-svr/app-wall/interface/model"
	"go-gateway/app/app-svr/app-wall/interface/model/unicom"
)

// ordersSync
// nolint:gomnd
func ordersSync(c *bm.Context) {
	res := map[string]interface{}{}
	unicom, err := requestJSONToMap(c.Request)
	if err != nil {
		res["result"] = "1"
		res["errorcode"] = "100"
		log.Error("unicom_ioutil.ReadAll error (%v)", err)
		returnDataJSON(c, res, ecode.RequestErr)
		return
	}
	usermob, err := url.QueryUnescape(unicom.Usermob)
	if err != nil {
		log.Error("unicom_url.QueryUnescape (%v) error (%v)", unicom.Usermob, err)
		res["result"] = "1"
		res["errorcode"] = "100"
		returnDataJSON(c, res, ecode.RequestErr)
		return
	}
	var (
		_aesKey = []byte("9ed226d9")
	)
	bs, err := base64.StdEncoding.DecodeString(usermob)
	if err != nil {
		log.Error("base64.StdEncoding.DecodeString(%s) error(%v)", usermob, err)
		c.JSON(nil, ecode.RequestErr)
		return
	}
	bs, err = unicomSvc.DesDecrypt(bs, _aesKey)
	if err != nil {
		log.Error("unicomSvc.DesDecrypt error(%v)", err)
		c.JSON(nil, ecode.RequestErr)
		return
	}
	var usermobStr string
	if len(bs) > 32 {
		usermobStr = string(bs[:32])
	} else {
		usermobStr = string(bs)
	}
	log.Info("unicomSvc.OrdersSync_usermob (%v) unicom (%v)", usermobStr, unicom)
	if err := unicomSvc.InOrdersSync(c, usermobStr, metadata.String(c, metadata.RemoteIP), unicom, time.Now()); err != nil {
		log.Error("unicomSvc.OrdersSync usermob (%v) unicom (%v) error (%v)", usermobStr, unicom, err)
		res["result"] = "1"
		res["errorcode"] = "100"
		returnDataJSON(c, res, err)
		return
	}
	res["result"] = "0"
	res["message"] = ""
	res["errorcode"] = ""
	returnDataJSON(c, res, nil)
}

// advanceSync
// nolint:gomnd
func advanceSync(c *bm.Context) {
	res := map[string]interface{}{}
	unicom, err := requestJSONToMap(c.Request)
	if err != nil {
		res["result"] = "1"
		res["errorcode"] = "100"
		log.Error("unicom_ioutil.ReadAll error (%v)", err)
		returnDataJSON(c, res, ecode.RequestErr)
		return
	}
	usermob, err := url.QueryUnescape(unicom.Usermob)
	if err != nil {
		log.Error("unicom_url.QueryUnescape (%v) error (%v)", unicom.Usermob, err)
		res["result"] = "1"
		res["errorcode"] = "100"
		returnDataJSON(c, res, ecode.RequestErr)
		return
	}
	var (
		_aesKey = []byte("9ed226d9")
	)
	bs, err := base64.StdEncoding.DecodeString(usermob)
	if err != nil {
		log.Error("base64.StdEncoding.DecodeString(%s) error(%v)", usermob, err)
		c.JSON(nil, ecode.RequestErr)
		return
	}
	bs, err = unicomSvc.DesDecrypt(bs, _aesKey)
	if err != nil {
		log.Error("unicomSvc.DesDecrypt error(%v)", err)
		c.JSON(nil, ecode.RequestErr)
		return
	}
	var usermobStr string
	if len(bs) > 32 {
		usermobStr = string(bs[:32])
	} else {
		usermobStr = string(bs)
	}
	log.Info("unicomSvc.AdvanceSync_usermob (%v) unicom (%v)", usermobStr, unicom)
	if err := unicomSvc.InAdvanceSync(c, usermobStr, metadata.String(c, metadata.RemoteIP), unicom, time.Now()); err != nil {
		log.Error("unicomSvc.InAdvanceSync usermob (%v) unicom (%v) error (%v)", usermobStr, unicom, err)
		res["result"] = "1"
		res["errorcode"] = "100"
		returnDataJSON(c, res, err)
		return
	}
	res["result"] = "0"
	res["message"] = ""
	res["errorcode"] = ""
	returnDataJSON(c, res, nil)
}

// flowSync
// nolint:gomnd
func flowSync(c *bm.Context) {
	res := map[string]interface{}{}
	unicom, err := requestJSONToMap(c.Request)
	if err != nil {
		res["result"] = "1"
		res["errorcode"] = "100"
		log.Error("unicom_ioutil.ReadAll error (%v)", err)
		returnDataJSON(c, res, ecode.RequestErr)
		return
	}
	var flowbyte float64
	if flowbyte, err = strconv.ParseFloat(unicom.FlowbyteStr, 64); err != nil {
		log.Error("unicom_flowbyte strconv.ParseFloat(%s) error(%v)", unicom.FlowbyteStr, err)
		res["result"] = "1"
		res["errorcode"] = "100"
		returnDataJSON(c, res, ecode.RequestErr)
		return
	}
	usermob, err := url.QueryUnescape(unicom.Usermob)
	if err != nil {
		log.Error("unicom_url.QueryUnescape (%v) error (%v)", unicom.Usermob, err)
		res["result"] = "1"
		res["errorcode"] = "100"
		returnDataJSON(c, res, ecode.RequestErr)
		return
	}
	var (
		_aesKey = []byte("9ed226d9")
	)
	bs, err := base64.StdEncoding.DecodeString(usermob)
	if err != nil {
		log.Error("unicom_base64.StdEncoding.DecodeString(%s) error(%v)", usermob, err)
		c.JSON(nil, ecode.RequestErr)
		return
	}
	bs, err = unicomSvc.DesDecrypt(bs, _aesKey)
	if err != nil {
		log.Error("unicomSvc.DesDecrypt error(%v)", err)
		c.JSON(nil, ecode.RequestErr)
		return
	}
	var usermobStr string
	if len(bs) > 32 {
		usermobStr = string(bs[:32])
	} else {
		usermobStr = string(bs)
	}
	if err := unicomSvc.FlowSync(c, int(flowbyte*1024), usermobStr, unicom.Time, metadata.String(c, metadata.RemoteIP), time.Now()); err != nil {
		log.Error("unicomSvc.FlowSync error (%v)", err)
		res["result"] = "1"
		res["errorcode"] = "100"
		returnDataJSON(c, res, err)
		return
	}
	res["result"] = "0"
	res["message"] = ""
	res["errorcode"] = ""
	returnDataJSON(c, res, nil)
}

// inIPSync
func inIPSync(c *bm.Context) {
	res := map[string]interface{}{}
	unicom, err := requestIPJSONToMap(c.Request)
	if err != nil {
		res["result"] = "1"
		res["errorcode"] = "100"
		log.Error("unicom_ioutil.ReadAll error (%v)", err)
		returnDataJSON(c, res, ecode.RequestErr)
		return
	}
	log.Info("unicomSvc.InIpSync_unicom (%v)", unicom)
	if err := unicomSvc.InIPSync(c, metadata.String(c, metadata.RemoteIP), unicom, time.Now()); err != nil {
		log.Error("unicomSvc.InIpSync unicom (%v) error (%v)", unicom, err)
		res["result"] = "1"
		res["errorcode"] = "100"
		returnDataJSON(c, res, err)
		return
	}
	res["result"] = "0"
	res["message"] = ""
	res["errorcode"] = ""
	returnDataJSON(c, res, nil)
}

// userFlow
func userFlow(c *bm.Context) {
	res := map[string]interface{}{}
	params := c.Request.Form
	usermob := params.Get("usermob")
	mobiApp := params.Get("mobi_app")
	buildStr := params.Get("build")
	build, _ := strconv.Atoi(buildStr)
	ipStr := metadata.String(c, metadata.RemoteIP)
	data, err := unicomSvc.UserFlow(c, usermob, mobiApp, ipStr, build, time.Now())
	if err != nil {
		c.JSON(nil, err)
		return
	}
	res["data"] = data
	returnDataJSON(c, res, nil)
}

// userFlowState
func userFlowState(c *bm.Context) {
	res := map[string]interface{}{}
	params := c.Request.Form
	usermob := params.Get("usermob")
	data := unicomSvc.UserFlowState(c, usermob, time.Now())
	res["data"] = data
	returnDataJSON(c, res, nil)
}

// userState
func userState(c *bm.Context) {
	res := map[string]interface{}{}
	params := c.Request.Form
	usermob := params.Get("usermob")
	mobiApp := params.Get("mobi_app")
	buildStr := params.Get("build")
	build, _ := strconv.Atoi(buildStr)
	ipStr := metadata.String(c, metadata.RemoteIP)
	data, err := unicomSvc.UserState(c, usermob, mobiApp, ipStr, build, time.Now())
	if err != nil {
		c.JSON(nil, err)
		return
	}
	res["data"] = data
	returnDataJSON(c, res, nil)
}

// unicomState
func unicomState(c *bm.Context) {
	res := map[string]interface{}{}
	params := c.Request.Form
	usermob := params.Get("usermob")
	mobiApp := params.Get("mobi_app")
	buildStr := params.Get("build")
	build, _ := strconv.Atoi(buildStr)
	ipStr := metadata.String(c, metadata.RemoteIP)
	data := unicomSvc.UnicomState(c, usermob, mobiApp, ipStr, build, time.Now())
	res["data"] = data
	returnDataJSON(c, res, nil)
}

// unicomStateM
// nolint:gomnd
func unicomStateM(c *bm.Context) {
	res := map[string]interface{}{}
	params := c.Request.Form
	usermob := params.Get("usermob")
	mobiApp := params.Get("mobi_app")
	buildStr := params.Get("build")
	build, _ := strconv.Atoi(buildStr)
	ipStr := metadata.String(c, metadata.RemoteIP)
	var (
		_aesKey = []byte("9ed226d9")
	)
	bs, err := base64.StdEncoding.DecodeString(usermob)
	if err != nil {
		log.Error("base64.StdEncoding.DecodeString(%s) error(%v)", usermob, err)
		c.JSON(nil, ecode.RequestErr)
		return
	}
	bs, err = unicomSvc.DesDecrypt(bs, _aesKey)
	if err != nil {
		log.Error("unicomSvc.DesDecrypt error(%v)", err)
		c.JSON(nil, ecode.RequestErr)
		return
	}
	var usermobStr string
	if len(bs) > 32 {
		usermobStr = string(bs[:32])
	} else {
		usermobStr = string(bs)
	}
	data := unicomSvc.UnicomState(c, usermobStr, mobiApp, ipStr, build, time.Now())
	res["data"] = data
	returnDataJSON(c, res, nil)
}

// RequestJsonToMap
func requestJSONToMap(request *http.Request) (unicom *unicom.UnicomJson, err error) {
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
	log.Info("unicom orders json body(%s)", body)
	if err = json.Unmarshal(body, &unicom); err != nil {
		log.Error("json.Unmarshal UnicomJson(%v) error (%v)", unicom, err)
		return
	}
	if err = unicom.UnicomJSONChange(); err != nil {
		log.Error("unicom.UnicomJSONChange unicom (%v) error (%v)", unicom, err)
	}
	return
}

// RequestIpJsonToMap
func requestIPJSONToMap(request *http.Request) (*unicom.UnicomIpJson, error) {
	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		log.Error("unicom_ioutil.ReadAll error (%v)", err)
		return nil, err
	}
	defer request.Body.Close()
	if len(body) == 0 {
		return nil, ecode.RequestErr
	}
	log.Info("unicom ip json body(%s)", body)
	var res *unicom.UnicomIpJson
	if err := json.Unmarshal(body, &res); err != nil {
		return nil, err
	}
	return res, nil
}

func pack(c *bm.Context) {
	params := c.Request.Form
	usermob := params.Get("usermob")
	if usermob == "" {
		c.JSON(nil, ecode.RequestErr)
	}
	midInter, _ := c.Get("mid")
	mid := midInter.(int64)
	err := unicomSvc.Pack(c, usermob, mid, time.Now())
	c.JSON(nil, err)
}

func isUnciomIP(c *bm.Context) {
	params := c.Request.Form
	mobiApp := params.Get("mobi_app")
	buildStr := params.Get("build")
	build, _ := strconv.Atoi(buildStr)
	ipStr := metadata.String(c, metadata.RemoteIP)
	ip := model.InetAtoN(ipStr)
	if err := unicomSvc.IsUnciomIP(ip, ipStr, mobiApp, build, time.Now()); err != nil {
		c.JSON(nil, err)
	}
	res := map[string]interface{}{
		"code": ecode.OK,
	}
	returnDataJSON(c, res, nil)
}

func userUnciomIP(c *bm.Context) {
	params := c.Request.Form
	usermob := params.Get("usermob")
	mobiApp := params.Get("mobi_app")
	buildStr := params.Get("build")
	build, _ := strconv.Atoi(buildStr)
	ipStr := metadata.String(c, metadata.RemoteIP)
	ip := model.InetAtoN(ipStr)
	res := map[string]interface{}{}
	res["data"] = unicomSvc.UserUnciomIP(ip, ipStr, usermob, mobiApp, build, time.Now())
	returnDataJSON(c, res, nil)
}

func orderPay(c *bm.Context) {
	res := map[string]interface{}{}
	params := c.Request.Form
	usermob := params.Get("usermob")
	channel := params.Get("channel")
	ordertypeStr := params.Get("ordertype")
	ordertype, _ := strconv.Atoi(ordertypeStr)
	data, msg, err := unicomSvc.Order(c, usermob, channel, ordertype, time.Now())
	if err != nil {
		if msg == "" {
			c.JSON(nil, err)
			return
		}
		res["message"] = msg
		returnDataJSON(c, res, err)
		return
	}
	res["data"] = data
	returnDataJSON(c, res, nil)
}

func orderCancel(c *bm.Context) {
	res := map[string]interface{}{}
	params := c.Request.Form
	usermob := params.Get("usermob")
	data, msg, err := unicomSvc.CancelOrder(c, usermob, time.Now())
	if err != nil {
		if msg == "" {
			c.JSON(nil, err)
			return
		}
		res["message"] = msg
		returnDataJSON(c, res, err)
		return
	}
	res["data"] = data
	returnDataJSON(c, res, nil)
}

func smsCode(c *bm.Context) {
	res := map[string]interface{}{}
	params := c.Request.Form
	phone := params.Get("phone")
	msg, err := unicomSvc.UnicomSMSCode(c, phone, time.Now())
	if err != nil {
		if msg == "" {
			c.JSON(nil, err)
			return
		}
		res["message"] = msg
		returnDataJSON(c, res, err)
		return
	}
	res["message"] = msg
	returnDataJSON(c, res, nil)
}

func bindUser(ctx *bm.Context) {
	params := ctx.Request.Form
	phoneStr := params.Get("phone")
	phone, err := strconv.Atoi(phoneStr)
	if err != nil {
		log.Error("%+v", err)
		ctx.JSON(nil, ecode.RequestErr)
		return
	}
	codeStr := params.Get("code")
	code, err := strconv.Atoi(codeStr)
	if err != nil {
		log.Error("%+v", err)
		ctx.JSON(nil, ecode.RequestErr)
		return
	}
	midInter, _ := ctx.Get("mid")
	mid := midInter.(int64)
	ctx.JSON(nil, unicomSvc.BindUser(ctx, phone, code, mid, time.Now()))
}

func unbindUser(ctx *bm.Context) {
	params := ctx.Request.Form
	phoneStr := params.Get("phone")
	phone, err := strconv.Atoi(phoneStr)
	if err != nil {
		log.Error("%+v", err)
		ctx.JSON(nil, err)
		return
	}
	midInter, _ := ctx.Get("mid")
	mid := midInter.(int64)
	ctx.JSON(nil, unicomSvc.UnbindUser(ctx, mid, phone))
}

func userBind(c *bm.Context) {
	res := map[string]interface{}{}
	var mid int64
	midInter, ok := c.Get("mid")
	if ok {
		mid = midInter.(int64)
	} else {
		res["message"] = "账号未登录"
		returnDataJSON(c, res, ecode.RequestErr)
		return
	}
	data, msg, err := unicomSvc.UserBind(c, mid)
	if err != nil {
		if msg == "" {
			c.JSON(nil, err)
			return
		}
		res["message"] = msg
		returnDataJSON(c, res, err)
		return
	}
	res["data"] = data
	res["message"] = ""
	returnDataJSON(c, res, nil)
}

func packList(c *bm.Context) {
	param := new(struct {
		Entry int `form:"entry"`
	})
	if err := c.Bind(param); err != nil {
		return
	}
	res := map[string]interface{}{
		"data": unicomSvc.UnicomPackList(param.Entry),
	}
	returnDataJSON(c, res, nil)
}

func packReceive(c *bm.Context) {
	params := c.Request.Form
	id, err := strconv.ParseInt(params.Get("id"), 10, 64)
	if err != nil {
		c.JSON(nil, err)
		return
	}
	midInter, _ := c.Get("mid")
	mid := midInter.(int64)
	msg, err := unicomSvc.PackReceive(c, mid, id, time.Now())
	res := map[string]interface{}{}
	if msg != "" {
		res["message"] = msg
	}
	c.JSONMap(res, err)
}

func flowPack(c *bm.Context) {
	params := c.Request.Form
	flowID := params.Get("flow_id")
	if flowID == "" {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	midInter, _ := c.Get("mid")
	mid := midInter.(int64)
	msg, err := unicomSvc.FlowPack(c, mid, flowID, time.Now())
	res := map[string]interface{}{}
	if msg != "" {
		res["message"] = msg
	}
	c.JSONMap(res, err)
}

func userPacksLog(c *bm.Context) {
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
	c.JSON(unicomSvc.UserPacksLog(c, startTime, time.Now(), pn, metadata.String(c, metadata.RemoteIP)))
}

func userBindLog(c *bm.Context) {
	res := map[string]interface{}{}
	var mid int64
	midInter, ok := c.Get("mid")
	if ok {
		mid = midInter.(int64)
	} else {
		res["message"] = "账号未登录"
		returnDataJSON(c, res, ecode.RequestErr)
		return
	}
	data, err := unicomSvc.UserBindLog(c, mid, time.Now())
	if err != nil {
		c.JSON(nil, err)
		return
	}
	res["data"] = data
	returnDataJSON(c, res, nil)
}

func welfareBindState(c *bm.Context) {
	params := c.Request.Form
	midStr := params.Get("mid")
	mid, err := strconv.ParseInt(midStr, 10, 64)
	if err != nil || mid < 1 {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	data := unicomSvc.WelfareBindState(c, mid)
	res := map[string]interface{}{
		"state": data,
	}
	c.JSON(res, nil)
}

func userBindInfoByPhone(c *bm.Context) {
	params := c.Request.Form
	phone := params.Get("phone")
	if phoneInt, err := strconv.ParseInt(phone, 10, 64); err != nil || phoneInt <= 0 {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	ipStr := metadata.String(c, metadata.RemoteIP)
	c.JSON(unicomSvc.UserBindInfoByPhone(c, phone, ipStr, time.Now()))
}

func addUserBindIntegral(c *bm.Context) {
	params := c.Request.PostForm
	integralStr := params.Get("integral")
	integral, err := strconv.Atoi(integralStr)
	if err != nil || integral <= 0 {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	midsStr := strings.Split(params.Get("mids"), ",")
	var mids []int64
	for _, midStr := range midsStr {
		mid, _ := strconv.ParseInt(midStr, 10, 64)
		if mid <= 0 {
			c.JSON(nil, ecode.RequestErr)
			return
		}
		mids = append(mids, mid)
	}
	ipStr := metadata.String(c, metadata.RemoteIP)
	c.JSON(unicomSvc.AddUserBindIntegral(c, mids, integral, ipStr))
}

func flowSign(c *bm.Context) {
	params := c.Request.Form
	integralStr := params.Get("timestamp")
	c.JSON(map[string]interface{}{
		"sign": unicomSvc.FlowSign(integralStr),
	}, nil)
}

func activate(c *bm.Context) {
	var mid int64
	midInter, ok := c.Get("mid")
	if ok {
		mid = midInter.(int64)
	}
	params := c.Request.Form
	pips := strings.Split(params.Get("pip"), ",")
	var pip string
	for _, p := range pips {
		if p != "" {
			pip = p
			break
		}
	}
	if !model.IsIPv4(pip) {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	ip := metadata.String(c, metadata.RemoteIP)
	data := unicomSvc.Activate(c, pip, ip, mid)
	c.JSON(data, nil)
}

func unicomActiveState(ctx *bm.Context) {
	param := &unicom.UserActiveParam{}
	if err := ctx.Bind(param); err != nil {
		return
	}
	midInter, ok := ctx.Get("mid")
	if ok {
		param.Mid = midInter.(int64)
	}
	header := ctx.Request.Header
	param.Buvid = header.Get("Buvid")
	param.IP = metadata.String(ctx, metadata.RemoteIP)
	if !param.Auto {
		if param.Usermob == "" {
			ctx.JSON(nil, ecode.Error(ecode.RequestErr, "手动激活伪码为空"))
			return
		}
		data, err := unicomSvc.ActiveState(ctx, param.Mid, param.Usermob, time.Now())
		unicomSvc.UserActiveLog(param, data, err, "手动激活")
		ctx.JSON(data, err)
		return
	}
	// 自动激活
	pips := strings.Split(param.Pip, ",")
	if len(pips) == 0 {
		ctx.JSON(nil, ecode.RequestErr)
		return
	}
	var pip string
	for _, val := range pips {
		if val != "" {
			pip = val
			break
		}
	}
	if !model.IsIPv4(pip) {
		ctx.JSON(nil, ecode.Error(ecode.RequestErr, fmt.Sprintf("私网IP:%s,不是IPV4", pip)))
		return
	}
	param.SinglePip = pip
	data, err := unicomSvc.AutoActiveStateByUsermob(ctx, param)
	unicomSvc.UserActiveLog(param, data, err, "自动激活")
	ctx.JSON(data, err)
}

func couponVerify(ctx *bm.Context) {
	param := &unicom.CouponParam{}
	if err := ctx.BindWith(param, binding.JSON); err != nil {
		return
	}
	data := make(map[string]interface{})
	if err := unicomSvc.CouponVerify(ctx, param); err != nil {
		data["result"] = false
		ctx.Render(http.StatusOK, render.MapJSON(data))
		return
	}
	data["result"] = true
	ctx.Render(http.StatusOK, render.MapJSON(data))
}

func unicomFlowTryout(ctx *bm.Context) {
	param := &unicom.UnicomFlowTryoutParam{}
	if err := ctx.Bind(param); err != nil {
		return
	}
	param.IP = metadata.String(ctx, metadata.RemoteIP)
	if !model.IsIPv4(param.Pip) {
		ctx.JSON(nil, ecode.Error(ecode.RequestErr, fmt.Sprintf("私网IP:%s,不是IPV4", param.Pip)))
		return
	}
	fakeID, err := unicomSvc.UnicomFlowTryout(ctx, param.FakeID, param.Pip, param.IP)
	data := map[string]string{
		"fake_id": fakeID,
	}
	ctx.JSON(data, err)
}
