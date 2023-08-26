package util

import (
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/web-svr/space/admin/model"
	"strconv"
	"strings"
	"time"

	report "go-common/library/queue/databus/actionlog"
)

// AddLogs add action logs
func AddLogs(logtype int, uname string, uid int64, oid int64, action string, obj interface{}) (err error) {
	err = report.Manager(&report.ManagerInfo{
		Uname:    uname,
		UID:      uid,
		Business: model.UserTabLog,
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

func UserInfo(c *bm.Context) (username string, uid int64) {
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

func RemoveRepByLoop(slc []int64) []int64 {
	result := []int64{} // 存放结果
	for i := range slc {
		flag := true
		for j := range result {
			if slc[i] == result[j] {
				flag = false // 存在重复元素，标识为false
				break
			}
		}
		if flag { // 标识为false，不添加进结果
			result = append(result, slc[i])
		}
	}
	return result
}

func RemoveRepByMap(slc []int64) []int64 {
	result := []int64{}
	tempMap := map[int64]byte{} // 存放不重复主键
	for _, e := range slc {
		l := len(tempMap)
		tempMap[e] = 0
		if len(tempMap) != l { // 加入map后，map长度变化，则元素不重复
			result = append(result, e)
		}
	}
	return result
}

const loopSliceLength = 1024

func RemoveRep(slc []int64) []int64 {
	if len(slc) < loopSliceLength {
		// 切片长度小于1024的时候，循环来过滤
		return RemoveRepByLoop(slc)
	}
	// 大于的时候，通过map来过滤
	return RemoveRepByMap(slc)
}

// ParamsFilter 多参数过滤
func ParamsFilterTo64(paramStr string) (list []int64) {
	if len(paramStr) > 0 {
		strs := strings.Split(paramStr, ",")
		for _, str := range strs {

			if tmp, err := strconv.ParseInt(str, 10, 64); err == nil {
				list = append(list, tmp)
			}

		}
	}
	return
}

func ParamsFilter(paramStr string) (list []int) {
	if len(paramStr) > 0 {
		strs := strings.Split(paramStr, ",")
		for _, str := range strs {

			if tmp, err := strconv.Atoi(str); err == nil {
				list = append(list, tmp)
			}

		}
	}
	return
}
