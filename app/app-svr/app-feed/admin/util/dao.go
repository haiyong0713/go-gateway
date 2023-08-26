package util

import (
	"strconv"
	"strings"
	"time"

	bm "go-common/library/net/http/blademaster"
	"go-common/library/queue/databus/actionlog"
	"go-common/library/queue/databus/report"
	"go-gateway/app/app-svr/app-feed/admin/model/common"
)

// AddLog add action log
func AddLog(id int, uname string, uid int64, oid int64, action string, obj interface{}, index ...interface{}) (err error) {
	//nolint:errcheck
	report.Manager(&report.ManagerInfo{
		Uname:    uname,
		UID:      uid,
		Business: id,
		Type:     0,
		Oid:      oid,
		Action:   action,
		Ctime:    time.Now(),
		// extra
		Index: index,
		Content: map[string]interface{}{
			"json": obj,
		},
	})
	return
}

// AddLogs add action logs
func AddLogs(logtype int, uname string, uid int64, oid int64, action string, obj interface{}) (err error) {
	//nolint:errcheck
	report.Manager(&report.ManagerInfo{
		Uname:    uname,
		UID:      uid,
		Business: common.BusinessID,
		Type:     logtype,
		Oid:      oid,
		Action:   action,
		Ctime:    time.Now(),
		// extra
		Index: []interface{}{},
		Content: map[string]interface{}{
			"json": obj,
		},
	})
	return
}

// AddWebModuleLogs add action logs
func AddWebModuleLogs(logtype int, uname string, uid int64, oid int64, action string, query string, obj interface{}) (err error) {
	//nolint:errcheck
	report.Manager(&report.ManagerInfo{
		Uname:    uname,
		UID:      uid,
		Business: common.BusinessID,
		Type:     logtype,
		Oid:      oid,
		Action:   action,
		Ctime:    time.Now(),
		// extra
		Index: []interface{}{query},
		Content: map[string]interface{}{
			"json": obj,
		},
	})
	return
}

// AddWebModuleLogs add action logs
func AddOgvLogs(uname string, uid int64, oid int64, action, query, title string) (err error) {
	//nolint:errcheck
	report.Manager(&report.ManagerInfo{
		Uname:    uname,
		UID:      uid,
		Business: common.BusinessID,
		Type:     common.LogOgvModule,
		Oid:      oid,
		Action:   action,
		Ctime:    time.Now(),
		// extra
		Index: []interface{}{query, title},
		Content: map[string]interface{}{
			"title": title,
			"query": query,
		},
	})
	return
}

// AddResourceCardLogs add brand blacklist logs
func AddResourceCardLogs(uname string, uid int64, oid int64, oType int, action string, before, after, affected interface{}) (err error) {
	content := map[string]interface{}{
		"before":   before,
		"after":    after,
		"affected": affected,
	}
	mInfo := &actionlog.ManagerInfo{
		Business: 1050,       // 业务 id, 请填写 info 中对应的业务 id
		Uname:    uname,      //全匹配, 默认:审核人员内网name, 业务方可自定义
		UID:      uid,        //全匹配, 默认:审核人员内网uid, 业务方可自定义
		Type:     oType,      //全匹配, 默认:操作对象的类型, 业务方可自定义
		Oid:      oid,        //全匹配, 默认:操作对象的id, 业务方可自定义
		Action:   action,     //全匹配, 默认:具体操作类型，如打回, 业务方可自定义
		Ctime:    time.Now(), // 可以时间排序
		Content:  content,    // 数据只展示, 不参与搜索, 在 es 中保存为一个 json 字符串
	}

	// 同步请求
	return actionlog.Manager(mInfo)
}

// AddBrandBlacklistLogs add brand blacklist logs
func AddBrandBlacklistLogs(uname string, uid int64, oid int64, action string, before, after, affected interface{}) (err error) {
	content := map[string]interface{}{
		"before":   before,
		"after":    after,
		"affected": affected,
	}
	mInfo := &actionlog.ManagerInfo{
		Business: 960,        // 业务 id, 请填写 info 中对应的业务 id
		Uname:    uname,      //全匹配, 默认:审核人员内网name, 业务方可自定义
		UID:      uid,        //全匹配, 默认:审核人员内网uid, 业务方可自定义
		Oid:      oid,        //全匹配, 默认:操作对象的id, 业务方可自定义
		Action:   action,     //全匹配, 默认:具体操作类型，如打回, 业务方可自定义
		Ctime:    time.Now(), // 可以时间排序
		Content:  content,    // 数据只展示, 不参与搜索, 在 es 中保存为一个 json 字符串
	}

	// 同步请求
	return actionlog.Manager(mInfo)
}

// AddWebRcmdCardLogs add tianma action logs
func AddWebRcmdCardLogs(uname string, uid int64, oid int64, action string, param interface{}) (err error) {
	// web特殊卡的Type取值为5，参考文档: http://bapi.bilibili.co/project/5510/interface/api/233518
	mInfo := &actionlog.ManagerInfo{
		Business: 930,                                    // 业务 id, 请填写 info 中对应的业务 id
		Type:     5,                                      //全匹配, 默认:操作对象的类型, 业务方可自定义
		Uname:    uname,                                  //全匹配, 默认:审核人员内网name, 业务方可自定义
		UID:      uid,                                    //全匹配, 默认:审核人员内网uid, 业务方可自定义
		Oid:      oid,                                    //全匹配, 默认:操作对象的id, 业务方可自定义
		Action:   action,                                 //全匹配, 默认:具体操作类型，如打回, 业务方可自定义
		Ctime:    time.Now(),                             // 可以时间排序
		Content:  map[string]interface{}{"param": param}, // 数据只展示, 不参与搜索, 在 es 中保存为一个 json 字符串
	}

	return actionlog.Manager(mInfo)
}

// AddFrontpageConfigLogs 版头配置管理日志
func AddFrontpageConfigLogs(uname string, uid int64, oid int64, action string, logType int, params []interface{}, extra interface{}) (err error) {
	mInfo := &actionlog.ManagerInfo{
		Business: 1060,       // 业务 id, 请填写 info 中对应的业务 id
		Type:     logType,    //全匹配, 默认:操作对象的类型, 业务方可自定义
		Uname:    uname,      //全匹配, 默认:审核人员内网name, 业务方可自定义
		UID:      uid,        //全匹配, 默认:审核人员内网uid, 业务方可自定义
		Oid:      oid,        //全匹配, 默认:操作对象的id, 业务方可自定义
		Action:   action,     //全匹配, 默认:具体操作类型，如打回, 业务方可自定义
		Ctime:    time.Now(), // 可以时间排序
		Index:    params,
		Content:  map[string]interface{}{"json": extra}, // 数据只展示, 不参与搜索, 在 es 中保存为一个 json 字符串
	}

	return actionlog.Manager(mInfo)
}

func AddPackagePushLog(uname string, uid int64, oid int64, action string, logType int, params []interface{}, extra interface{}) (err error) {
	mInfo := &actionlog.ManagerInfo{
		Business: 1190,       // 业务 id, 请填写 info 中对应的业务 id
		Type:     logType,    //全匹配, 默认:操作对象的类型, 业务方可自定义
		Uname:    uname,      //全匹配, 默认:审核人员内网name, 业务方可自定义
		UID:      uid,        //全匹配, 默认:审核人员内网uid, 业务方可自定义
		Oid:      oid,        //全匹配, 默认:操作对象的id, 业务方可自定义
		Action:   action,     //全匹配, 默认:具体操作类型，如打回, 业务方可自定义
		Ctime:    time.Now(), // 可以时间排序
		Index:    params,
		Content:  map[string]interface{}{"operation": extra}, // 数据只展示, 不参与搜索, 在 es 中保存为一个 json 字符串
	}

	return actionlog.Manager(mInfo)
}

// UserInfo get login userinfo
func UserInfo(c *bm.Context) (uid int64, username string) {
	if nameInter, ok := c.Get("username"); ok {
		username = nameInter.(string)
	}
	if uidInter, ok := c.Get("uid"); ok {
		uid = uidInter.(int64)
	}
	if username == "" {
		cookie, _ := c.Request.Cookie("username")
		if cookie == nil || cookie.Value == "" {
			return
		}
		username = cookie.Value
		cookie, _ = c.Request.Cookie("uid")
		if cookie == nil || cookie.Value == "" {
			return
		}
		uidInt, _ := strconv.Atoi(cookie.Value)
		uid = int64(uidInt)
	}
	return
}

// TrimStrSpace trim string space
func TrimStrSpace(v string) string {
	return strings.TrimSpace(v)
}

// CTimeStr current time string
func CTimeStr() (cTime string) {
	return time.Unix(time.Now().Unix(), 0).Format("2006-01-02 15:04:05")
}

// CTimeDay current day time
func CTimeDay() (cTime string) {
	return time.Unix(time.Now().Unix(), 0).Format("2006-01-02")
}

func PartMap(s string, sep string) map[string]struct{} {
	parts := strings.Split(s, sep)
	m := make(map[string]struct{}, len(parts))
	for _, part := range parts {
		m[part] = struct{}{}
	}
	return m
}
