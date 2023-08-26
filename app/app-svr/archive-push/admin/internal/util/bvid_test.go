package util

import (
	"fmt"
	"github.com/glycerine/goconvey/convey"
	"testing"
)

func Test_AvToBv(t *testing.T) {
	convey.Convey("AV to BV", t, func() {
		avid := int64(440076924)
		bvid, err := AvToBv(avid)
		fmt.Printf("AV(%d) to BV(%s)\n", avid, bvid)
		convey.ShouldBeNil(err)
		convey.ShouldNotBeNil(bvid)
	})
}

func Test_BvToAv(t *testing.T) {
	convey.Convey("BV to AV", t, func() {
		bvid := "BV1qu4y1n7kG"
		avid, err := BvToAv(bvid)
		fmt.Printf("BV(%s) to AV(%d)\n", bvid, avid)
		convey.ShouldBeNil(err)
		convey.ShouldBeGreaterThan(avid, 0)
	})
}

func Test_Attributes(t *testing.T) {
	convey.Convey("attributes", t, func() {
		attr := int32(16768)
		fmt.Printf("%d\n", attr>>24&int32(1))
		fmt.Printf("%d\n", attr>>29&int32(1))
	})
}
