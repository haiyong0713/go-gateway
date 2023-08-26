package http

import (
	"bytes"
	"encoding/csv"
	"fmt"
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/web-svr/activity/admin/model/datamanage"
	"strings"
	"time"
)

func dataManageSelectData(c *bm.Context) (*datamanage.ResDataManageSelect, error) {
	req := new(datamanage.ReqDataManageSelect)
	if err := c.Bind(req); err != nil {
		return nil, err
	}
	where := make(map[string]interface{})
	for key := range c.Request.Form {
		if strings.HasPrefix(key, "_") {
			continue
		}
		where[key] = c.Request.Form.Get(key)
	}
	return dataMgeSrv.DataManageSelect(c, req, where)
}

func dataManageExport(c *bm.Context) {
	res, err := dataManageSelectData(c)
	if err != nil {
		c.JSON(res, err)
		return
	}
	b := &bytes.Buffer{}
	b.WriteString("\xEF\xBB\xBF")
	wr := csv.NewWriter(b)
	wr.Write(res.Columns)
	for _, r := range res.List {
		record := make([]string, 0, len(res.Columns))
		for _, k := range res.Columns {
			record = append(record, fmt.Sprint(r[k]))
		}
		wr.Write(record)
	}
	wr.Flush()
	c.Writer.Header().Set("Content-Type", "text/csv")
	c.Writer.Header().Set("Content-Disposition", fmt.Sprintf("attachment;filename=%s.%s.csv", res.Req.Table, time.Now().Format("20060102")))
	tet := b.String()
	c.String(200, tet)
}

func dataManageSelect(c *bm.Context) {
	c.JSON(dataManageSelectData(c))
}

func dataManageUpdate(c *bm.Context) {
	req := new(datamanage.ReqDataManageUpdate)
	if err := c.Bind(req); err != nil {
		c.JSON(nil, err)
		return
	}
	file, _, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(nil, err)
		return
	}
	defer file.Close()
	reader := csv.NewReader(file)
	c.JSON(dataMgeSrv.DataManageUpdate(c, req, reader))
}

func dataManageDiff(c *bm.Context) {
	req := new(datamanage.ReqDataManageUpdate)
	if err := c.Bind(req); err != nil {
		c.JSON(nil, err)
		return
	}
	file, _, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(nil, err)
		return
	}
	defer file.Close()
	reader := csv.NewReader(file)
	c.JSON(dataMgeSrv.DataManageDiff(c, req, reader))
}
