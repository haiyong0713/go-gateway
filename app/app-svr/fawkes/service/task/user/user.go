package user

import (
	"context"
	"time"

	"go-common/library/railgun"

	"go-gateway/app/app-svr/fawkes/service/conf"
	"go-gateway/app/app-svr/fawkes/service/dao/fawkes"
	appmdl "go-gateway/app/app-svr/fawkes/service/model/app"
	log "go-gateway/app/app-svr/fawkes/service/tools/logger"
)

type ReloadTask struct {
	conf  *conf.Config
	fkDao *fawkes.Dao
	name  string
}

func NewReloadTask(c *conf.Config, fkDao *fawkes.Dao, name string) *ReloadTask {
	r := &ReloadTask{
		conf:  c,
		fkDao: fkDao,
		name:  name,
	}
	return r
}

func (r *ReloadTask) TaskName() string {
	return r.name
}

func (r *ReloadTask) HandlerFunc(ctx context.Context) railgun.MsgPolicy {
	var (
		wxUsers []*appmdl.User
		token   string
		err     error
	)
	log.Warnc(ctx, "UserUpdateTask start at %v", time.Now())
	if token, err = r.fkDao.GetWechatToken(ctx, conf.Conf.WXNotify.CorpSecret); err != nil {
		log.Errorc(ctx, "get token error: %v", err)
		return railgun.MsgPolicyFailure
	}
	if wxUsers, err = r.fkDao.GetWXUsersByDepartmentId(ctx, conf.Conf.WXNotify.DepartmentIDs, token); err != nil {
		log.Errorc(ctx, "get users error: %v", err)
		return railgun.MsgPolicyFailure
	}
	log.Warnc(ctx, "users len: %v", len(wxUsers))
	if len(wxUsers) > 0 {
		log.Warnc(ctx, "user : %v", wxUsers[0])
	}
	usersGroup := splitSlice(wxUsers, 1000)
	for _, u := range usersGroup {
		if err = r.fkDao.BatchSetUser(ctx, u); err != nil {
			log.Errorc(ctx, "add user error: %v", err)
			return railgun.MsgPolicyFailure
		}
	}
	log.Warnc(ctx, "UserUpdateTask finished at %v", time.Now())
	return railgun.MsgPolicyNormal
}

func splitSlice(input []*appmdl.User, num int64) [][]*appmdl.User {
	max := int64(len(input))
	// 判断数组大小是否小于等于指定分割大小的值，是则把原数组放入二维数组返回
	if max <= num {
		return [][]*appmdl.User{input}
	}
	// 获取应该数组分割为多少份
	var quantity int64
	if max%num == 0 {
		quantity = max / num
	} else {
		quantity = (max / num) + 1
	}
	var segments = make([][]*appmdl.User, 0)
	// 声明分割数组的截止下标
	var start, end, i int64
	for i = 1; i <= quantity; i++ {
		end = i * num
		if i != quantity {
			segments = append(segments, input[start:end])
		} else {
			segments = append(segments, input[start:])
		}
		start = i * num
	}
	return segments
}
