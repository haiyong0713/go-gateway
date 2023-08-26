package service

import (
	"context"
	"time"

	"go-gateway/app/app-svr/distribution/distribution/admin/internal/model"
	ac "go-gateway/app/app-svr/distribution/distribution/admin/internal/model/actionlog"
	"go-gateway/app/app-svr/distribution/distribution/admin/logcontext"

	"go-common/library/ecode"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/queue/databus/actionlog"

	"github.com/pkg/errors"
)

func (s *Service) LogAction(c *bm.Context, param *ac.Log) (res map[string]interface{}, err error) {
	res = map[string]interface{}{}
	searchRes, err := s.dao.LogAction(c, param)
	if err != nil {
		res["message"] = err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	res["data"] = searchRes.Item
	res["pager"] = searchRes.Pager
	return
}

// AddLog add action log
func addLog(ctx context.Context) (err error) {
	logContext, ok := logcontext.FromContext(ctx)
	if !ok {
		return errors.Errorf("no logContext found")
	}
	mInfo := &actionlog.UserInfo{
		Business: logContext.BusinessId(), // 业务 id, 请填写 info 中对应的业务 id
		Type:     logContext.ActionType(), //全匹配, 默认:操作对象的类型, 业务方可自定义
		Action:   model.ActionSave,        //全匹配, 默认:具体操作类型，如打回, 业务方可自定义
		Ctime:    time.Now(),
		Index:    []interface{}{logContext.UserName()},
		// 可以时间排序
		Content: map[string]interface{}{
			"json": logContext.ExtraContext(),
		}, // 数据只展示, 不参与搜索, 在 es 中保存为一个 json 字符串
	}

	// 同步请求
	return actionlog.User(mInfo)
}

func (s *Service) AsyncLog(ctx context.Context) {
	_ = s.fanout.Do(ctx, func(ctx context.Context) {
		if err := addLog(ctx); err != nil {
			log.Error("%+v", err)
		}
	})
}
