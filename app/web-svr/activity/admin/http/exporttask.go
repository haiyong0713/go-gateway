package http

import (
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/web-svr/activity/admin/model/exporttask"
	"net/http"
	"strings"
)

func exportTaskAdd(c *bm.Context) {
	req := new(exporttask.ReqExportTaskAdd)
	if err := c.Bind(req); err != nil {
		return
	}
	data := make(map[string]string)
	for key := range c.Request.PostForm {
		if strings.HasPrefix(key, "_") {
			continue
		}
		data[key] = c.Request.PostForm.Get(key)
	}
	if req.ExportType == exporttask.ExportTypeCsv {
		data, err := exportSrv.ExportTaskAdd(c, req, data)
		if err != nil {
			c.JSON(nil, err)
		} else {
			task := data.(*exporttask.ExportTask)
			if task.DownURL != "" {
				c.Redirect(http.StatusFound, task.DownURL)
			} else {
				c.JSON(task, err)
			}
		}
	} else {
		c.JSON(exportSrv.ExportTaskAdd(c, req, data))
	}
}

func exportTaskState(c *bm.Context) {
	req := new(exporttask.ReqExportTaskState)
	if err := c.Bind(req); err != nil {
		return
	}
	c.JSON(exportSrv.ExportTaskState(c, req.TaskID))
}

func exportTaskRedo(c *bm.Context) {
	req := new(exporttask.ReqExportTaskState)
	if err := c.Bind(req); err != nil {
		return
	}
	c.JSON(exportSrv.ExportTaskRedo(c, req.TaskID))
}

func exportTaskList(c *bm.Context) {
	req := new(exporttask.ReqExportTaskList)
	if err := c.Bind(req); err != nil {
		return
	}
	c.JSON(exportSrv.ExportTaskList(c, req))
}

func exportTaskWeChatUserID(c *bm.Context) {
	c.JSON(exportSrv.GetUserIDMap(), nil)
}

func exportTaskWeChatUpdateMemberInfo(c *bm.Context) {
	err := exportSrv.UpdateMemberInfo()
	c.JSON(exportSrv.GetUserIDMap(), err)
}
