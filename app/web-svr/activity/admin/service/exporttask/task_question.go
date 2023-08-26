package exporttask

import (
	"context"
	"encoding/json"
	"fmt"
	"go-gateway/app/web-svr/activity/admin/model"
	"net/http"
	"strings"

	"go-common/library/database/sql"
	"go-common/library/log"
	"go-gateway/app/web-svr/activity/admin/component"
)

type taskExportQuestion struct {
	questionSQL *taskExportSQL
	Append      []appended
	formatter   *simpleFormatter
	idx         map[string]struct{}
}

func (t *taskExportQuestion) decodeMessage(c context.Context, message string) map[string]string {
	one := map[string]string{}
	if len(message) > 0 {
		for _, topic := range strings.Split(message, "\\t") {
			detail := struct {
				Value string          `json:"value"`
				Text  json.RawMessage `json:"text"`
				Name  interface{}     `json:"name"`
			}{}
			if err := json.Unmarshal([]byte(topic), &detail); err != nil {
				log.Errorc(c, "taskExportQuestion json.Unmarshal([]byte(topic), &detail) err[%v]", err)
				continue
			}
			value := detail.Value
			var text string
			if err := json.Unmarshal(detail.Text, &text); err != nil {
				var textArr []struct {
					Name string `json:"name"`
					Text string `json:"text"`
				}
				if err := json.Unmarshal(detail.Text, &textArr); err != nil {
					log.Errorc(c, "taskExportQuestion json.Unmarshal(detail.Text, &textArr) err[%v]", err)
				} else {
					for _, textObj := range textArr {
						text = text + " " + textObj.Text
					}
				}
			}
			text = strings.TrimSpace(text)
			if value != "" {
				if text != "" {
					value = value + " " + text
				}
			} else {
				value = text
			}
			title := fmt.Sprintf("Q-%v", detail.Name)
			if _, ok := t.idx[title]; !ok {
				t.idx[title] = struct{}{}
				t.formatter.Output = append(t.formatter.Output, &ExportOutputField{
					Name: title,
				})
			}
			one[title] = value
		}
	}
	return one
}

func (t *taskExportQuestion) Do(c context.Context, db *sql.DB, data map[string]string, writer *readerWriter) error {
	// 查询活动数据，判断活动类型，计算表名
	actSubject := new(model.ActSubject)
	if err := component.GlobalOrm.Where("id = ?", data["sid"]).Last(actSubject).Error; err != nil {
		log.Errorc(c, "Do s.DB.Where(id = ?, %d).Last(%v) error(%v)", data["sid"], actSubject, err)
		return err
	}
	if actSubject.IsQuestionnaire() {
		data["like_content_table"] = "like_content_new"
	} else {
		data["like_content_table"] = "like_content"
	}
	// 查询问卷数据
	return t.questionSQL.GetData(c, db, data, func(question []map[string]string) error {
		// 结构化问卷数据
		taskRet := make([]map[string]string, 0, len(question))
		for _, q := range question {
			one := map[string]string{
				"sid":   q["sid"],
				"mid":   q["mid"],
				"mtime": q["mtime"],
			}
			for k, v := range t.decodeMessage(c, q["message"]) {
				one[k] = v
			}
			taskRet = append(taskRet, one)
		}
		if len(t.Append) > 0 {
			for _, apd := range t.Append {
				taskRet = apd.Append(c, taskRet)
			}
		}
		dataSet, err := t.formatter.Formatter(c, taskRet)
		if err != nil {
			return err
		}
		writer.Put(dataSet)
		return nil
	})
}

func (t *taskExportQuestion) Header(c context.Context, data map[string]string) ([]string, error) {
	formatter := &simpleFormatter{
		Output: []*ExportOutputField{
			{
				Name:  "sid",
				Title: "数据源ID",
			},
			{
				Name: "mid",
			},
			{
				Name:   "mtime",
				Title:  "日期",
				Format: formatTimeString,
			},
		},
	}
	t.formatter = formatter
	t.idx = map[string]struct{}{}
	url := fmt.Sprintf("https://activity.hdslb.com/blackboard/static/questionconfig/%s/question_map.json", data["sid"])
	if req, err := http.NewRequest("GET", url, nil); err != nil {
		log.Errorc(c, "Header http.NewRequest url(%s) error(%v)", url, err)
		return nil, err
	} else {
		res := struct {
			Questions []struct {
				Name  string `json:"name"`
				Title string `json:"title"`
			} `json:"questions"`
		}{}
		if err := httpClient.Do(c, req, &res); err != nil {
			log.Errorc(c, "Header httpClient.Do url(%s) error(%v)", url, err)
			p := make([]struct {
				Message string `json:"message"`
			}, 0, 1000)
			if err := component.GlobalOrm.Raw("SELECT likes.id,sid,mid,wid,likes.mtime,message FROM likes force index (ix_like_0) INNER JOIN like_content ON likes.id = like_content.id WHERE likes.sid = ? LIMIT 1000", data["sid"]).Scan(&p).Error; err != nil {
				return nil, err
			}
			for _, q := range p {
				t.decodeMessage(c, q.Message)
			}
		}
		for _, p := range res.Questions {
			title := fmt.Sprintf("Q-%v", p.Name)
			t.idx[title] = struct{}{}
			t.formatter.Output = append(t.formatter.Output, &ExportOutputField{
				Name:  title,
				Title: p.Title,
			})
		}
	}
	return formatter.Headers(), nil
}
