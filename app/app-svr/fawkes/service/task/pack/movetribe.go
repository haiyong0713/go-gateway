package pack

import (
	"context"
	"os"
	"path"
	"strings"

	"go-common/library/railgun"

	"go-gateway/app/app-svr/fawkes/service/conf"
	"go-gateway/app/app-svr/fawkes/service/dao/fawkes"
	appmdl "go-gateway/app/app-svr/fawkes/service/model/app"
	"go-gateway/app/app-svr/fawkes/service/model/tribe"
	log "go-gateway/app/app-svr/fawkes/service/tools/logger"
	"go-gateway/app/app-svr/fawkes/service/tools/utils"
)

type MoveTribeTask struct {
	fkDao *fawkes.Dao
	name  string
}

func (t *MoveTribeTask) HandlerFunc(ctx context.Context) railgun.MsgPolicy {
	var (
		app []*appmdl.APP
		err error
	)
	log.Infoc(ctx, "开始移动tribe包产物")
	mvTask := conf.Conf.Task.MoveTribe
	log.Infoc(ctx, "配置：%+v", mvTask)
	appKey := mvTask.Apps
	if len(appKey) == 0 {
		log.Infoc(ctx, "没有配置app_keys")
		if app, err = t.fkDao.AppAll(ctx); err != nil {
			log.Errorc(ctx, "error:%v", err)
			return railgun.MsgPolicyFailure
		}
		for _, v := range app {
			appKey = append(appKey, v.AppKey)
		}
	}
	log.Infoc(ctx, "本次移动的appKey:%+v", appKey)
	for _, v := range appKey {
		log.Infoc(ctx, "appKey:%v 开始执行", v)
		var count int
		if count, err = t.fkDao.CountTribeBuildPack(ctx, v, 0, 0, 0, 0, 0, 0, "", "", "", 0, ""); err != nil {
			log.Errorc(ctx, "%v", err)
			return railgun.MsgPolicyFailure
		}
		pageNum := (count + mvTask.BatchSize - 1) / mvTask.BatchSize
		log.Infoc(ctx, "将会分%d次执行", min(pageNum, mvTask.Batch))
		// 分页执行
		for i := 0; i < min(pageNum, mvTask.Batch); i++ {
			log.Infoc(ctx, "第%d次执行", i+1)
			var tribePacks []*tribe.BuildPack
			if tribePacks, err = t.fkDao.SelectTribeBuildPackByArg(ctx, v, 0, 0, 0, 0, 0, 0, 0, "", "", "", "", "id", "", int64(mvTask.BatchSize), int64(i+1)); err != nil {
				log.Errorc(ctx, "error %v", err)
				return railgun.MsgPolicyFailure
			}
			// 1. 过滤出需要移动的包
			movePack := needMove(tribePacks)
			var id []int64
			for _, p := range movePack {
				id = append(id, p.Id)
			}
			log.Infoc(ctx, "appKey: %s, 需要移动的tribe包：%v", v, id)
			// 2. 将需要移动的包拷贝到目标路径
			if err = copyPack(movePack); err != nil {
				log.Errorc(ctx, "copy error: %v", err)
				return railgun.MsgPolicyFailure
			}
			log.Infoc(ctx, "拷贝完成")
			// 3. 修改数据库中存储的路径
			if err = t.updatePkgPath(ctx, movePack); err != nil {
				log.Errorc(ctx, "update pkg path error: %v", err)
				return railgun.MsgPolicyFailure
			}
			log.Infoc(ctx, "数据库更新完成")
			// 4. 删除原文件
			if err = deleteOldPack(movePack); err != nil {
				log.Errorc(ctx, "delete pkg error: %v", err)
				return railgun.MsgPolicyFailure
			}
			log.Infoc(ctx, "源文件删除完成")
		}
		log.Infoc(ctx, "appKey:%s 执行结束", v)
	}
	return railgun.MsgPolicyNormal
}

func (t *MoveTribeTask) updatePkgPath(ctx context.Context, pack []*tribe.BuildPack) (err error) {
	mvCfg := conf.Conf.Task.MoveTribe
	for _, v := range pack {
		newPkgPath := mvCfg.NewDir + strings.TrimPrefix(v.PkgPath, mvCfg.OldDir)
		newPkgUrl := mvCfg.NewUrl + strings.TrimPrefix(v.PkgUrl, mvCfg.OldUrl)
		newMappingUrl := mvCfg.NewUrl + strings.TrimPrefix(v.MappingUrl, mvCfg.OldUrl)
		newBbrUrl := mvCfg.NewUrl + strings.TrimPrefix(v.BbrUrl, mvCfg.OldUrl)
		// 更新tribe_build_pack (ci记录)
		if _, err = t.fkDao.UpdateTribeBuildPackPkgInfo(ctx, v.Id, newPkgPath, newPkgUrl, newMappingUrl, newBbrUrl, "", "", "", 0, "", 0, 0, 0); err != nil {
			return
		}
		if v.DidPush == 1 {
			// 如果已经同步过cd 需要同时更新cd表中的包路径
			if _, err = t.fkDao.UpdateTribePackPkgInfo(ctx, v.TribeId, v.GlJobId, newPkgPath, newPkgUrl, newMappingUrl, newBbrUrl); err != nil {
				return
			}
		}
	}
	return
}

func needMove(packs []*tribe.BuildPack) []*tribe.BuildPack {
	var need []*tribe.BuildPack
	mvCfg := conf.Conf.Task.MoveTribe
	for _, v := range packs {
		if len(v.PkgPath) != 0 {
			if strings.HasPrefix(v.PkgPath, mvCfg.OldDir) {
				need = append(need, v)
			}
		}
	}
	return need
}

func copyPack(pack []*tribe.BuildPack) (err error) {
	mvCfg := conf.Conf.Task.MoveTribe
	for _, v := range pack {
		pkgDir := path.Dir(strings.TrimPrefix(v.PkgPath, mvCfg.OldDir))
		if err = utils.CopyDir(path.Dir(v.PkgPath), mvCfg.NewDir+pkgDir); err != nil {
			return err
		}
	}
	return
}

func deleteOldPack(pack []*tribe.BuildPack) (err error) {
	for _, v := range pack {
		if err = os.RemoveAll(path.Dir(v.PkgPath)); err != nil {
			return
		}
	}
	return
}

func (t *MoveTribeTask) TaskName() string {
	return t.name
}

func NewMoveTribeTask(fkDao *fawkes.Dao, name string) *MoveTribeTask {
	t := &MoveTribeTask{
		fkDao: fkDao,
		name:  name,
	}
	return t
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
