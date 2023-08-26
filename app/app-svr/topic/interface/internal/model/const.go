package model

import (
	"fmt"
	"time"

	"go-common/library/log"

	topicapi "git.bilibili.co/bapis/bapis-go/topic/service"
)

func MakeCardTimeDesc(card *topicapi.TrafficCard) string {
	if card.StartTime >= card.EndTime {
		log.Warn("MakeCardTimeDesc parse failed: startTime >= endTime card=%+v", card)
		return ""
	}
	nowTimestamp := time.Now().Unix()
	switch {
	case nowTimestamp < card.StartTime:
		timeObj := time.Unix(card.StartTime, 0)
		return fmt.Sprintf("%d-%02d-%02d %02d:%02d开始", timeObj.Year(), timeObj.Month(), timeObj.Day(), timeObj.Hour(), timeObj.Minute())
	case nowTimestamp >= card.EndTime:
		return "活动已结束"
	default:
		timeObj := time.Unix(card.EndTime, 0)
		return fmt.Sprintf("%d-%02d-%02d %02d:%02d截止", timeObj.Year(), timeObj.Month(), timeObj.Day(), timeObj.Hour(), timeObj.Minute())
	}
}
