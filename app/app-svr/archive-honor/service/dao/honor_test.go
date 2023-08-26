package dao

import (
	"context"
	"fmt"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestHonorsByAid(t *testing.T) {
	var (
		c   = context.TODO()
		aid = int64(1)
	)
	convey.Convey("honorsByAid", t, func(ctx convey.C) {
		ctx.Convey("Then err should be nil.", func(ctx convey.C) {
			res, err := d.honorsByAid(c, aid)
			fmt.Printf("%+v", res)
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func TestDelHonor(t *testing.T) {
	var (
		c   = context.TODO()
		aid = int64(1)
		typ = int32(1)
	)
	convey.Convey("delHonor", t, func(ctx convey.C) {
		ctx.Convey("Then err should be nil.", func(ctx convey.C) {
			_, err := d.delHonor(c, aid, typ)
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func TestHonorUpdate(t *testing.T) {
	var (
		c     = context.TODO()
		aid   = int64(2)
		typ   = int32(4)
		url   = "xxx"
		desc  = "desc"
		naUrl = ""
	)
	convey.Convey("HonorUpdate", t, func(ctx convey.C) {
		ctx.Convey("Then err should be nil.", func(ctx convey.C) {
			//honor := &api.Honor{Aid: 1, Type: 2, Url: "xxx"}
			//honors, err := proto.Marshal(honor)
			//fmt.Printf("proto.Marshal %v err %v", honors, err)
			//h := &api.Honor{}
			//if err := proto.Unmarshal(honors, h); err != nil {
			//	fmt.Printf("jsonpb.UnmarshalString(%s) error(%v)", string(honors), err)
			//	return
			//}
			//fmt.Printf("%v", h)
			delRows, err := d.HonorUpdate(c, aid, typ, url, desc, naUrl)
			fmt.Println(delRows)
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}
