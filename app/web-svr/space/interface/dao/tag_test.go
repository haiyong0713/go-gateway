package dao

import (
	"context"
	"testing"

	"gopkg.in/h2non/gock.v1"

	"github.com/smartystreets/goconvey/convey"
)

func TestDao_TagSub(t *testing.T) {
	convey.Convey("test tag sub", t, func(ctx convey.C) {
		mid := int64(2089809)
		tid := int64(600)
		defer gock.OffAll()
		httpMock("POST", d.tagSubURL).Reply(200).JSON(`{"code": 0}`)
		err := d.TagSub(context.Background(), mid, tid)
		convey.So(err, convey.ShouldBeNil)
	})
}

func TestDao_TagCancelSub(t *testing.T) {
	convey.Convey("test cancel tag cancel sub", t, func(ctx convey.C) {
		mid := int64(2089809)
		tid := int64(600)
		defer gock.OffAll()
		httpMock("POST", d.tagCancelSubURL).Reply(200).JSON(`{"code": 0}`)
		err := d.TagCancelSub(context.Background(), mid, tid)
		convey.So(err, convey.ShouldBeNil)
	})
}

func TestDao_TagSubList(t *testing.T) {
	convey.Convey("test tag sub list", t, func(ctx convey.C) {
		defer gock.OffAll()
		httpMock("GET", d.tagSubListURL).Reply(200).JSON(`{"code":0,"data":[{"tag_id":4594,"tag_name":"DRRR","cover":"http://i1.hdslb.com/sp/37/37843e1b19020f6336b18c12c525bf3e_s.jpg","head_cover":"","content":"《无头骑士异闻录 DuRaRaRa!!》是由日本小说家成田良悟所写的轻小说作品，日文版由电击文库出版，中文版则由台湾角川出版、港澳地区由香港角川洲立代理销售。小说由ヤスダスズヒト（安田典生）作插画。以池袋为舞台，一个以向往非日常的少年、爱找","short_content":"","type":0,"state":0,"ctime":1433151310,"count":{"view":0,"use":0,"atten":0},"is_atten":1,"likes":0,"hates":0,"attribute":0,"liked":0,"hated":0}],"message":"0","total":400}`)
		vmid := int64(88889018)
		pn := 1
		ps := 15
		data, count, err := d.TagSubList(context.Background(), vmid, pn, ps)
		convey.So(err, convey.ShouldBeNil)
		convey.Printf("%+v,%d", data, count)
	})
}
