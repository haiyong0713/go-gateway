package http

import (
	"encoding/csv"
	"github.com/pkg/errors"
	"io/ioutil"
	"strconv"
	"strings"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/net/http/blademaster"
	"go-gateway/app/web-svr/activity/admin/model/question"
)

func baseList(c *blademaster.Context) {
	v := new(struct {
		Pn int64 `form:"pn" default:"1" validate:"min=1"`
		Ps int64 `form:"ps" default:"20" validate:"min=1,max=50"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	list, count, err := quesSrv.BaseList(c, v.Pn, v.Ps)
	if err != nil {
		c.JSON(nil, err)
		return
	}
	data := make(map[string]interface{}, 2)
	page := map[string]int64{
		"num":   v.Pn,
		"size":  v.Ps,
		"total": count,
	}
	data["page"] = page
	data["list"] = list
	c.JSON(data, nil)
}

func baseItem(c *blademaster.Context) {
	v := new(struct {
		ID int64 `form:"id" validate:"min=1"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	c.JSON(quesSrv.BaseItem(c, v.ID))
}

func baseAdd(c *blademaster.Context) {
	v := new(question.AddBaseArg)
	if err := c.Bind(v); err != nil {
		return
	}
	c.JSON(nil, quesSrv.AddBase(c, v))
}

func baseSave(c *blademaster.Context) {
	v := new(question.SaveBaseArg)
	if err := c.Bind(v); err != nil {
		return
	}
	c.JSON(nil, quesSrv.SaveBase(c, v))
}

func detailList(c *blademaster.Context) {
	v := new(struct {
		BaseID int64 `form:"base_id" validate:"min=1"`
		Pn     int64 `form:"pn" default:"1" validate:"min=1"`
		Ps     int64 `form:"ps" default:"20" validate:"min=1,max=50"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	list, count, err := quesSrv.DetailList(c, v.BaseID, v.Pn, v.Ps)
	if err != nil {
		c.JSON(nil, err)
		return
	}
	data := make(map[string]interface{}, 2)
	page := map[string]int64{
		"num":   v.Pn,
		"size":  v.Ps,
		"total": count,
	}
	data["page"] = page
	data["list"] = list
	c.JSON(data, nil)
}

func detailAdd(c *blademaster.Context) {
	v := new(question.AddDetailArg)
	if err := c.Bind(v); err != nil {
		return
	}
	c.JSON(nil, quesSrv.AddDetail(c, v))
}

func detailSave(c *blademaster.Context) {
	v := new(question.SaveDetailArg)
	if err := c.Bind(v); err != nil {
		return
	}
	c.JSON(nil, quesSrv.SaveDetail(c, v))
}

func detailDel(c *blademaster.Context) {
	v := new(struct {
		ID int64 `form:"id" validate:"min=1"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	c.JSON(nil, quesSrv.UpdateDetailState(c, v.ID, question.StateOffline))
}

func detailOnline(c *blademaster.Context) {
	v := new(struct {
		ID int64 `form:"id" validate:"min=1"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	c.JSON(nil, quesSrv.UpdateDetailState(c, v.ID, question.StateOnline))
}

func importDetailCSV(c *blademaster.Context) {
	var (
		err  error
		data []byte
	)
	v := new(struct {
		BaseID int64 `form:"base_id" validate:"min=1"`
	})
	if err = c.Bind(v); err != nil {
		return
	}
	file, _, err := c.Request.FormFile("file")
	if err != nil {
		log.Error("importDetailCSV upload err(%v)", err)
		c.JSON(nil, ecode.RequestErr)
		return
	}
	defer file.Close()
	data, err = ioutil.ReadAll(file)
	if err != nil {
		log.Error("importDetailCSV ioutil.ReadAll err(%v)", err)
		c.JSON(nil, ecode.RequestErr)
		return
	}
	r := csv.NewReader(strings.NewReader(string(data)))
	records, err := r.ReadAll()
	if err != nil {
		log.Error("importDetailCSV r.ReadAll() err(%v)", err)
		c.JSON(nil, ecode.RequestErr)
		return
	}
	if l := len(records); l > 300 || l <= 1 {
		log.Error("importDetailCSV len(%d) err", l)
		c.JSON(nil, ecode.RequestErr)
		return
	}
	var args []*question.AddDetailArg
Loop:
	for i, row := range records {
		// continue first row
		if i == 0 {
			continue
		}
		// import csv state online
		arg := &question.AddDetailArg{State: question.StateOnline, BaseID: v.BaseID}
		for field, value := range row {
			value = strings.TrimSpace(value)
			switch field {
			case 0:
				if value == "" {
					log.Warn("importDetailCSV name empty(%s)", value)
					continue Loop
				}
				arg.Name = value
			case 1:
				if value == "" {
					log.Warn("importDetailCSV right answer empty(%s)", value)
					continue Loop
				}
				arg.RightAnswer = value
			case 2:
				if value == "" {
					log.Warn("importDetailCSV wrang answer empty(%s)", value)
					continue Loop
				}
				arg.WrongAnswer = value
			case 3:
				if strings.HasPrefix(value, "http") {
					arg.Pic = value
				}
			case 4:
				if value != "" {
					if arg.Attribute, err = strconv.ParseInt(value, 10, 64); err != nil {
						c.JSON(nil, errors.Wrapf(ecode.RequestErr, "parse Attribute err:%+v", err))
						return
					}
				}
			}
		}

		args = append(args, arg)
	}
	if len(args) == 0 {
		log.Errorc(c, "importDetailCSV args no after filter")
		c.JSON(nil, ecode.RequestErr)
		return
	}
	c.JSON(nil, quesSrv.BatchAddDetail(c, v.BaseID, args))
}
