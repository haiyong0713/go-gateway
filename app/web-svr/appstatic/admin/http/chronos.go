package http

import (
	"encoding/json"
	"io/ioutil"
	"time"

	"go-common/library/ecode"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/queue/databus/report"

	"go-gateway/app/web-svr/appstatic/admin/model"

	"github.com/pkg/errors"
)

const (
	logAction = "publish"
)

// for other systems
func uploadChronos(c *bm.Context) {
	var content []byte
	if err := func() error {
		file, _, err := c.Request.FormFile("file")
		if err != nil {
			return err
		}
		defer file.Close()
		if content, err = ioutil.ReadAll(file); err != nil {
			return err
		}
		if len(content) >= model.BfsMaxSize {
			return errors.New("bfs最大允许上传不超过20M的文件")
		}
		return nil
	}(); err != nil {
		res := map[string]interface{}{}
		res["message"] = err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(apsSvc.UploadChronos(c, content))
}

func saveChronos(c *bm.Context) {
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}

	bytes, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	rules := make([]*model.ChronosRule, 0)
	if err = func() error {
		if err := json.Unmarshal(bytes, &rules); err != nil {
			return err
		}
		if errMsg := model.RulesValidate(rules); errMsg != "" {
			return errors.New(errMsg)
		}
		return nil
	}(); err != nil {
		res := map[string]interface{}{}
		res["message"] = err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if err = apsSvc.SaveRules(c, rules); err != nil {
		c.JSON(nil, err)
		return
	}
	if err := report.Manager(&report.ManagerInfo{
		Uname:    userName,
		UID:      0,
		Business: 580,
		Type:     0,
		Oid:      0,
		Action:   logAction,
		Ctime:    time.Now(),
		Index:    []interface{}{},
		Content: map[string]interface{}{
			"rules": rules,
		},
	}); err != nil {
		log.Error("ReportManager userName %s, err %v", userName, err)
		return
	}
	log.Warn("ReportManager userName %s Succ", userName)
	c.JSON(nil, nil)
}

func listChronos(c *bm.Context) {
	c.JSON(apsSvc.ListChronos(c))
}
