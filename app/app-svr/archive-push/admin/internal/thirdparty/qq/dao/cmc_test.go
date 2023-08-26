package dao

import (
	"fmt"
	"github.com/glycerine/goconvey/convey"
	"net/http"
	"strconv"
	"testing"
	"time"

	qqModel "go-gateway/app/app-svr/archive-push/admin/internal/thirdparty/qq/model"
)

func Test_PushPGCAdmin(t *testing.T) {
	convey.Convey("Push", t, func() {
		var (
			err error
			res *qqModel.PushPGCAdminReply
		)

		testReq := &qqModel.PushPGCAdminReq{
			SCreater:       "888",
			SCreaterHeader: "https://i2.hdslb.com/bfs/face/adcf8e24fd25f2001c5b77634252e6b770d45522.jpg@140w_140h_1c.jpg",
			STitle:         "测试用稿件标题3",
			SIMG:           "https://i2.hdslb.com/bfs/archive/71a3f2bbfc5d94cadd931f16c70d8d5b5daf07f5.jpg@206w_116h_1c_100q.webp",
			SDESC:          "测试用稿件描述3",
			SAuthor:        "风尘小飞侠",
			IType:          "0",
			ISubType:       "0",
			SOriginID:      "",
			SURL:           "",
			SCreated:       time.Now().Format("2006-01-02 15:04:05"),
			SCreatedOther:  time.Now().Format("2006-01-02 15:04:05"),
			STagsOther:     "生活,娱乐",
			SSource:        qqModel.DefaultSSource,
			SVID:           "BV1XT4y137T4",
			ITime:          strconv.FormatInt(300, 10),
			IFrom:          "11",
			SVideoSize:     fmt.Sprintf("%d*%d", 1920, 1080),
		}

		res, err = testD.PushPGCAdmin(testReq)
		if err != nil {
			fmt.Printf("error: %v", err)
		} else {
			fmt.Printf("response: %v", res)
		}
	})
}

func Test_DetailAdmin(t *testing.T) {
	convey.Convey("Detail", t, func() {
		var (
			err   error
			res   *qqModel.DetailAdminReply
			docid = "15152030122585241170"
		)
		if testD.httpClient == nil {
			testD.httpClient = http.DefaultClient
		}

		testQuery := &qqModel.DetailAdminQuery{
			BasePGCAdminQuery: &qqModel.BasePGCAdminQuery{CType: "2"},
			ID:                docid,
		}

		res, err = testD.DetailAdmin(testQuery)
		if err != nil {
			fmt.Printf("error: %v", err)
		} else {
			fmt.Printf("response: %v", res)
		}
	})
}

func Test_ModifyPGCAdmin_0(t *testing.T) {
	convey.Convey("Modify", t, func() {
		var (
			err   error
			res   *qqModel.ModifyPGCAdminReply
			docid = "15152030122585241170"
			mode  = qqModel.ModifyModeModify
		)

		testReq := &qqModel.ModifyPGCAdminReq{
			STitle: "测试用稿件标题4xxx",
		}

		res, err = testD.ModifyPGCAdmin(docid, mode, testReq)
		if err != nil {
			fmt.Printf("error: %v", err)
		} else {
			fmt.Printf("response: %v", res)
		}
	})
}

func Test_ModifyPGCAdmin_1(t *testing.T) {
	convey.Convey("Withdraw", t, func() {
		var (
			err   error
			res   *qqModel.ModifyPGCAdminReply
			docid = "15152030122585241170"
			mode  = qqModel.ModifyModeWithdraw
		)

		testReq := &qqModel.ModifyPGCAdminReq{
			STitle: "测试用稿件标题4xxx",
		}

		res, err = testD.ModifyPGCAdmin(docid, mode, testReq)
		if err != nil {
			fmt.Printf("error: %v", err)
		} else {
			fmt.Printf("response: %v", res)
		}
	})
}

func Test_UserContentListAdmin(t *testing.T) {
	convey.Convey("Detail", t, func() {
		var (
			err error
			res *qqModel.UserContentListAdminReply
		)

		testQuery := &qqModel.UserContentListAdminQuery{
			BasePGCAdminQuery: &qqModel.BasePGCAdminQuery{CType: "2"},
			Creater:           "888",
		}

		res, err = testD.UserContentListAdmin(testQuery)
		if err != nil {
			fmt.Printf("error: %v\n", err)
		} else {
			fmt.Printf("response: %v\n", res)
		}
	})
}
