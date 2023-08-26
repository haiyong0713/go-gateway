package ci

import (
	"context"
	"testing"

	"go-common/library/log"

	appmdl "go-gateway/app/app-svr/fawkes/service/model/app"
	cimdl "go-gateway/app/app-svr/fawkes/service/model/ci"

	. "github.com/smartystreets/goconvey/convey"
)

// 测试邮件拼接
func TestCombineMailContent(t *testing.T) {
	Convey("", t, WithService(func(s *Service) {
		var (
			app *appmdl.APP
			ci  *cimdl.BuildPack
			c   = context.Background()
		)
		app, _ = s.fkDao.AppPass(c, "android_missevan")
		ci, _ = s.fkDao.BuildPackById(c, 273475)
		var (
			content string
			err     error
		)
		if content, err = s.combineWeChatContent(c, app, ci); err != nil {
			log.Error("combineWeChatContent err %v", err)
		}
		So(content, ShouldEqual, "")
	}))
}
