package feedback

import (
	"context"
	"fmt"
	"strings"

	appmdl "go-gateway/app/app-svr/fawkes/service/model/app"
	log "go-gateway/app/app-svr/fawkes/service/tools/logger"
)

const (
	// UpdateEvent feedback 更新事件
	UpdateEvent = "inner.feedback.update"
)

func (s *Service) feedbackUpdateAction(ctx context.Context, origin *appmdl.FeedbackDB) (err error) {
	defer func(c context.Context, err2 error) {
		if err != nil {
			log.Errorc(c, "EventError-%s %v", UpdateEvent, err2)
		}
	}(ctx, err)
	var current appmdl.FeedbackDB
	if current, err = s.fkDao.FeedbackQueryByPk(ctx, origin.ID); err != nil {
		log.Errorc(ctx, "%v", err)
		return
	}
	if origin.Status != appmdl.Processed && current.Status == appmdl.Processed {
		// 现在是已处理 发送微信提醒
		link := fmt.Sprintf("%s/#/usertrace/feedback?app_key=%s&mid=%d&buvid=&version_code=&status=&operator=%s&principal=&description=&robot_key=&create_start_time=&create_end_time=&pn=1", s.c.Host.Fawkes, current.AppKey, current.Mid, current.Operator)
		content := fmt.Sprintf("[MID]:%d\n[内容]:%s\n[原因]:%s\n已经解决，点击查看详情。", current.Mid, current.Description, current.CrashReason)
		if err = s.fkDao.WechatCardMessageNotify(
			"客诉反馈问题已处理",
			content,
			link,
			"",
			strings.Join([]string{current.Operator}, "|"),
			s.c.Comet.FawkesAppID); err != nil {
			log.Errorc(ctx, "%v", err)
			return
		}
	}
	return
}
