package dao

import (
	"context"
	"encoding/json"
	"fmt"
	"gopkg.in/h2non/gock.v1"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestDao_Upload(t *testing.T) {
	Convey("TestDao_UploadBoss", t, WithDao(func(d *Dao) {
		path := "testbb/test"
		payload := []byte("aaabbccc")
		res, err := d.UploadBoss(context.Background(), path, payload)
		So(err, ShouldBeNil)
		fmt.Println(err)
		fmt.Println(res)
	}))
}

func TestDao_CdnDoPreload(t *testing.T) {
	Convey("TestDao_CdnDoPreload", t, WithDao(func(d *Dao) {
		defer gock.OffAll()
		httpMock("POST", d.host.Cdn+_preHeat).Reply(200).JSON(`{"code": 0,"data":{}}`)
		urls := []string{"https://boss.hdslb.com/appstatic/Mod_414-fdd75498ef29582d316164a4918f3a81/ar.zip"}
		err := d.CdnDoPreload(context.Background(), urls)
		So(err, ShouldBeNil)
		fmt.Println(err)
	}))
}

func TestDao_CdnPreheatQuery(t *testing.T) {
	Convey("TestDao_CdnPreheatQuery", t, WithDao(func(d *Dao) {
		defer gock.OffAll()
		httpMock("POST", d.host.Cdn+_preHeatQuery).Reply(200).JSON(`{"code": 0,"data":{}}`)
		urls := []string{"https://boss.hdslb.com/appstatic/Mod_414-fdd75498ef29582d316164a4918f3a81/ar.zip"}
		res, err := d.CdnPreloadQuery(context.Background(), urls)
		So(err, ShouldBeNil)
		fmt.Println(err)
		bs, _ := json.Marshal(res)
		fmt.Println(string(bs))
	}))
}
