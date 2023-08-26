package http

import (
	"encoding/json"
	bm "go-common/library/net/http/blademaster"
)

type Upload struct {
	ID         string `json:"id"`
	Mid        string `json:"mid"`
	ImagePath  string `json:"img_path"`
	Platfrom   string `json:"platfrom"`
	Status     string `json:"status"`
	Deleted    string `json:"deleted"`
	UploadDate string `json:"upload_date"`
}

func addLog(c *bm.Context) {
	param := &struct {
		Name   string `form:"name" validate:"required"`
		Action string `form:"action" validate:"required"`
		OID    int64  `form:"oid" validate:"required"`
		Obj    string `form:"obj" validate:"required"`
	}{}
	if err := c.Bind(param); err != nil {
		return
	}
	up := Upload{}
	//nolint:errcheck
	json.Unmarshal([]byte(param.Obj), &up)
	spcSvc.AddLog(param.Name, 0, param.OID, param.Action, up)
	c.JSON(nil, nil)
}

func clearMessage(c *bm.Context) {
	arg := new(struct {
		Type   int    `form:"type" validate:"min=1,max=4"`
		Reason int    `form:"reason" validate:"required"`
		Mid    int64  `form:"mid" validate:"min=1"`
		ID     int64  `form:"id" validate:"min=0"`
		Uname  string `form:"-"`
		Uid    int64  `form:"-"`
	})
	if err := c.Bind(arg); err != nil {
		return
	}
	if usernameCtx, ok := c.Get("username"); ok {
		arg.Uname = usernameCtx.(string)
	}
	if uidCtx, ok := c.Get("uid"); ok {
		arg.Uid = uidCtx.(int64)
	}
	c.JSON(nil, spcSvc.ClearMsg(c, arg.Type, arg.Reason, arg.Mid, arg.ID, arg.Uid, arg.Uname))
}
