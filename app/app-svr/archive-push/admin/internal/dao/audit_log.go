package dao

import (
	"context"
	"github.com/pkg/errors"
	"net/url"
	"strconv"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/queue/databus/actionlog"

	"go-gateway/app/app-svr/archive-push/admin/internal/model"
)

const (
	appID        = "log_audit"
	logSearchURL = "/x/admin/search/log"
)

// SearchAuditLog 搜索行为日志
func (d *Dao) SearchAuditLog(params *model.AuditLogSearchParams) (res *model.AuditLogSearchResRawData, err error) {
	rawRes := &model.AuditLogSearchResRaw{}
	query := url.Values{}
	query.Set("appid", appID)
	if params.UName != "" {
		query.Set("uname", params.UName)
	}
	if params.UID != 0 {
		query.Set("uid", strconv.FormatInt(params.UID, 10))
	}
	if params.Business != 0 {
		query.Set("business", strconv.FormatInt(int64(params.Business), 10))
	} else {
		query.Set("business", strconv.FormatInt(int64(model.BusinessIDBatch), 10))
	}
	if params.Type != 0 {
		query.Set("type", strconv.FormatInt(int64(params.Type), 10))
	}
	if params.Int0 != 0 {
		query.Set("int_0", strconv.FormatInt(params.Int0, 10))
	}
	if params.Int1 != 0 {
		query.Set("int_1", strconv.FormatInt(params.Int1, 10))
	}

	query.Set("order", "ctime")
	if err = d.bmClient.Get(context.Background(), d.hosts.Manager+logSearchURL, "", query, rawRes); err != nil {
		log.Error("archive-push-admin.dao.SearchAuditLog.Get Error (%v)", err)
		return
	}
	if rawRes.Code != 0 {
		err = errors.WithMessage(ecode.RequestErr, rawRes.Message)
		return nil, err
	}
	res = rawRes.Data
	return
}

// AddLogs add action logs
func (d *Dao) AddAuditLog(params *model.AuditLogInitParams) (err error) {
	mInfo := &actionlog.ManagerInfo{
		Business: params.Business,                                // 业务 id, 请填写 info 中对应的业务 id
		Uname:    params.UName,                                   //全匹配, 默认:审核人员内网name, 业务方可自定义
		UID:      params.UID,                                     //全匹配, 默认:审核人员内网uid, 业务方可自定义
		Type:     params.Type,                                    //全匹配, 默认:操作对象的类型, 业务方可自定义
		Oid:      params.OID,                                     //全匹配, 默认:操作对象的id, 业务方可自定义
		Action:   params.Action,                                  //全匹配, 默认:具体操作类型，如打回, 业务方可自定义
		Ctime:    params.CTime,                                   // 可以时间排序
		Index:    params.Index,                                   // 为预留自定义字段, 根据传入的数据格式 string转化为 str_0~str_9(全匹配),
		Content:  map[string]interface{}{"json": params.Content}, // 数据只展示, 不参与搜索, 在 es 中保存为一个 json 字符串
	}

	// 异步请求, batchSize 条数(默认:10)据或每隔 workInterval 时间(默认:1s)上传一次数据,
	// 应用奔溃,可能造成数据丢失
	// 上传三次数据未成功, 会丢弃数据, 造成数据的丢失
	err = actionlog.AsyncManager(mInfo)
	return
}
