package unicom

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"net/url"
	"strconv"
	"time"

	"go-common/library/ecode"
	"go-common/library/stat/metric"
	"go-gateway/app/app-svr/app-wall/interface/model/unicom"

	"github.com/pkg/errors"
)

// curl 'http://op.10010.com:8802/api?user=000007&tick=1568115350&key=d24325eb7576e9b7&service=0&function=1&ip=112.64.119.196&pip=10.51.123.40'
const (
	_activateURL = "/api"
	_usermobURL  = "/api"
)

var (
	metricTotal     = metric.NewBusinessMetricCount("unicom_activate", "result", "errcode")
	autoMetricTotal = metric.NewBusinessMetricCount("unicom_auto_activate", "result", "resp_type", "resp_code")
)

var (
	_activateResult = map[string]string{
		"0":   "查询成功",
		"1":   "查询失败",
		"10":  "用户账号不存在",
		"11":  "密码校验错误",
		"13":  "IP地址认证错误（合作方使用了未授权的IP地址）",
		"14":  "超过最大并发连接数限制",
		"21":  "功能编号（function）参数错误",
		"22":  "参数缺失或者参数长度错误",
		"23":  "时间戳参数（tick）无效",
		"403": "用户没有指定业务/功能的使用权限",
		"404": "错误的页面调用",
		"503": "服务暂时不可用（系统忙/系统队列满）",
		"504": "服务暂时不可用",
		"128": "其他未知错误",
	}
	_activateErrcode = map[string]string{
		"1": "参数校验错误",
		"2": "数据返回错误",
		"3": "用户套餐查询失败",
		"4": "用户身份识别失败(无法通过公网IP识别)",
		"5": "用户身份识别失败(无法通过私网IP识别)",
		"6": "用户身份识别失败(其他错误)",
	}
	_fakeIDResult = map[string]string{
		"0":   "请求处理成功",
		"1":   "请求处理失败",
		"10":  "用户账号不存在",
		"11":  "密码校验错误",
		"13":  "IP地址认证错误（合作方使用了未授权的IP地址）",
		"22":  "参数缺失或者参数长度错误",
		"23":  "时间戳参数（tick）无效",
		"24":  "取号失败",
		"25":  "获取用户 IMEI 失败",
		"403": "用户没有指定业务/功能的使用权限",
		"503": "服务暂时不可用（系统忙/系统队列满）",
		"504": "服务暂时不可用",
		"128": "其他未知错误",
	}
	_fakeIDResultMsg = map[string]map[string]string{
		"0": {
			"-1": "网关内部错误",
			"1":  "网关内部错误",
			"2":  "网关内部错误",
			"3":  "网关内部错误",
			"4":  "app_id 不存在",
			"5":  "app_id 不合法",
			"6":  "无法用公网 ip 地址找到对应的省份",
			"7":  "app_id 不属于该用户",
			"8":  "用户传入 code 和 appid、user 不匹配",
			"9":  "网关内部错误",
		},
		"1": {
			"0":  "查询成功",
			"1":  "无法用私网 ip 地址找到对应的号码",
			"2":  "无法用公网 ip 地址找到对应的省份",
			"3":  "无法用 key 找到对应的号码",
			"4":  "无效的 app_id",
			"5":  "key 的格式错误",
			"6":  "key 已经过期",
			"7":  "取号功能暂时不可用",
			"8":  "不支持此功能",
			"9":  "内部路由错误",
			"10": "此号码是黑名单号码",
		},
	}
)

func (d *Dao) Activate(c context.Context, pip, ip string) (*unicom.ActivateResponse, error) {
	var resp *unicom.ActivateResponse
	user := d.c.Unicom.Activate.User
	tick := strconv.FormatInt(time.Now().Unix(), 10)
	params := url.Values{}
	params.Set("user", user)
	params.Set("tick", tick)
	params.Set("key", activateSign(user, tick, d.c.Unicom.Activate.Password))
	params.Set("service", "0")
	params.Set("function", "1")
	params.Set("ip", ip)
	params.Set("pip", pip)
	if err := d.uclient.Get(c, d.activateURL, "", params, &resp); err != nil {
		return nil, errors.Wrapf(ecode.Error(ecode.ServerErr, "联通接口维护中,请稍后再试"), "%v", err)
	}
	metricTotal.Inc(resp.Result, resp.Errcode)
	if resp.Result != "0" {
		return nil, errors.Wrapf(ecode.Error(ecode.String(resp.Result), "联通接口失败,请稍后再试"), "接口请求:%s,响应:%+v,错误信息:%s", d.activateURL+"?"+params.Encode(), resp, _activateResult[resp.Result]+","+_activateErrcode[resp.Errcode])
	}
	return resp, nil
}

func activateSign(user, tick, pwd string) string {
	text := user + tick + pwd
	h := md5.New()
	_, _ = h.Write([]byte(text))
	return hex.EncodeToString(h.Sum(nil))[:16]
}

func (d *Dao) GetFakeIDInfo(c context.Context, pip, ip string) (*unicom.FakeIDInfoResponse, error) {
	const (
		service  = "0"
		function = "8"
	)
	tick := strconv.FormatInt(time.Now().Unix(), 10)
	uu := d.c.Unicom.UnicomUsermob
	user := uu.User
	password := uu.Pass
	appid := uu.AppID
	apiUrl := d.usermobURL
	params := url.Values{}
	params.Set("user", user)
	params.Set("tick", tick)
	params.Set("key", activateSign(user, tick, password))
	params.Set("service", service)
	params.Set("function", function)
	params.Set("ip", ip)
	params.Set("pip", pip)
	params.Set("appid", appid)
	resp := &unicom.FakeIDInfoResponse{}
	if err := d.uclient.Get(c, apiUrl, "", params, &resp); err != nil {
		return nil, errors.Wrapf(ecode.Error(ecode.ServerErr, "联通接口维护中,请稍后再试"), "%v", err)
	}
	autoMetricTotal.Inc(resp.Result, resp.RespType, resp.RespCode)
	if resp.Result != "0" {
		return nil, errors.Wrapf(ecode.Error(ecode.String(resp.Result), "联通接口失败,请稍后再试"), "接口请求:%s,响应:%+v,错误信息:%s", apiUrl+"?"+params.Encode(), resp, _fakeIDResult[resp.Result]+","+_fakeIDResultMsg[resp.RespType][resp.RespCode])
	}
	return resp, nil
}
