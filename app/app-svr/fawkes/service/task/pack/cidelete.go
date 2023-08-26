package pack

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"go-common/library/railgun"

	"go-gateway/app/app-svr/fawkes/service/conf"
	"go-gateway/app/app-svr/fawkes/service/dao/fawkes"
	cimdl "go-gateway/app/app-svr/fawkes/service/model/ci"
	taskmdl "go-gateway/app/app-svr/fawkes/service/model/task"
	log "go-gateway/app/app-svr/fawkes/service/tools/logger"
	"go-gateway/app/app-svr/fawkes/service/tools/utils"
)

type CIDeleteTask struct {
	conf  *conf.Config
	fkDao *fawkes.Dao
	name  string
}

func (t *CIDeleteTask) HandlerFunc(ctx context.Context) railgun.MsgPolicy {
	var (
		statistic *taskmdl.Statistics
		err       error
	)
	emptyDate := time.Date(1, 1, 1, 0, 0, 0, 0, time.UTC)
	CiDeleteConfig := conf.Conf.Task.NasClean.CIDelete
	startTime := CiDeleteConfig.Start
	endTime := CiDeleteConfig.End
	if emptyDate == startTime && emptyDate == endTime {
		// 全空 时间为六个月前
		nowTime := time.Now()
		startTime = nowTime.AddDate(0, -CiDeleteConfig.Persistence, -2)
		endTime = nowTime.AddDate(0, -CiDeleteConfig.Persistence, -1)
	}
	if statistic, err = t.Clean(ctx, startTime, endTime, CiDeleteConfig.PackType, CiDeleteConfig.AppKey); err != nil {
		log.Errorc(ctx, "CleanRailgun occur error: %+v", err)
		return railgun.MsgPolicyFailure
	}
	log.Infoc(ctx, "CleanRailgun statistics clean startTime[%s] endTime[%s], statistics[%+v]", startTime.String(), endTime.String(), statistic)
	return railgun.MsgPolicyNormal
}

func (t *CIDeleteTask) Clean(ctx context.Context, tStart, tEnd time.Time, pkgTypes []int64, appKey string) (re *taskmdl.Statistics, err error) {
	var (
		list      []*cimdl.BuildPack
		keys      []*taskmdl.BuildKey
		deleteRes *taskmdl.DeleteResult
		sum       int64 // 需要删除的总条数
	)
	if list, err = t.fkDao.CINasList(ctx, appKey, pkgTypes, tStart, tEnd); err != nil {
		log.Errorc(ctx, "query pack error(%v)", err)
		return
	}
	for _, v := range list {
		keys = append(keys, &taskmdl.BuildKey{BuildId: v.BuildID, AppKey: v.AppKey})
	}
	if sum = int64(len(list)); sum == 0 {
		re = &taskmdl.Statistics{
			BatchSum: 0,
		}
		return
	}
	if deleteRes, err = t.deleteCINas(ctx, keys); err != nil {
		log.Errorc(ctx, "delete nas error: %v", err)
		return
	}
	if deleteRes == nil {
		log.Warnc(ctx, "delete expired pack response is nil")
		return
	}
	tt, _ := strconv.ParseFloat(fmt.Sprintf("%.4f", float64(sum-int64(len(deleteRes.FailedId)))/float64(sum)), 64)
	return &taskmdl.Statistics{
		BatchSum:     sum,
		RateStr:      fmt.Sprintf("%.2f", tt*100) + "%",
		Rate:         tt,
		FailList:     deleteRes.FailedId,
		DeleteFailed: int64(len(deleteRes.FailedId)),
		UpdateFail:   sum - int64(len(deleteRes.FailedId)) - deleteRes.AffectedRows,
	}, err
}

// DeleteCINas 删除CI文件并且更新过期状态
func (t *CIDeleteTask) deleteCINas(ctx context.Context, keys []*taskmdl.BuildKey) (res *taskmdl.DeleteResult, err error) {
	var (
		fileDeleted []int64
		bp          *cimdl.BuildPack
		fileBytes   int64
	)
	res = &taskmdl.DeleteResult{}
	now := time.Now()
	for _, deleteKey := range keys {
		if deleteKey.AppKey == "" || deleteKey.BuildId == 0 {
			continue
		}
		if bp, err = t.fkDao.BuildPack(ctx, deleteKey.AppKey, deleteKey.BuildId); err != nil {
			log.Errorc(ctx, "BuildPack error(%v)", err)
			continue
		}
		if now.AddDate(0, -conf.Conf.Task.NasClean.CIDelete.Persistence, 0).Before(time.Unix(bp.CTime, 0)) {
			log.Errorc(ctx, "buildId: %d CTime: %s 在当前时间%d个月之内创建的包不可删除。", deleteKey.BuildId, time.Unix(bp.CTime, 0), conf.Conf.Task.NasClean.CIDelete.Persistence)
			continue
		}
		filePath := filepath.Join(conf.Conf.LocalPath.LocalDir, "pack", deleteKey.AppKey, strconv.FormatInt(bp.GitlabJobID, 10))
		if !utils.FileExists(filePath) {
			log.Errorc(ctx, "appKey[%s], buildID[%d] pkgType[%d], filePath: %s doesn't exist.", deleteKey.AppKey, deleteKey.BuildId, bp.PkgType, filePath)
			continue
		}
		res.NeedDelete = append(res.NeedDelete, deleteKey.BuildId)
		if fileBytes, err = utils.DirSizeB(filePath); err != nil {
			log.Errorc(ctx, "appKey[%s], buildID[%d] pkgType[%d], filePath: %s calc file size error.", deleteKey.AppKey, deleteKey.BuildId, bp.PkgType, filePath)
			continue
		}
		_metricCleanNasSize.Add(float64(fileBytes), deleteKey.AppKey, "CI")
		_metricCleanNasCount.Inc(deleteKey.AppKey, "CI")
		if err = os.RemoveAll(filePath); err != nil {
			res.FailedId = append(res.FailedId, deleteKey.BuildId)
			log.Errorc(ctx, "appKey[%s], buildID[%d] pkgType[%d], delete file: %s FAILED! err: %+v", deleteKey.AppKey, deleteKey.BuildId, bp.PkgType, filePath, err)
		} else {
			log.Infoc(ctx, "appKey[%s], buildID[%d] pkgType[%d], delete file: %s SUCCESS!", deleteKey.AppKey, deleteKey.BuildId, bp.PkgType, filePath)
			fileDeleted = append(fileDeleted, deleteKey.BuildId)
		}
	}
	if len(fileDeleted) == 0 {
		if len(res.NeedDelete) != 0 {
			log.Errorc(ctx, "以下buildId: %v需要删除但是删除失败", res.NeedDelete)
		}
		return
	}
	log.Infoc(ctx, "删除的包 build_pack id：%v", fileDeleted)
	if res.AffectedRows, err = t.fkDao.UpdateCIExpiredStatus(ctx, fileDeleted); err != nil {
		log.Errorc(ctx, "UpdateCIExpiredStatus File Fail(%v)", err)
	}
	if res.AffectedRows != int64(len(fileDeleted)) {
		log.Errorc(ctx, "更新包的删除状态与总数不符，需要更新%d条，实际更新%d条", len(fileDeleted), res.AffectedRows)
	}
	log.Infoc(ctx, "statistical result: %+v", &res)
	return
}

func (t *CIDeleteTask) TaskName() string {
	return t.name
}

func NewCIDeleteTask(c *conf.Config, fkDao *fawkes.Dao, name string) *CIDeleteTask {
	r := &CIDeleteTask{
		conf:  c,
		fkDao: fkDao,
		name:  name,
	}
	return r
}
