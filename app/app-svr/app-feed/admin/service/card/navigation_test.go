package card

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	model "go-gateway/app/app-svr/app-feed/admin/model/card"

	"github.com/smartystreets/goconvey/convey"
)

func TestService_AddNavigationCard(t *testing.T) {
	convey.Convey("AddNavigationCard", t, func(ctx convey.C) {
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			corner := &model.NavCorner{
				Type:     1,
				Text:     "corner text",
				SunPic:   "bbb",
				NightPic: "aaa",
				Width:    100,
				Height:   50,
			}
			button := &model.NavButton{
				Type: 1,
				Text: "custom button text",
			}
			child3nd1 := &model.Navigation3rd{
				Title:     "3rd title 1",
				ReType:    1,
				ReValue:   "www.bilibili.com",
				Deletable: 0,
			}
			child3nd2 := &model.Navigation3rd{
				Title:     "3rd title 2",
				ReType:    1,
				ReValue:   "bilibili.com",
				Deletable: 0,
			}
			child2nd := &model.Navigation2nd{
				Title:     "",
				Deletable: 0,
				Children:  []*model.Navigation3rd{child3nd1, child3nd2},
			}
			navigation := &model.Navigation{Children: []*model.Navigation2nd{child2nd}}
			param := &model.AddNavigationCardReq{
				Uid:        1,
				Username:   "litongyu",
				Title:      "title",
				Desc:       "desc",
				Corner:     corner,
				Button:     button,
				Navigation: navigation,
			}
			res, err := s.AddNavigationCard(context.Background(), param)
			ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
			bs, _ := json.Marshal(res)
			fmt.Println(string(bs))
		})
	})
}
