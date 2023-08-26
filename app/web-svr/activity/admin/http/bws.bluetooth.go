package http

import (
	"encoding/csv"
	"strconv"
	"strings"

	"go-common/library/ecode"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/web-svr/activity/admin/model/lottery"
	bwsmdl "go-gateway/app/web-svr/activity/interface/model/bws"
)

func bwsBluetoothUpAdd(c *bm.Context) {
	param := new(struct {
		Bid int64 `form:"bid" default:"0"`
	})
	if err := c.Bind(param); err != nil {
		return
	}
	if param.Bid == 0 {
		param.Bid = actSrv.GetBid(c)
	}
	file, _, err := c.Request.FormFile("file")
	if err != nil {
		log.Errorc(c, "csv文件解析失败， error(%v)", err)
		c.JSON(nil, ecode.Error(ecode.RequestErr, "csv文件解析失败"))
		return
	}
	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		log.Errorc(c, "csv文件读取析失败， error(%v)", err)
		c.JSON(nil, ecode.Error(ecode.RequestErr, "csv文件读取失败"))
		return
	}
	var ups []*bwsmdl.BluetoothUp
	for _, line := range records {
		if len(line) <= 0 {
			continue
		}
		mid, err := strconv.ParseInt(strings.TrimSpace(line[0]), 10, 64)
		if err != nil || mid == 0 {
			continue
		}
		up := &bwsmdl.BluetoothUp{
			Mid:  mid,
			Key:  line[1],
			Desc: line[2],
		}
		ups = append(ups, up)
	}
	c.JSON(nil, lotterySrv.AddBluetoothUps(c, param.Bid, ups))
}

func bwsBluetoothUpSave(c *bm.Context) {
	param := &lottery.EditBluetoothUpParam{}
	if err := c.Bind(param); err != nil {
		return
	}
	c.JSON(nil, lotterySrv.SaveBluetoothUp(c, param))
}

func bwsBluetoothUpDel(c *bm.Context) {
	param := &lottery.EditBluetoothUpParam{}
	if err := c.Bind(param); err != nil {
		return
	}
	c.JSON(nil, lotterySrv.DelBluetoothUp(c, param))
}

func bwsBluetoothUpList(c *bm.Context) {
	param := &lottery.BluetoothUpListParam{}
	if err := c.Bind(param); err != nil {
		return
	}
	data, count, err := lotterySrv.BluetoothUpList(c, param)
	c.JSON(struct {
		List  []*bwsmdl.BluetoothUp `json:"list"`
		Count int                   `json:"count"`
	}{List: data, Count: count}, err)
}
