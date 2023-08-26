package tool

import (
	"errors"
	"fmt"
	"testing"
	"time"
)

func setUp() {
	initOne := CorpWeChat{
		WebhookUrl:      "https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=8a541600-2c1e-48c3-90cc-535cf59a0fd6",
		MentionUserIDs:  []string{"leijiru", "wuliang02"},
		MentionUserTels: []string{"15555742725"},
	}
	cropWeChat.Store(initOne)
}

func TestCorpWeChatAlarm(t *testing.T) {
	setUp()
	t.Run("send internal error message", sendInternalErrorTip)
	t.Run("send archive score biz message", sendArchiveScoreTip)
}

func sendArchiveScoreTip(t *testing.T) {
	archiveScoreAlarmMsgTemplate := `赛事视频得分更新耗时：<font color=\"info\">%v</font>，请相关同事注意。\n
>预计更新赛事视频数:<font color=\"comment\">%v</font> \n
>成功更新赛事视频数:<font color=\"info\">%v</font> \n
>失败更新赛事视频数:<font color=\"warning\">%v</font> \n
>详情未匹配赛事视频数:<font color=\"warning\">%v</font> \n
>实时统计未匹配赛事视频数:<font color=\"warning\">%v</font> \n
>%v`
	robot, robotErr := Robot()
	if robotErr != nil {
		t.Errorf("Robot >>> unexpected err: %v", robotErr)

		return
	}

	d := time.Second * 1412

	content := fmt.Sprintf(
		archiveScoreAlarmMsgTemplate,
		d,
		8888,
		6666,
		2222,
		0,
		0,
		MentionUserIDs(robot, AlarmMsgTypeOfMarkdown))

	bs, err := GenAlarmMsgDataByType(AlarmMsgTypeOfMarkdown, content)
	if err != nil {
		t.Errorf("GenAlarmMsgDataByType >>> unexpected err: %v", err)

		return
	}

	if err := SendCorpWeChatRobotAlarm(bs); err != nil {
		t.Errorf("SendCorpWeChatRobotAlarm >>> unexpected err: %v", err)
	}
}

func sendInternalErrorTip(t *testing.T) {
	alarmErr := errors.New("Connection timed out")
	bs, err := GenAlarmMsgDataByType(
		AlarmMsgTypeOfText,
		fmt.Sprintf(`服务内部错误：%v, 请管理员及时查看`, alarmErr))
	if err != nil {
		t.Errorf("GenAlarmMsgDataByType >>> unexpected err: %v", err)

		return
	}

	if err := SendCorpWeChatRobotAlarm(bs); err != nil {
		t.Errorf("SendCorpWeChatRobotAlarm >>> unexpected err: %v", err)
	}
}
