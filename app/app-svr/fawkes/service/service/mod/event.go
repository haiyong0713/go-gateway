package mod

import (
	"context"
	"fmt"

	"go-gateway/app/app-svr/fawkes/service/model/mod"
	log "go-gateway/app/app-svr/fawkes/service/tools/logger"
)

const (
	// VersionRelease mod版本发布
	VersionRelease = "inner.mod.version.release"

	// VersionConfigChange mod配置变化
	VersionConfigChange = "inner.mod.version.config.change"

	// VersionGrayChange mod灰度变化
	VersionGrayChange = "inner.mod.version.gray.change"

	// VersionPush mod推送正式
	VersionPush = "inner.mod.version.push"
)

type VersionPushArgs struct {
	Ctx       context.Context
	VersionId int64
	UserName  string
}

type VersionReleaseArgs struct {
	Ctx       context.Context
	VersionId int64
	UserName  string
}

type VersionConfigChangeArgs struct {
	Ctx      context.Context
	old      *mod.Config
	new      *mod.Config
	UserName string
}

type VersionGrayChangeArgs struct {
	Ctx      context.Context
	old      *mod.Gray
	new      *mod.Gray
	UserName string
}

func (s *Service) EventInit() {
	if err := s.event.SubscribeAsync(VersionRelease, s.versionReleaseAction, false); err != nil {
		panic(err)
	}
	if err := s.event.SubscribeAsync(VersionConfigChange, s.versionConfigChangeAction, false); err != nil {
		panic(err)
	}
	if err := s.event.SubscribeAsync(VersionGrayChange, s.versionGrayChangeAction, false); err != nil {
		panic(err)
	}
	if err := s.event.SubscribeAsync(VersionPush, s.versionPushAction, false); err != nil {
		panic(err)
	}
}

func (s *Service) versionReleaseAction(args *VersionReleaseArgs) {
	c := args.Ctx
	versionId := args.VersionId
	user := args.UserName
	log.Infoc(c, "user:%v release version: %v", user, versionId)
	v, err := s.fkDao.ModVersionByID(c, versionId)
	if err != nil {
		log.Errorc(c, "mod_moni_action error: %v", err)
		return
	}
	estimate, err := s.ModReleaseTrafficEstimate(c, versionId, user)
	if v.Env == mod.EnvProd {
		if err = s.Alert(c, estimate, mod.Release); err != nil {
			log.Errorc(c, fmt.Sprintf("%v", err))
			return
		}
	}
}

func (s *Service) versionConfigChangeAction(args *VersionConfigChangeArgs) {
	var (
		err    error
		c      = args.Ctx
		oldCfg = args.old
		newCfg = args.new
		user   = args.UserName
	)
	version, err := s.fkDao.ModVersionByID(c, newCfg.VersionID)
	if err != nil {
		log.Errorc(c, "mod_moni_action error: %v", err)
		return
	}
	if version.Env == mod.EnvTest || !version.Released {
		log.Warnc(c, "测试环境或未发布资源跳过检测")
		return
	}
	newEstimate, err := s.ModReleaseTrafficEstimate(c, newCfg.VersionID, args.UserName)
	if err != nil {
		log.Errorc(c, "mod_moni_action error: %v", err)
		return
	}
	oldEstimate := newEstimate
	oldEstimate.Config = oldCfg
	log.Infoc(c, "user:%v config change old->new: %+v->%+v", user, oldCfg, newCfg)
	oldDetail, newDetail := getDetail(oldEstimate), getDetail(newEstimate)
	log.Infoc(c, "config change  cost %v  %v", newDetail.Cost, oldDetail.Cost)
	if newDetail.Cost > oldDetail.Cost {
		log.Infoc(c, "config change  true")
		// 变更后会造成成本上升
		if err = s.Alert(c, newEstimate, mod.ConfigChange); err != nil {
			log.Errorc(c, fmt.Sprintf("%v", err))
			return
		}
	}
}

func (s *Service) versionPushAction(args *VersionPushArgs) {
	var (
		err     error
		ctx     = args.Ctx
		prodMod *mod.Mod
	)
	log.Infoc(ctx, "version push action %v", args)
	// 资源同步到pcdn files
	if prodMod, err = s.GetProdMod(ctx, args.VersionId); err != nil {
		log.Errorc(ctx, "mod event error: %v", err)
		return
	}
	if err = s.PushPcdn(ctx, prodMod.Pool.AppKey, prodMod.File, prodMod.Patches); err != nil {
		log.Errorc(ctx, "mod event error: %v", err)
		return
	}
}

func (s *Service) versionGrayChangeAction(args *VersionGrayChangeArgs) {

}
