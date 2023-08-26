package ecode

import (
	"go-common/library/ecode"
)

var (
	TopicAlreadyExisted   = ecode.New(4402001) //发布话题已存在
	TopicPubTooFrequency  = ecode.New(4402301) //发布话题过频
	TopicPubUsrBanned     = ecode.New(4402302) //提交失败，账号已被封禁
	TopicPubExceedMaxCnt  = ecode.New(4402303) //提交失败，本周提交次数已达上限
	TopicUsrNoCreateRight = ecode.New(4402004) //提交失败，没有权限发布话题
	TopicSubFavFailed     = ecode.New(4402308) //订阅/取消订阅失败
)

func HandlePubEcodeToastErr(err error) error {
	if ecode.EqualError(TopicAlreadyExisted, err) {
		return ecode.Error(TopicAlreadyExisted, "发布话题已存在")
	}
	if ecode.EqualError(TopicPubTooFrequency, err) {
		return ecode.Error(TopicPubTooFrequency, "发布话题过频")
	}
	if ecode.EqualError(TopicPubUsrBanned, err) {
		return ecode.Error(TopicPubUsrBanned, "提交失败，账号已被封禁")
	}
	if ecode.EqualError(TopicPubExceedMaxCnt, err) {
		return ecode.Error(TopicPubExceedMaxCnt, "提交失败，本周提交次数已达上限")
	}
	if ecode.EqualError(TopicUsrNoCreateRight, err) {
		return ecode.Error(TopicUsrNoCreateRight, "提交失败，没有权限发布话题")
	}
	return err
}
