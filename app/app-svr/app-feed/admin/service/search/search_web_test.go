package search

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"go-gateway/app/app-svr/app-feed/admin/model/show"

	"github.com/smartystreets/goconvey/convey"
)

func TestSearchShieldList(t *testing.T) {
	convey.Convey("SearchShieldList", t, func(ctx convey.C) {
		var (
			c = context.Background()
		)
		p := &show.SearchShieldLP{
			Pn:    1,
			Ps:    10,
			Query: "aaa",
		}
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			res, err := s.SearchShieldList(c, p)
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(res, convey.ShouldNotBeNil)
			bs, _ := json.Marshal(res)
			fmt.Println(string(bs))
		})
	})
}

func TestOpenSearchShieldList(t *testing.T) {
	convey.Convey("OpenSearchShieldList", t, func(ctx convey.C) {
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			res, err := s.OpenSearchShieldList(context.Background())
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(res, convey.ShouldNotBeNil)
			bs, _ := json.Marshal(res)
			fmt.Println(string(bs))
		})
	})
}

func TestValidateShield(t *testing.T) {
	var (
		c = context.Background()
	)
	p := &show.SearchShieldValid{
		CardValue: "10",
		CardType:  1,
		Query:     `[{"value":"test"}]`,
	}
	convey.Convey("ValidateShield", t, func(ctx convey.C) {
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			err := s.ValidateShield(c, p)
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func TestAddSearchShield(t *testing.T) {
	var (
		c = context.Background()
	)
	convey.Convey("ValidateShield", t, func(ctx convey.C) {
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			p := &show.SearchShieldAP{
				CardType:  1,
				CardValue: "1210",
				Person:    "guolin",
				Reason:    "test",
				Query:     `[{"value":"test"}]`,
			}
			err := s.AddSearchShield(c, p, 500)
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func TestUpdateSearchShield(t *testing.T) {
	var (
		c = context.Background()
	)
	convey.Convey("ValidateShield", t, func(ctx convey.C) {
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			p := &show.SearchShieldUP{
				ID:        1,
				CardType:  1,
				CardValue: "1",
				Person:    "guolin",
				Reason:    "test1",
				Query:     `[{"value":"111"}]`,
			}
			err := s.UpdateSearchShield(c, p, 500)
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func TestOptionSearchShield(t *testing.T) {
	convey.Convey("ValidateShield", t, func(ctx convey.C) {
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			param := &show.SearchShieldOption{
				ID:    1,
				Check: 2,
				Name:  "test",
				UID:   500,
			}
			err := s.OptionSearchShield(param)
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}
