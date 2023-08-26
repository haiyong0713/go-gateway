package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"sync/atomic"

	"go-common/library/conf/env"
	"go-common/library/log"
	"go-gateway/app/web-svr/activity/admin/model/component"
)

type CorpWeChatRes struct {
	ErrCode int32  `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}

const (
	ContentTypeOfJson = "application/json"

	AlarmMsgTypeOfMarkdown = iota
	AlarmMsgTypeOfText

	AlarmMsgValueOfMarkdown = "markdown"
	AlarmMsgValueOfText     = "text"
)

var (
	cropWeChat atomic.Value
)

func init() {
	initOne := component.CorpWeChat{}
	cropWeChat.Store(initOne)
}

func UpdateCropWeChat(newOne component.CorpWeChat) {
	cropWeChat.Store(newOne)
}

func MentionUserIDs(robot component.CorpWeChat, alarmMsgType int) interface{} {
	switch alarmMsgType {
	case AlarmMsgTypeOfMarkdown:
		userIDs := ""
		for _, v := range robot.MentionUserIDs {
			if userIDs == "" {
				userIDs = fmt.Sprintf("@%v", v)
			} else {
				userIDs = fmt.Sprintf("%v @%v", userIDs, v)
			}
		}

		return userIDs
	case AlarmMsgTypeOfText:
		return robot.MentionUserIDs
	default:
		return ""
	}
}

func Robot() (robot component.CorpWeChat, err error) {
	if d, ok := cropWeChat.Load().(component.CorpWeChat); ok && d.WebhookUrl != "" && len(d.MentionUserIDs) != 0 {
		robot = d
		return
	}
	err = errors.New("robot is uninitialized")
	return
}

func SendCorpWeChatRobotAlarmByRobot(robot component.CorpWeChat, bs []byte) error {
	// escaping handling
	str := string(bs)
	str = strings.ReplaceAll(str, `\\\`, `\`)
	str = strings.ReplaceAll(str, `\\n`, `\n`)
	resp, err := http.Post(robot.WebhookUrl, ContentTypeOfJson, strings.NewReader(str))
	if err != nil {
		return err
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	corpWeChatRes := CorpWeChatRes{}
	err = json.Unmarshal(body, &corpWeChatRes)
	if err != nil {
		return err
	}

	if corpWeChatRes.ErrCode != 0 {
		return errors.New(
			fmt.Sprintf(
				"CorpWeChat response is not expected, req(%v), res(%v), http status code(%v)",
				string(bs),
				string(body),
				resp.StatusCode))
	}

	return nil
}

func SendCorpWeChatRobotAlarm(bs []byte) error {
	robot, err := Robot()
	if err != nil {
		return err
	}
	return SendCorpWeChatRobotAlarmByRobot(robot, bs)
}

func GenAlarmMsgDataByType(alarmMsgType int, content string) ([]byte, error) {
	bs := make([]byte, 0)

	robot, err := Robot()
	if err != nil {
		return bs, err
	}

	buffer := &bytes.Buffer{}
	encoder := json.NewEncoder(buffer)
	encoder.SetEscapeHTML(false)

	switch alarmMsgType {
	case AlarmMsgTypeOfMarkdown:
		m := make(map[string]interface{}, 2)
		{
			mdM := make(map[string]interface{}, 2)
			{
				mdM["mentioned_list"] = MentionUserIDs(robot, alarmMsgType)
				mdM["content"] = content
			}
			m["msgtype"] = AlarmMsgValueOfMarkdown
			m["markdown"] = mdM
		}

		err := encoder.Encode(m)

		return buffer.Bytes(), err
	case AlarmMsgTypeOfText:
		m := make(map[string]interface{}, 2)
		{
			mdM := make(map[string]interface{}, 2)
			{
				mdM["mentioned_list"] = MentionUserIDs(robot, alarmMsgType)
				mdM["content"] = content

				if len(robot.MentionUserTels) > 0 {
					mdM["mentioned_mobile_list"] = robot.MentionUserTels
				}
			}
			m["msgtype"] = AlarmMsgValueOfText
			m["text"] = mdM
		}

		err := encoder.Encode(m)

		return buffer.Bytes(), err
	default:
		return bs, errors.New("alarm msg type only support(markdown) now")
	}
}

func SendWeChatMessage(ctx context.Context, notification4FixMsg string) (err error) {
	if env.DeployEnv == env.DeployEnvUat {
		log.Errorc(ctx, "SendWeChatMessage  DeployEnvUat notification4FixMsg(%s)", notification4FixMsg)
		return
	}
	var bs []byte
	if bs, err = GenAlarmMsgDataByType(AlarmMsgTypeOfMarkdown, notification4FixMsg); err != nil {
		log.Errorc(ctx, "SendWeChatMessage  GenAlarmMsgDataByType error(%+v)", err)
		return
	}
	if err = SendCorpWeChatRobotAlarm(bs); err != nil {
		log.Errorc(ctx, "SendWeChatMessage SendCorpWeChatRobotAlarm error(%+v)", err)
	}
	return
}
