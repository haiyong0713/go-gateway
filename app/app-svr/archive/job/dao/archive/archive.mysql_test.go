package archive

import (
	"context"
	"time"

	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func Test_Videos2(t *testing.T) {
	Convey("Videos2", t, func() {
		vs, err := d.RawVideos(context.TODO(), 10098500)
		So(err, ShouldBeNil)
		for _, v := range vs {
			Printf("%+v", v)
		}
	})
}

func Test_TypeMapping(t *testing.T) {
	Convey("TypeMapping", t, func() {
		_, err := d.RawTypes(context.TODO())
		So(err, ShouldBeNil)
	})
}

func Test_Archive(t *testing.T) {
	Convey("Archive", t, func() {
		archive, err := d.RawArchive(context.TODO(), 1)

		t := archive.MTime.Time()
		Println(int64(time.Now().Sub(t).Seconds()))

		So(err, ShouldBeNil)
		Println(archive)
	})
}

func Test_Addit(t *testing.T) {
	Convey("Addit", t, func() {
		addit, err := d.RawAddit(context.TODO(), 1)
		So(err, ShouldBeNil)
		Println(addit)
	})
}
