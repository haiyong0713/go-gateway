package archive

import (
	"context"
	"fmt"
	"testing"

	cv "github.com/smartystreets/goconvey/convey"
)

func TestArcFromTaishan(t *testing.T) {
	var (
		c           = context.TODO()
		aid         = int64(640002042)
		notExistAid = int64(640002041)
	)
	cv.Convey("TestArcFromTaishan", t, func(ctx cv.C) {
		a, err := d.getArcFromTaishan(c, aid)
		cv.So(err, cv.ShouldBeNil)
		cv.So(a, cv.ShouldNotBeNil)

		a2, err := d.getArcFromTaishan(c, notExistAid)
		cv.So(err, cv.ShouldNotBeNil)
		cv.So(a2, cv.ShouldBeNil)
	})
}

func TestBatchArcFromTaishan(t *testing.T) {
	var (
		c            = context.Background()
		aids         = []int64{280039050, 880112625, 520075756}
		notExistAids = []int64{640002041}
	)
	cv.Convey("TestBatchArcFromTaishan", t, func(ctx cv.C) {
		am, _, err := d.batchGetArcFromTaishan(c, aids)
		cv.So(am, cv.ShouldNotBeNil)
		cv.So(err, cv.ShouldBeNil)

		am2, _, err := d.batchGetArcFromTaishan(c, notExistAids)
		cv.So(am2, cv.ShouldBeNil)
		cv.So(err, cv.ShouldNotBeNil)
	})
}

func TestGetPagesFromTaishan(t *testing.T) {
	var (
		c           = context.Background()
		aid         = int64(520075756)
		notExistAid = int64(640002041)
	)
	cv.Convey("TestGetPagesFromTaishan", t, func(ctx cv.C) {
		p, err := d.getPagesFromTaishan(c, aid)
		cv.So(err, cv.ShouldBeNil)
		cv.So(p, cv.ShouldNotBeNil)
		cv.Println(p)

		p2, err := d.getPagesFromTaishan(c, notExistAid)
		cv.So(err, cv.ShouldNotBeNil)
		cv.Println(err)
		cv.So(p2, cv.ShouldBeNil)
	})
}

func TestBatchGetPagesFromTaishan(t *testing.T) {
	var (
		c    = context.Background()
		aids = []int64{960095866, 200101644, 560001857, 520075756}
		//notExistAids = []int64{640002041}
	)
	cv.Convey("TestBatchGetPagesFromTaishan", t, func(ctx cv.C) {
		ps, _, err := d.batchGetPagesFromTaishan(c, aids)
		cv.So(err, cv.ShouldBeNil)
		cv.So(ps, cv.ShouldNotBeNil)
		cv.Println(ps)

		//ps2, _, err := d.batchGetPagesFromTaishan(c, notExistAids)
		//cv.So(err, cv.ShouldNotBeNil)
		//cv.Println(err)
		//cv.So(ps2, cv.ShouldBeNil)
	})
}

func TestGetSimpleArcFromTaishan(t *testing.T) {
	var (
		c           = context.Background()
		aid         = int64(840079730)
		notExistAid = int64(640002041)
	)
	cv.Convey("TestGetSimpleArcFromTaishan", t, func(ctx cv.C) {
		sa, err := d.getSimpleArcFromTaishan(c, aid)
		cv.So(err, cv.ShouldBeNil)
		cv.So(sa, cv.ShouldNotBeNil)
		cv.Println(sa)

		sa2, err := d.getSimpleArcFromTaishan(c, notExistAid)
		cv.So(err, cv.ShouldNotBeNil)
		cv.Println(err)
		cv.So(sa2, cv.ShouldBeNil)
	})
}

func TestBatchGetSimpleArcFromTaishan(t *testing.T) {
	var (
		c            = context.Background()
		aids         = []int64{640002042, 640002041, 280080197, 840079730}
		notExistAids = []int64{640002041}
	)
	cv.Convey("TestBatchGetSimpleArcFromTaishan", t, func(ctx cv.C) {
		sam, err := d.batchGetSimpleArcFromTaishan(c, aids)
		cv.So(err, cv.ShouldBeNil)
		cv.So(sam, cv.ShouldNotBeNil)
		cv.Println(sam)

		sam2, err := d.batchGetSimpleArcFromTaishan(c, notExistAids)
		cv.So(err, cv.ShouldNotBeNil)
		cv.Println(err)
		cv.So(sam2, cv.ShouldBeNil)
	})
}

func TestGetVideoFromTaishan(t *testing.T) {
	var (
		c           = context.Background()
		aid         = int64(680049839)
		cid         = int64(10226342)
		notExistCid = int64(1)
	)
	cv.Convey("TestGetVideoFromTaishan", t, func(ctx cv.C) {
		p, err := d.getVideoFromTaishan(c, aid, cid)
		cv.So(err, cv.ShouldBeNil)
		cv.So(p, cv.ShouldNotBeNil)
		cv.Println(p)

		p2, err := d.getVideoFromTaishan(c, aid, notExistCid)
		cv.So(err, cv.ShouldNotBeNil)
		cv.Println(err)
		cv.So(p2, cv.ShouldBeNil)
	})
}

func TestBatchPutTaishan(t *testing.T) {
	var (
		c    = context.Background()
		keys = []string{"key1", "key2"}
	)
	kvMap := make(map[string][]byte, len(keys))
	for _, key := range keys {
		kvMap[key] = []byte("可莉-测试value")
	}
	cv.Convey("TestBatchPutTaishan", t, func(ctx cv.C) {
		err := d.batchPutTaishan(c, kvMap)
		cv.So(err, cv.ShouldBeNil)
	})
}

func TestBatchGetTaishan(t *testing.T) {
	var (
		c    = context.Background()
		keys = []string{"key1", "key2"}
	)

	cv.Convey("TestBatchGetTaishan", t, func(ctx cv.C) {
		res, err := d.batchGetFromTaishan(c, keys)
		cv.So(err, cv.ShouldBeNil)
		for _, re := range res {
			fmt.Println(string(re))
		}
	})
}
