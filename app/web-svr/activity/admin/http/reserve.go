package http

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"go-gateway/app/web-svr/activity/admin/model/reserve"
	"go-gateway/app/web-svr/activity/admin/service/exporttask"

	"go-common/library/ecode"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/http/blademaster/binding"
)

func reserveList(c *bm.Context) {
	arg := new(reserve.ParamList)
	if err := c.Bind(arg); err != nil {
		return
	}
	c.JSON(actSrv.ReserveList(c, arg))
}

func addReserve(c *bm.Context) {
	arg := new(reserve.ParamAddReserve)
	if err := c.Bind(arg); err != nil {
		return
	}
	c.JSON(nil, actSrv.AddReserve(c, arg))
}

func importReserve(c *bm.Context) {
	arg := new(reserve.ParamImportReserve)
	if err := c.Bind(arg); err != nil {
		c.JSON(nil, err)
		return
	}
	file, _, err := c.Request.FormFile("file")
	if err != nil {
		log.Errorc(c, "importReserve upload err(%v)", err)
		c.JSON(nil, ecode.RequestErr)
		return
	}
	defer file.Close()
	r := csv.NewReader(file)
	var mids []int64
	if records, err := r.ReadAll(); err != nil {
		log.Errorc(c, "importReserve ioutil.ReadAll err(%v)", err)
		c.JSON(nil, ecode.RequestErr)
		return
	} else {
		mids = make([]int64, 0, len(records))
		for _, record := range records {
			if len(record) >= 1 {
				if mid, err := strconv.ParseInt(record[0], 10, 64); err != nil {
					log.Errorc(c, "importReserve ioutil.ReadAll err(%v)", err)
					c.JSON(nil, ecode.RequestErr)
					return
				} else {
					mids = append(mids, mid)
				}
			}
		}
	}
	go func() {
		c := context.Background()
		err := actSrv.ImportReserve(c, arg, mids)
		if err != nil {
			exporttask.SendWeChatTextMessage(c, arg.Username, fmt.Sprintf("预约数据源%d导入发生异常，异常信息:%s", arg.Sid, err.Error()))
		} else {
			exporttask.SendWeChatTextMessage(c, arg.Username, fmt.Sprintf("预约数据源%d导入完成。", arg.Sid))
		}
	}()
	c.JSON(nil, nil)
}

func reserveScoreUpdate(c *bm.Context) {
	arg := new(reserve.ParamReserveScoreUpdate)
	if err := c.Bind(arg); err != nil {
		return
	}
	c.JSON(nil, actSrv.ReserveScoreUpdate(c, arg))
}

func reserveNotifyUpdate(c *bm.Context) {
	arg := new(reserve.ParamReserveNotifyUpdate)
	if err := c.Bind(arg); err != nil {
		return
	}
	notifies := make([]*reserve.ActSubjectNotify, 0, 0)
	if err := json.Unmarshal([]byte(arg.Notify), &notifies); err != nil {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	for _, notify := range notifies {
		if err := binding.Validator.ValidateStruct(notify); err != nil {
			c.JSON(notify, ecode.Error(ecode.RequestErr, err.Error()))
			return
		}
		notify.Sid = arg.Sid
		notify.Author = arg.Author
		if notify.TemplateID == 0 {
			notify.TemplateID = 1
		}
	}
	c.JSON(nil, actSrv.ReserveNotifyUpdate(c, arg.Sid, notifies))
}

func reserveNotifyDelete(c *bm.Context) {
	arg := new(reserve.ParamReserveNotifyDelete)
	if err := c.Bind(arg); err != nil {
		return
	}
	c.JSON(nil, actSrv.ReserveNotifyDelete(c, arg.Sid, arg.Author, strings.Split(arg.NotifyID, ",")))
}

func isReserveSpringfestival(c *bm.Context) {
	v := new(struct {
		Mid int64 `form:"mid" validate:"min=1"`
		Sid int64 `form:"sid" validate:"min=1"`
	})
	if err := c.Bind(v); err != nil {
		return
	}

	c.JSON(actSrv.Following(c, v.Sid, v.Mid))
}

func isReserveCards(c *bm.Context) {
	v := new(struct {
		Mid int64 `form:"mid" validate:"min=1"`
		Sid int64 `form:"sid" validate:"min=1"`
	})
	if err := c.Bind(v); err != nil {
		return
	}

	c.JSON(actSrv.Following(c, v.Sid, v.Mid))
}

func reserveCounterGroupList(c *bm.Context) {
	v := new(struct {
		Sid int64 `form:"sid" validate:"min=1"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	c.JSON(actSrv.ReserveCounterGroupList(c, v.Sid))
}

func reserveCounterGroupUpdate(c *bm.Context) {
	v := new(reserve.ParamCounterGroupUpdate)
	if err := c.Bind(v); err != nil {
		return
	}
	if v.Dim1 == reserve.CounterGroupDim1Personal {
		v.Threshold = 0
	}
	if v.NodeStr != "" {
		if err := json.Unmarshal([]byte(v.NodeStr), &v.Nodes); err != nil {
			c.JSON(nil, err)
			return
		}
		for _, n := range v.Nodes {
			if err := binding.Validator.ValidateStruct(n); err != nil {
				c.JSON(n, ecode.Error(ecode.RequestErr, err.Error()))
				return
			}
		}
	}
	c.JSON(actSrv.ReserveCounterGroupUpdate(c, v))
}

func reserveCounterNodeList(c *bm.Context) {
	v := new(struct {
		GroupID int64 `form:"group_id" validate:"min=1"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	c.JSON(actSrv.ReserveCounterNodeList(c, v.GroupID))
}
