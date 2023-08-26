package tool

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"go-gateway/app/web-svr/activity/job/conf"
	"io/ioutil"
	"net/http"
	"strings"
	"sync/atomic"
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
	cropWeChat        atomic.Value
	cropWeChatForVote atomic.Value
)

func init() {
	initOne := conf.CorpWeChat{}
	cropWeChat.Store(initOne)
	initVote := conf.CorpWeChat{}
	cropWeChatForVote.Store(initVote)
}

func UpdateCropWeChat(conf *conf.Config) {
	cropWeChat.Store(conf.Notifier)
	cropWeChatForVote.Store(conf.NotifierForVote)
}

func GenAlarmMsgDataByTypeByRobot(robot conf.CorpWeChat, alarmMsgType int, content string, shouldMentionUser bool) ([]byte, error) {
	bs := make([]byte, 0)
	buffer := &bytes.Buffer{}
	encoder := json.NewEncoder(buffer)
	encoder.SetEscapeHTML(false)

	switch alarmMsgType {
	case AlarmMsgTypeOfMarkdown:
		m := make(map[string]interface{}, 2)
		{
			mdM := make(map[string]interface{}, 2)
			{
				if shouldMentionUser {
					mdM["mentioned_list"] = MentionUserIDs(robot, alarmMsgType)
				}
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
				if shouldMentionUser {
					mdM["mentioned_list"] = MentionUserIDs(robot, alarmMsgType)
				}

				mdM["content"] = content

				if len(robot.MentionUserTels) > 0 && shouldMentionUser {
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

func GenAlarmMsgDataByType(alarmMsgType int, content string, shouldMentionUser bool) ([]byte, error) {
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
				if shouldMentionUser {
					mdM["mentioned_list"] = MentionUserIDs(robot, alarmMsgType)
					if len(robot.MentionUserTels) > 0 {
						mdM["mentioned_mobile_list"] = robot.MentionUserTels
					}
				}
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
				if shouldMentionUser {
					mdM["mentioned_list"] = MentionUserIDs(robot, alarmMsgType)
					if len(robot.MentionUserTels) > 0 {
						mdM["mentioned_mobile_list"] = robot.MentionUserTels
					}
				}
				mdM["content"] = content

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

func MentionUserIDs(robot conf.CorpWeChat, alarmMsgType int) interface{} {
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

func Robot() (robot conf.CorpWeChat, err error) {
	if d, ok := cropWeChat.Load().(conf.CorpWeChat); ok && d.WebhookUrl != "" && len(d.MentionUserIDs) != 0 {
		robot = d

		return
	}

	err = errors.New("robot is uninitialized")

	return
}

func RobotVote() (robot conf.CorpWeChat, err error) {
	if d, ok := cropWeChatForVote.Load().(conf.CorpWeChat); ok && d.WebhookUrl != "" && len(d.MentionUserIDs) != 0 {
		robot = d

		return
	}

	err = errors.New("robot is uninitialized")

	return
}

func SendCorpWeChatRobotAlarmByRobot(robot conf.CorpWeChat, bs []byte) error {
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

func SendCorpWeChatRobotAlarmForVote(bs []byte) error {
	robot, err := RobotVote()
	if err != nil {
		return err
	}

	return SendCorpWeChatRobotAlarmByRobot(robot, bs)
}
