package service

import (
	"context"
	"fmt"
	pb "go-gateway/app/web-svr/space/interface/api/v1"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestServiceUpActivityTab(t *testing.T) {
	Convey("测试up主空间发起活动", t, WithService(func(s *Service) {
		Convey("正常操作", func() {
			Convey("当前无生效的tab,应该添加", func() {
				var (
					c   = context.Background()
					req = &pb.UpActivityTabReq{
						Mid:     27515233,
						State:   1,
						TabCont: 3115,
						TabName: "测试活动5",
					}
				)
				ret, err := s.UpActivityTab(c, req)
				So(err, ShouldBeNil)
				So(ret.Success, ShouldBeTrue)
				Convey("已新增Tab，应该获取最新tab", func() {
					tabReq := &pb.UserTabReq{
						Mid: 27515233,
					}
					tabResp, err2 := s.UserTab(c, tabReq)
					So(err2, ShouldBeNil)
					So(tabResp, ShouldNotBeNil)
					fmt.Printf(" {TabName:%s, Mid:%d, TabCont:%d} ", tabResp.TabName, tabResp.Mid, tabResp.TabCont)
				})
			})
			Convey("当前有生效的tab,应该更新", func() {
				var (
					c   = context.Background()
					req = &pb.UpActivityTabReq{
						Mid:     27515233,
						State:   1,
						TabCont: 3115,
						TabName: "测活动5",
					}
				)
				ret, err := s.UpActivityTab(c, req)
				So(err, ShouldBeNil)
				So(ret.Success, ShouldBeTrue)
				Convey("已更新Tab，应该获取最新tab", func() {
					tabReq := &pb.UserTabReq{
						Mid: 27515233,
					}
					tabResp, err2 := s.UserTab(c, tabReq)
					So(err2, ShouldBeNil)
					So(tabResp, ShouldNotBeNil)
					fmt.Printf(" {TabName:%s, Mid:%d, TabCont:%d} ", tabResp.TabName, tabResp.Mid, tabResp.TabCont)
				})
			})
			Convey("当前有生效的tab,更新非UP主Mid,应该失败", func() {
				var (
					c   = context.Background()
					req = &pb.UpActivityTabReq{
						Mid:     27515245,
						State:   1,
						TabCont: 3115,
						TabName: "测活动5",
					}
				)
				ret, err := s.UpActivityTab(c, req)
				So(err, ShouldNotBeNil)
				So(ret.Success, ShouldBeFalse)
			})
			Convey("当前有生效的tab,更新非UP主发起native id,应该失败", func() {
				var (
					c   = context.Background()
					req = &pb.UpActivityTabReq{
						Mid:     27515233,
						State:   1,
						TabCont: 4865,
						TabName: "测活动5",
					}
				)
				ret, err := s.UpActivityTab(c, req)
				So(err, ShouldNotBeNil)
				So(ret.Success, ShouldBeFalse)
			})
		})
		Convey("下线操作", func() {
			var (
				c   = context.Background()
				req = &pb.UpActivityTabReq{
					Mid:     27515233,
					State:   0,
					TabCont: 3115,
					TabName: "测试活动5",
				}
			)
			_, err := s.UpActivityTab(c, req)
			Convey("当前已有上线配置，应该下线成功", func() {
				So(err, ShouldBeNil)
				Convey("已下线tab，应该获取失败", func() {
					tabReq := &pb.UserTabReq{
						Mid: 27515233,
					}
					tabResp, err2 := s.UserTab(c, tabReq)
					So(err2, ShouldNotBeNil)
					So(tabResp, ShouldBeNil)
				})
			})
			Convey("当前没有有上线配置，应该下线失败", func() {
				So(err, ShouldNotBeNil)
			})
		})
		Convey("无效MID", func() {
			var (
				c   = context.Background()
				req = &pb.UpActivityTabReq{
					Mid:     27525240,
					State:   1,
					TabCont: 3115,
					TabName: "测试活动4",
				}
			)
			_, err := s.UpActivityTab(c, req)
			Convey("应该出现错误", func() {
				So(err, ShouldNotBeNil)
			})
		})
		Convey("无效NativeID", func() {
			var (
				c   = context.Background()
				req = &pb.UpActivityTabReq{
					Mid:     27515245,
					State:   1,
					TabCont: 12345,
					TabName: "测试活动4",
				}
			)
			_, err := s.UpActivityTab(c, req)
			Convey("应该报错", func() {
				So(err, ShouldNotBeNil)
			})
		})
		Convey("非up主发起NativeID", func() {
			var (
				c   = context.Background()
				req = &pb.UpActivityTabReq{
					Mid:     27515245,
					State:   1,
					TabCont: 5149,
					TabName: "测试活动4",
				}
			)
			_, err := s.UpActivityTab(c, req)
			Convey("应该报错", func() {
				So(err, ShouldNotBeNil)
			})
		})
	}))
}
