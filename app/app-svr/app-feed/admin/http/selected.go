package http

import (
	"context"
	"fmt"
	xtime "time"
	"unicode/utf8"

	"go-common/library/ecode"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/app-svr/app-feed/admin/model/common"
	"go-gateway/app/app-svr/app-feed/admin/model/selected"
	"go-gateway/app/app-svr/app-feed/admin/util"

	"github.com/siddontang/go-mysql/mysql"
)

func selSeries(c *bm.Context) {
	param := new(struct {
		Type string `form:"type" validate:"required"`
	})
	if err := c.Bind(param); err != nil {
		return
	}
	c.JSON(selSvc.SelSeries(c, param.Type))
}

func selSort(c *bm.Context) {
	param := new(selected.SelSortReq)
	if err := c.Bind(param); err != nil {
		return
	}
	uid, name := util.UserInfo(c)
	if name == "" {
		c.JSONMap(map[string]interface{}{"message": "请重新登录"}, ecode.Unauthorized)
		c.Abort()
		return
	}
	c.JSON(nil, selSvc.SelResSort(c, param, &selected.Operator{UID: uid, Uname: name}))
}

func selExport(c *bm.Context) {
	req := &selected.SelResReq{}
	if err := c.Bind(req); err != nil {
		return
	}
	buf, err := selSvc.SelExport(c, req)
	if err != nil {
		c.JSON(nil, err)
		return
	}
	c.Writer.Header().Set("Content-Type", "application/csv")
	c.Writer.Header().Set("Content-Disposition",
		fmt.Sprintf("attachment;filename=\"Weekly_selected_%s.csv\"", xtime.Now().Format(mysql.TimeFormat)))
	//nolint:errcheck
	c.Writer.Write(buf.Bytes())
}

func selList(c *bm.Context) {
	req := &selected.SelResReq{}
	if err := c.Bind(req); err != nil {
		return
	}
	c.JSON(selSvc.SelList(c, req))
}

func selAdd(c *bm.Context) {
	req := &selected.ReqSelAdd{}
	if err := c.Bind(req); err != nil {
		return
	}
	uid, name := util.UserInfo(c)
	if name == "" {
		c.JSONMap(map[string]interface{}{"message": "请重新登录"}, ecode.Unauthorized)
		c.Abort()
		return
	}
	c.JSON(nil, selSvc.SelResAdd(c, req, &selected.Operator{UID: uid, Uname: name}))
}

func selEdit(c *bm.Context) {
	req := &selected.ReqSelEdit{}
	if err := c.Bind(req); err != nil {
		return
	}
	uid, name := util.UserInfo(c)
	if name == "" {
		c.JSONMap(map[string]interface{}{"message": "请重新登录"}, ecode.Unauthorized)
		c.Abort()
		return
	}
	c.JSON(nil, selSvc.SelResEdit(c, req, &selected.Operator{UID: uid, Uname: name}))
}

func selDelete(c *bm.Context) {
	param := new(selected.OpReq)
	if err := c.Bind(param); err != nil {
		return
	}
	uid, name := util.UserInfo(c)
	if name == "" {
		c.JSONMap(map[string]interface{}{"message": "请重新登录"}, ecode.Unauthorized)
		c.Abort()
		return
	}
	c.JSON(nil, selSvc.SelResDel(c, param.ID, &selected.Operator{UID: uid, Uname: name}))
}

func selReject(c *bm.Context) {
	param := new(selected.OpReq)
	if err := c.Bind(param); err != nil {
		return
	}
	uid, name := util.UserInfo(c)
	if name == "" {
		c.JSONMap(map[string]interface{}{"message": "请重新登录"}, ecode.Unauthorized)
		c.Abort()
		return
	}
	c.JSON(nil, selSvc.SelResReject(c, param.ID, &selected.Operator{UID: uid, Uname: name}))
}

func selPreview(c *bm.Context) {
	req := &selected.PreviewReq{}
	if err := c.Bind(req); err != nil {
		return
	}
	c.JSON(selSvc.SelPreview(c, req))
}

func arcPreview(c *bm.Context) {
	param := new(struct {
		Aid string `form:"aid" validate:"required"`
		ID  int64
	})
	if err := c.Bind(param); err != nil {
		return
	}
	var idErr error
	if param.ID, idErr = common.GetAvID(param.Aid); idErr != nil {
		res := map[string]interface{}{}
		res["message"] = idErr.Error()
		c.JSONMap(res, ecode.RequestErr)
	}
	if arc, err := selSvc.ArcTitle(c, param.ID); err != nil {
		res := map[string]interface{}{}
		res["message"] = err.Error()
		c.JSONMap(res, ecode.RequestErr)
	} else {
		c.JSON(arc, nil)
	}
}

func selSerieAudit(c *bm.Context) {
	req := &selected.PreviewReq{}
	if err := c.Bind(req); err != nil {
		return
	}
	uid, name := util.UserInfo(c)
	if name == "" {
		c.JSONMap(map[string]interface{}{"message": "请重新登录"}, ecode.Unauthorized)
		c.Abort()
		return
	}
	c.JSON(nil, selSvc.SelAudit(c, req, &selected.Operator{UID: uid, Uname: name}))
}

func selSerieEdit(c *bm.Context) {
	req := &selected.SerieEditReq{}
	if err := c.Bind(req); err != nil {
		return
	}
	uid, name := util.UserInfo(c)
	if name == "" {
		c.JSONMap(map[string]interface{}{"message": "请重新登录"}, ecode.Unauthorized)
		c.Abort()
		return
	}
	c.JSON(nil, selSvc.SerieEdit(c, req, &selected.Operator{UID: uid, Uname: name}))
}

const _moduleIDWeeklySelected = "weekly-selected"

func selTouchUsers(c *bm.Context) {
	param := new(selected.SeriePush)
	if err := c.Bind(param); err != nil {
		return
	}
	uid, name := util.UserInfo(c)
	if name == "" {
		c.JSONMap(map[string]interface{}{"message": "请重新登录"}, ecode.Unauthorized)
		c.Abort()
		return
	}

	titleLen := utf8.RuneCountInString(param.PushTitle)
	subTitleLen := utf8.RuneCountInString(param.PushSubtitle)
	if titleLen == 0 || titleLen > 20 || subTitleLen == 0 || subTitleLen > 40 {
		c.JSON(nil, ecode.Error(-400, "Push推送信息不符合预期，请检查文案长度！"))
		c.Abort()
		return
	}

	var (
		serie *selected.Serie
		err   error
	)

	if serie, err = selSvc.SelValidBeforeTouchUsers(c, &selected.FindSerie{ID: param.ID}); err != nil {
		c.JSON(nil, err)
		c.Abort()
		return
	}

	if err = selSvc.SelEditPushInfo(c, param); err != nil {
		c.JSON(nil, err)
		c.Abort()
		return
	}
	if serie.PushTitle == "" || param.PushTitle != "" {
		serie.PushTitle = param.PushTitle
	}
	if serie.PushSubtitle == "" || param.PushSubtitle != "" {
		serie.PushSubtitle = param.PushSubtitle
	}

	//nolint:gomnd
	redPointTaskStatus := serie.TaskStatus / 10
	//nolint:gomnd
	pushTaskStatus := serie.TaskStatus % 10

	if redPointTaskStatus == 0 || redPointTaskStatus == 2 {
		log.Warn("每周必看-更新红点-开始")
		// 每周必看不需要指定第三参数ID
		if err = popularSvc.RedDotUpdate(c, name, _moduleIDWeeklySelected, 0); err != nil {
			redPointTaskStatus = 2
			log.Warn("【日志报警】每周必看-更新红点-失败, 任务ID：%d, Error: %s", param.ID, err.Error())
		} else {
			redPointTaskStatus = 1
			log.Warn("每周必看-更新红点-成功, 期数：%d", serie.Number)
		}
	}

	if pushTaskStatus == 0 || pushTaskStatus == 2 {
		pushTaskStatus = 3
		// 异步执行push任务
		//nolint:biligowordcheck
		go (func() {
			ctx := context.Background()
			log.Warn("每周必看-PUSH用户-开始")
			newPushTaskStatus := 1
			if err := selSvc.PushSerie(ctx, serie); err != nil {
				newPushTaskStatus = 2
				log.Error("【日志报警】每周必看-PUSH用户-失败, 任务ID：%d, Error: %s", param.ID, err.Error())
			} else {
				log.Warn("每周必看-PUSH用户-成功, 期数：%d", serie.Number)
			}

			if err = selSvc.UpdatePushTaskStatus(ctx, param.ID, newPushTaskStatus, &selected.Operator{UID: uid, Uname: name}); err != nil {
				//nolint:govet
				log.Error("【日志报警】更新任务结果失败，当前任务ID: %d, 任务状态：%s， err:%s", param.ID, newPushTaskStatus, err.Error())
				return
			}
		})()
	}

	//nolint:gomnd
	taskStatus := redPointTaskStatus*10 + pushTaskStatus

	if err = selSvc.UpdateTaskStatus(c, param.ID, taskStatus, &selected.Operator{UID: uid, Uname: name}); err != nil {
		//nolint:govet
		log.Error("【日志报警】更新任务结果失败，当前任务ID: %d, 任务状态：%s， err:%s", param.ID, taskStatus, err.Error())
		c.JSON(nil, ecode.Error(-500, "更新任务结果失败"))
		c.Abort()
		return
	}

	c.JSON(nil, nil)
}

func selSeriesInUse(c *bm.Context) {
	if selSvc.SeriesInUse == nil {
		c.JSON(nil, ecode.Error(-777, "【每周必看-日志报警】无法获取每周必看内容"))
		c.Abort()
		return
	}
	var err error
	for _, v := range selSvc.SeriesInUse {
		if v == nil {
			err = ecode.Error(-776, "【每周必看-日志报警】无法获取完整的每周必看内容")
			break
		}
	}
	c.JSON(selSvc.SeriesInUse, err)

}

// 给创作中心用，返回最新一期每周必看核心信息
func latestSelPreview(c *bm.Context) {
	c.JSON(selSvc.LatestSelPreview(c))
}
