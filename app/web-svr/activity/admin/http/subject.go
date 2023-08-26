package http

import (
	"gopkg.in/go-playground/validator.v9"
	"reflect"
	"time"

	tagrpc "git.bilibili.co/bapis/bapis-go/community/interface/tag"
	"go-common/library/ecode"
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/web-svr/activity/admin/model"
)

func listInfosAll(c *bm.Context) {
	arg := new(model.ListSub)
	if err := c.Bind(arg); err != nil {
		return
	}
	c.JSON(actSrv.SubjectList(c, arg))
}

func subInfos(c *bm.Context) {
	arg := new(struct {
		Sids []int64 `form:"sids,split" validate:"required,min=1,max=50,dive,gt=0"`
	})
	if err := c.Bind(arg); err != nil {
		return
	}
	c.JSON(actSrv.SubInfos(c, arg.Sids))
}

func videoList(c *bm.Context) {
	c.JSON(actSrv.VideoList(c))
}

func addActSubject(c *bm.Context) {
	arg := new(model.AddList)
	if err := c.Bind(arg); err != nil {
		return
	}
	if arg.Type <= 0 {
		c.JSON(struct {
		}{}, ecode.Error(ecode.RequestErr, "type 不能为空"))
		return
	}
	c.JSON(actSrv.AddActSubject(c, arg, tagrpc.TagType_TypeBiliActivity))
}

func updateInfoAll(c *bm.Context) {
	type upStr struct {
		model.AddList
		Sid int64 `form:"id" validate:"min=1"`
	}
	arg := new(upStr)
	if err := c.Bind(arg); err != nil {
		return
	}
	elem := reflect.ValueOf(arg.AddList)
	relType := elem.Type()
	data := make(map[string]interface{})
	// 老php代码直接取data，无数据类型，为了兼容老的写法，写的好恶心
	for i := 0; i < relType.NumField(); i++ {
		key := elem.Type().Field(i).Tag.Get("form")
		if _, ok := c.Request.Form[key]; ok {
			tf := elem.Type().Field(i).Tag.Get("time_format")
			if tf != "" {
				v, err := time.ParseInLocation(tf, elem.Field(i).String(), time.Local)
				if err != nil {
					data[key] = time.Unix(0, 0)
				} else {
					data[key] = v
				}
			} else {
				data[key] = elem.Field(i).Interface()
			}
		}
	}
	delete(data, "id")
	c.JSON(actSrv.UpActSubject(c, &arg.AddList, arg.Sid, data))
}

func subPro(c *bm.Context) {
	type subStr struct {
		Sid int64 `form:"sid" validate:"min=1"`
	}
	arg := new(subStr)
	if err := c.Bind(arg); err != nil {
		return
	}
	c.JSON(actSrv.SubProtocol(c, arg.Sid))
}

func timeConf(c *bm.Context) {
	type subStr struct {
		Sid int64 `form:"sid" validate:"required"`
	}
	arg := new(subStr)
	if err := c.Bind(arg); err != nil {
		return
	}
	c.JSON(actSrv.TimeConf(c, arg.Sid))
}

func article(c *bm.Context) {
	type subStr struct {
		Aids []int64 `form:"aids,split" validate:"min=1,required"`
	}
	arg := new(subStr)
	if err := c.Bind(arg); err != nil {
		return
	}
	c.JSON(actSrv.GetArticleMetas(c, arg.Aids))
}

func optVideoList(c *bm.Context) {
	var (
		err  error
		cnt  int
		list = make([]*model.SubProtocol, 0)
	)
	arg := new(model.OptVideoListSub)
	if err = c.Bind(arg); err != nil {
		return
	}
	if arg.Types != "" {
		if err = validator.New().Var(arg.Types, "number"); err != nil {
			c.JSON(nil, err)
			return
		}
	}
	if list, cnt, err = actSrv.OptVideoList(c, arg); err != nil {
		c.JSON(nil, err)
		return
	}
	data := map[string]interface{}{
		"page": map[string]int{
			"num":   arg.Page,
			"size":  arg.PageSize,
			"total": cnt,
		},
		"list": list,
	}
	c.JSON(data, nil)
}

func subjectRules(c *bm.Context) {
	v := new(struct {
		Sid   int64 `form:"sid" validate:"min=1"`
		State int64 `form:"state"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	rules, err := actSrv.SubjectRules(c, v.Sid, v.State)
	if err != nil {
		c.JSON(nil, err)
		return
	}
	c.JSON(map[string]interface{}{"list": rules}, nil)
}

func subjectRuleUserState(c *bm.Context) {
	v := new(struct {
		Mid int64 `form:"mid" validate:"min=1"`
		Sid int64 `form:"sid" validate:"min=1"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	data, err := actSrv.SubjectRuleUserState(c, v.Mid, v.Sid)
	if err != nil {
		c.JSON(nil, err)
		return
	}
	c.JSON(data, nil)
}

func addSubjectRule(c *bm.Context) {
	v := new(model.AddSubjectRuleArg)
	if err := c.Bind(v); err != nil {
		return
	}
	c.JSON(nil, actSrv.AddSubjectRule(c, v))
}

func saveSubjectRule(c *bm.Context) {
	v := new(model.SaveSubjectRuleArg)
	if err := c.Bind(v); err != nil {
		return
	}
	c.JSON(nil, actSrv.SaveSubjectRule(c, v))
}

func upSubRuleState(c *bm.Context) {
	v := new(struct {
		ID    int64 `form:"id" validate:"min=1"`
		State int64 `form:"state" validate:"min=1,max=3"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	c.JSON(nil, actSrv.UpSubRuleState(c, v.ID, v.State))
}

func addPush(c *bm.Context) {
	v := new(model.SubjectTunnelParam)
	if err := c.Bind(v); err != nil {
		return
	}
	if err := actSrv.AddPush(c, v); err != nil {
		res := map[string]interface{}{}
		res["message"] = "添加推送配制失败 " + err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON("", nil)
}

func editPush(c *bm.Context) {
	v := new(model.SubjectTunnelParam)
	if err := c.Bind(v); err != nil {
		return
	}
	if err := actSrv.EditPush(c, v); err != nil {
		res := map[string]interface{}{}
		res["message"] = "修改推送配制失败 " + err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON("", nil)
}

func startPush(c *bm.Context) {
	var err error
	v := new(struct {
		Sid int64 `json:"sid" form:"sid" validate:"min=1,required"`
	})
	if err = c.Bind(v); err != nil {
		return
	}
	if err = actSrv.StartPush(c, v.Sid); err != nil {
		res := map[string]interface{}{}
		res["message"] = "开始推送失败 " + err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(nil, nil)
}

func infoPush(c *bm.Context) {
	var (
		res interface{}
		err error
	)
	v := new(struct {
		Sid int64 `json:"sid" form:"sid" validate:"min=1,required"`
	})
	if err = c.Bind(v); err != nil {
		return
	}
	if res, err = actSrv.InfoPush(c, v.Sid); err != nil {
		res := map[string]interface{}{}
		res["message"] = "推送配制信息失败 " + err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(res, nil)
}

func pushTemplate(c *bm.Context) {
	v := new(struct {
		Type int64 `form:"type" validate:"min=1"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	c.JSON(actSrv.HaveTemplate(c, v.Type))
}

func fixTunnelAdd(c *bm.Context) {
	var err error
	type subStr struct {
		Sid int64 `form:"sid" validate:"min=1"`
	}
	arg := new(subStr)
	if err = c.Bind(arg); err != nil {
		return
	}
	if err = actSrv.TunnelGroupAdd(c, arg.Sid); err != nil {
		res := map[string]interface{}{}
		res["message"] = "添加人群包失败 " + err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(nil, nil)
}

func fixTunnelUp(c *bm.Context) {
	var err error
	type subStr struct {
		Sid int64 `form:"sid" validate:"min=1"`
	}
	arg := new(subStr)
	if err = c.Bind(arg); err != nil {
		return
	}
	if err = actSrv.TunnelGroupUp(c, arg.Sid); err != nil {
		res := map[string]interface{}{}
		res["message"] = "修复人群包失败 " + err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(nil, nil)
}

func fixTunnelDel(c *bm.Context) {
	var err error
	type subStr struct {
		Sid int64 `form:"sid" validate:"min=1"`
	}
	arg := new(subStr)
	if err = c.Bind(arg); err != nil {
		return
	}
	if err = actSrv.TunnelGroupDel(c, arg.Sid); err != nil {
		res := map[string]interface{}{}
		res["message"] = "删除人群包失败 " + err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(nil, nil)
}
