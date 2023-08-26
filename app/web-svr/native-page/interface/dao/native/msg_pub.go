package native

import (
	"context"
	"strconv"

	"go-common/library/log"
	v1 "go-gateway/app/web-svr/native-page/interface/api"
	lmdl "go-gateway/app/web-svr/native-page/interface/model/like"
)

var (
	_retryTime = 2
)

//nolint:bilirailguncheck
func (d *Dao) SendMsg(c context.Context, page *v1.NativePage, isOnLine bool) (err error) {
	var onLine int
	if isOnLine {
		onLine = 1
	}
	// 关闭消息通知
	if !d.openDynamic {
		return
	}
	msg := &lmdl.PageMsgPub{
		Category: page.TypeToString(),
		Value: &lmdl.DynamicMsg{
			PageID:       page.ID,
			TopicID:      page.ForeignID,
			TopicName:    page.Title,
			Online:       onLine,
			TopicLink:    page.SkipURL,
			Uid:          page.RelatedUid,
			ActType:      page.ActType,
			Hot:          page.Hot,
			DynamicID:    page.DynamicID,
			Attribute:    page.Attribute,
			Stime:        page.Stime,
			Etime:        page.Etime,
			PcURL:        page.PcURL,
			AnotherTitle: page.AnotherTitle,
			FromType:     page.FromType,
			State:        page.State,
		},
	}
	//重试机制
	for i := 0; i < _retryTime; i++ {
		if err = d.nativePub.Send(c, strconv.FormatInt(page.ID, 10), msg); err == nil {
			log.Info("SendMsg: d.nativePub.Send(%v)", msg.Value)
			return nil
		}
		log.Error("Fail to send DynamicMsg, msg=%+v error=%+v", msg.Value, err)
	}
	log.Error("日志告警:通知动态绑定关系失败-d.nativePub.Send(%v) error(%v)", msg.Value, err)
	return
}
