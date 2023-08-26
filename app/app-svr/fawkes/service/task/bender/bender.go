package bender

import (
	"context"
	"time"

	"go-common/library/railgun"

	"go-gateway/app/app-svr/fawkes/service/conf"
	"go-gateway/app/app-svr/fawkes/service/dao/fawkes"
	"go-gateway/app/app-svr/fawkes/service/model/bender"
	"go-gateway/app/app-svr/fawkes/service/model/pcdn"
	log "go-gateway/app/app-svr/fawkes/service/tools/logger"
)

type Task struct {
	conf  *conf.Config
	fkDao *fawkes.Dao
	name  string
}

func NewTask(c *conf.Config, fkDao *fawkes.Dao, name string) *Task {
	r := &Task{
		conf:  c,
		fkDao: fkDao,
		name:  name,
	}
	return r
}

func (r *Task) TaskName() string {
	return r.name
}

func (r *Task) HandlerFunc(ctx context.Context) railgun.MsgPolicy {
	var (
		err      error
		ridSet   *map[string]bool
		res      *bender.ResourceData
		newRes   []*bender.Item
		newFiles []*pcdn.Files
	)
	log.Infoc(ctx, "bender task start at %v", time.Now())
	if res, err = r.fkDao.BenderTopResource(ctx); err != nil {
		log.Errorc(ctx, "%v", err)
		return railgun.MsgPolicyFailure
	}
	if ridSet, err = r.fkDao.PcdnRidAll(ctx); err != nil {
		log.Errorc(ctx, "%v", err)
		return railgun.MsgPolicyFailure
	}
	for _, r := range res.Resources {
		if _, ok := (*ridSet)[r.Key]; !ok {
			newRes = append(newRes, r)
		}
	}
	log.Infoc(ctx, "新增资源数量：%v", len(newRes))
	if len(newRes) > 0 {
		log.Infoc(ctx, "new resource : %v", newRes[0])
	}
	for _, v := range newRes {
		newFiles = append(newFiles, &pcdn.Files{
			Rid:       v.Key,
			Url:       v.Url,
			Md5:       v.Md5,
			Size:      v.Size,
			Business:  string(pcdn.Bender),
			VersionId: pcdn.VersionId(time.Now()),
		})
	}
	filesGroup := splitSlice(newFiles, 1000)
	for _, fg := range filesGroup {
		if err = r.fkDao.BatchAddPcdnFile(ctx, fg); err != nil {
			log.Errorc(ctx, "add file error: %v", err)
			return railgun.MsgPolicyFailure
		}
	}
	log.Infoc(ctx, "bender file sync finished at %v", time.Now())
	return railgun.MsgPolicyNormal
}

func splitSlice(input []*pcdn.Files, num int64) [][]*pcdn.Files {
	max := int64(len(input))
	// 判断数组大小是否小于等于指定分割大小的值，是则把原数组放入二维数组返回
	if max <= num {
		return [][]*pcdn.Files{input}
	}
	// 获取应该数组分割为多少份
	var quantity int64
	if max%num == 0 {
		quantity = max / num
	} else {
		quantity = (max / num) + 1
	}
	var segments = make([][]*pcdn.Files, 0)
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
