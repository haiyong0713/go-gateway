package telecom

import (
	"context"
	"flag"
	"os"
	"strings"
	"testing"
	"time"

	"go-common/library/conf/paladin.v2"
	"go-gateway/app/app-svr/app-wall/interface/conf"
	"go-gateway/app/app-svr/app-wall/interface/model/telecom"

	. "github.com/smartystreets/goconvey/convey"
	gock "gopkg.in/h2non/gock.v1"
)

var (
	d *Dao
)

func ctx() context.Context {
	return context.Background()
}

func init() {
	if os.Getenv("DEPLOY_ENV") != "" {
		flag.Set("app_id", "main.app-svr.app-wall")
		flag.Set("conf_token", "yvxLjLpTFMlbBbc9yWqysKLMigRHaaiJ")
		flag.Set("tree_id", "2283")
		flag.Set("conf_version", "docker-1")
		flag.Set("deploy_env", "uat")
		flag.Set("conf_host", "config.bilibili.co")
		flag.Set("conf_path", "/tmp")
		flag.Set("region", "sh")
		flag.Set("zone", "sh001")
	}
	flag.Parse()
	if err := paladin.Init(); err != nil {
		panic(err)
	}
	defer paladin.Close()
	cfg := &conf.Config{}
	if err := paladin.Get("app-wall.toml").UnmarshalTOML(&cfg); err != nil {
		panic(err)
	}
	d = New(cfg)
	time.Sleep(time.Second)
}

func httpMock(method, url string) *gock.Request {
	r := gock.New(url)
	r.Method = strings.ToUpper(method)
	return r
}

func TestOrdersUserFlow(t *testing.T) {
	Convey("OrdersUserFlow", t, func() {
		_, err := d.OrdersUserFlow(ctx(), 1)
		So(err, ShouldBeNil)
	})
}

func TestOrdersUserByOrderID(t *testing.T) {
	Convey("OrdersUserByOrderID", t, func() {
		_, err := d.OrdersUserByOrderID(ctx(), 1)
		So(err, ShouldBeNil)
	})
}

func TestInOrderSync(t *testing.T) {
	Convey("InOrderSync", t, func() {
		t := &telecom.TelecomJSON{}
		_, err := d.InOrderSync(ctx(), 1, 1, "1", t)
		So(err, ShouldBeNil)
	})
}

func TestInRechargeSync(t *testing.T) {
	Convey("InRechargeSync", t, func() {
		t := &telecom.RechargeJSON{}
		_, err := d.InRechargeSync(ctx(), t)
		So(err, ShouldBeNil)
	})
}

func TestInCardOrderSync(t *testing.T) {
	Convey("InCardOrderSync", t, func() {
		t := &telecom.CardOrderJson{
			Biz: &telecom.CardOrderBizJson{
				Phone: "1",
			},
		}
		err := d.InCardOrderSync(ctx(), t.Biz)
		So(err, ShouldNotBeNil)
	})
}

func TestOrderUserByPhone(t *testing.T) {
	Convey("OrderUserByPhone", t, func() {
		_, err := d.OrderUserByPhone(ctx(), 17373189548)
		So(err, ShouldBeNil)
	})
}

func TestPayInfo(t *testing.T) {
	Convey("PayInfo", t, func() {
		d.client.SetTransport(gock.DefaultTransport)
		httpMock("POST", d.payInfoURL).Reply(200).JSON(`{
			"resCode": 0,
			"detail": {
				"orderId": 1111,
				"payInfo": {
					"payUrl": "www.t11.com"
				}
			},
			"resMsg": ""
		}`)
		res, _, _ := d.PayInfo(ctx(), 111, 111, 11, 1, 1, 1, "1", time.Now(), time.Now())
		So(res, ShouldNotBeEmpty)
	})
}

func TestCancelRepeatOrder(t *testing.T) {
	Convey("CancelRepeatOrder", t, func() {
		d.client.SetTransport(gock.DefaultTransport)
		httpMock("POST", d.cancelRepeatOrderURL).Reply(200).JSON(`{
			"resCode": 0,
			"resMsg": ""
		}`)
		res, _ := d.CancelRepeatOrder(ctx(), 111, "1111")
		So(res, ShouldBeEmpty)
	})
}

func TestSucOrderList(t *testing.T) {
	Convey("SucOrderList", t, func() {
		d.client.SetTransport(gock.DefaultTransport)
		httpMock("POST", d.sucOrderListURL).Reply(200).JSON(`{
			"resCode": 0,
			"detail": {
				"accessToken": "xxxxxxx",
				"orders": [{
					"flowPackageId": "xxxxx",
					"orderid": 111111
				},{
					"flowPackageId": "xxxxx",
					"orderid": 111111
				},{
					"flowPackageId": "xxxxx",
					"orderid": 111111
				},{
					"flowPackageId": "xxxxx",
					"orderid": 111111
				}]
			},
			"resMsg": ""
		}`)
		res, _, _ := d.SucOrderList(ctx(), 111)
		So(res, ShouldBeNil)
	})
}

func TestPhoneArea(t *testing.T) {
	Convey("PhoneArea", t, func() {
		d.client.SetTransport(gock.DefaultTransport)
		httpMock("POST", d.phoneAreaURL).Reply(200).JSON(`{
			"resCode": 0,
			"detail": {
				"regionCode": "000",
				"areaName": "xxx"
			},
			"resMsg": ""
		}`)
		res, _, _ := d.PhoneArea(ctx(), 111)
		So(res, ShouldBeEmpty)
	})
}

func TestOrderState(t *testing.T) {
	Convey("OrderState", t, func() {
		d.client.SetTransport(gock.DefaultTransport)
		httpMock("POST", d.orderStateURL).Reply(200).JSON(`{
			"resCode": 0,
			"detail": {
				"flowPackageId": 00000,
				"phoneId": "111111"
			},
			"resMsg": ""
		}`)
		res, _ := d.OrderState(ctx(), 111)
		So(res, ShouldNotBeEmpty)
	})
}

func TestAddPhoneCode(t *testing.T) {
	Convey("AddPhoneCode", t, func() {
		err := d.AddPhoneCode(ctx(), 111, "11")
		So(err, ShouldNotBeEmpty)
	})
}

func TestPhoneCode(t *testing.T) {
	Convey("PhoneCode", t, func() {
		res, _ := d.PhoneCode(ctx(), 111)
		So(res, ShouldBeEmpty)
	})
}

func TestAddPayPhone(t *testing.T) {
	Convey("AddPayPhone", t, func() {
		err := d.AddPayPhone(ctx(), 111, "")
		So(err, ShouldNotBeEmpty)
	})
}

func TestPayPhone(t *testing.T) {
	Convey("PayPhone", t, func() {
		res, _ := d.PayPhone(ctx(), 111)
		So(res, ShouldBeEmpty)
	})
}

func TestAddTelecomCache(t *testing.T) {
	Convey("AddTelecomCache", t, func() {
		err := d.AddTelecomCache(ctx(), 111, &telecom.OrderInfo{PhoneID: 111})
		So(err, ShouldBeNil)
	})
}

func TestTelecomCache(t *testing.T) {
	Convey("TelecomCache", t, func() {
		res, _ := d.TelecomCache(ctx(), 111)
		So(res, ShouldNotBeEmpty)
	})
}

func TestAddTelecomOrderIDCache(t *testing.T) {
	Convey("AddTelecomOrderIDCache", t, func() {
		var (
			orderID = int64(111)
			u       = &telecom.OrderInfo{
				PhoneID: 111,
			}
		)
		err := d.AddTelecomOrderIDCache(ctx(), orderID, u)
		So(err, ShouldBeNil)
	})
}

func TestTelecomOrderIDCache(t *testing.T) {
	Convey("TelecomOrderIDCache", t, func() {
		res, _ := d.TelecomOrderIDCache(ctx(), 111)
		So(res, ShouldNotBeEmpty)
	})
}

func TestAddTelecomCardCache(t *testing.T) {
	Convey("AddTelecomCardCache", t, func() {
		var (
			phone = 111
			u     = []*telecom.CardOrder{
				{
					Phone: 1,
				},
			}
		)
		err := d.AddTelecomCardCache(ctx(), phone, u)
		So(err, ShouldBeNil)
	})
}

func TestTelecomCardCache(t *testing.T) {
	Convey("TelecomCardCache", t, func() {
		res, _ := d.TelecomCardCache(ctx(), 111)
		So(res, ShouldNotBeEmpty)
	})
}

func TestDeleteTelecomCardCache(t *testing.T) {
	Convey("DeleteTelecomCardCache", t, func() {
		err := d.DeleteTelecomCardCache(ctx(), 111)
		So(err, ShouldBeNil)
	})
}

func TestSendSMS(t *testing.T) {
	Convey("SendSMS", t, func() {
		d.client.SetTransport(gock.DefaultTransport)
		httpMock("POST", d.smsSendURL).Reply(200).JSON(`{
			"code": 0,
		}`)
		err := d.SendSMS(ctx(), 111, "1", "1")
		So(err, ShouldNotBeNil)
	})
}

func TestSendTelecomSMS(t *testing.T) {
	Convey("SendTelecomSMS", t, func() {
		d.client.SetTransport(gock.DefaultTransport)
		httpMock("POST", d.smsSendURL).Reply(200).JSON(`{
			"code": 0,
		}`)
		err := d.SendTelecomSMS(ctx(), 11, "1")
		So(err, ShouldNotBeNil)
	})
}
