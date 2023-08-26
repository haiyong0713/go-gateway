package favorite

import (
	"context"
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	gock "gopkg.in/h2non/gock.v1"
)

func httpMock(method, url string) *gock.Request {
	r := gock.New(url)
	r.Method = strings.ToUpper(method)
	return r
}

func TestIsFavDefault(t *testing.T) {
	Convey(t.Name(), t, func() {
		var (
			mid, aid int64
		)
		d.client.SetTransport(gock.DefaultTransport)
		httpMock("GET", d.isFavDef).Reply(200).JSON(`{
			"code": 0,
			"data": {
				"default": true
			},
			"message": "ok"
		}`)
		res, err := d.IsFavDefault(context.Background(), mid, aid)
		So(res, ShouldNotBeNil)
		So(err, ShouldBeNil)
	})
}

func TestIsFav(t *testing.T) {
	Convey(t.Name(), t, func() {
		var (
			mid, aid int64
		)
		d.client.SetTransport(gock.DefaultTransport)
		httpMock("GET", d.isFav).Reply(200).JSON(`{
			"code": 0,
			"data": {
				"favoured": true,
				"count": 1
			},
			"message": "ok"
		}`)
		res, err := d.IsFav(context.Background(), mid, aid)
		So(res, ShouldNotBeNil)
		So(err, ShouldBeNil)
	})
}

func TestAddFav(t *testing.T) {
	Convey(t.Name(), t, func() {
		var (
			mid, aid int64
		)
		d.client.SetTransport(gock.DefaultTransport)
		httpMock("GET", d.addFav).Reply(200).JSON(`{
			"code": 0,
			"message": "ok"
		  }`)
		err := d.AddFav(context.Background(), mid, aid)
		So(err, ShouldBeNil)
	})
}
