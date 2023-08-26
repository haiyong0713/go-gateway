package mod

import (
	"context"
	"time"

	"go-gateway/app/app-svr/fawkes/service/conf"
	"go-gateway/app/app-svr/fawkes/service/model/mod"
	"go-gateway/app/app-svr/fawkes/service/model/pcdn"
	log "go-gateway/app/app-svr/fawkes/service/tools/logger"
	"go-gateway/app/app-svr/fawkes/service/tools/utils"
)

func (s *Service) PushPcdn(ctx context.Context, appKey string, file *mod.File, patches []*mod.Patch) (err error) {
	if !utils.Contain(appKey, conf.Conf.Mod.PCDN.AppKey) {
		log.Infoc(ctx, "%v该应用暂不支持PCDN", appKey)
		return
	}
	var files []*pcdn.Files

	files = append(files, convertFile(file))
	for _, p := range patches {
		files = append(files, convertPatch(p))
	}

	if err = s.fkDao.BatchAddPcdnFile(ctx, files); err != nil {
		log.Errorc(ctx, "PushPcdn %v", err)
		return
	}
	return
}

func convertPatch(p *mod.Patch) *pcdn.Files {
	return &pcdn.Files{
		Rid:       p.Name + "_" + p.Md5,
		Url:       p.URL,
		Md5:       p.Md5,
		Size:      p.Size,
		Business:  string(pcdn.MOD),
		VersionId: pcdn.VersionId(time.Now()),
	}
}

func convertFile(f *mod.File) *pcdn.Files {
	return &pcdn.Files{
		Rid:       f.Name + "_" + f.Md5,
		Url:       f.URL,
		Md5:       f.Md5,
		Size:      f.Size,
		Business:  string(pcdn.MOD),
		VersionId: pcdn.VersionId(time.Now()),
	}
}
