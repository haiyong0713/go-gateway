package dao

import (
	"context"
	"encoding/json"
	"go-gateway/app/web-svr/appstatic/admin/model"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestDaoSaveRawRules(t *testing.T) {
	convey.Convey("SaveRawRules", t, func(convCtx convey.C) {
		var (
			c     = context.Background()
			rules = []*model.ChronosRule{
				{

					Avids: "all",
					BuildLimit: []*model.VerLimit{
						{
							Platform:  "android",
							Value:     800,
							Condition: "gt",
						},
						{
							Platform:  "ios",
							Condition: "all",
						},
					},
				},
				{
					Mids:  "1,23",
					Avids: "1,2,2,3",
				},
			}
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			err := d.SaveRawRules(c, rules)
			convCtx.Convey("Then err should be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestDaoRawRules(t *testing.T) {
	convey.Convey("RawRules", t, func(convCtx convey.C) {
		var (
			c = context.Background()
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			rules, err := d.RawRules(c)
			str, _ := json.Marshal(rules)
			convey.Println(string(str))
			convCtx.Convey("Then err should be nil.rules should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.So(rules, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestDaoSavePlayerRules(t *testing.T) {
	convey.Convey("SavePlayerRules", t, func(convCtx convey.C) {
		var (
			c     = context.Background()
			rules = []*model.ChronosRule{
				{
					Avids: "all",
					BuildLimit: []*model.VerLimit{
						{
							Platform:  "android",
							Value:     888,
							Condition: "gt",
						},
						{
							Platform:  "ios",
							Condition: "all",
						},
					},
				},
				{
					Mids:  "1,23",
					Avids: "1,2,2,3",
				},
			}
		)
		playerRules := make([]*model.PlayerRule, 0)
		for _, v := range rules {
			playerRules = append(playerRules, &model.PlayerRule{
				ChronosRule: *v,
				MD5:         "33344",
			})
		}
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			err := d.SavePlayerRules(c, playerRules)
			convCtx.Convey("Then err should be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}
