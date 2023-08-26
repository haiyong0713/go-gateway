package tribe

import (
	"context"
	"strconv"
	"strings"

	"go-gateway/app/app-svr/fawkes/service/conf"
	fkdao "go-gateway/app/app-svr/fawkes/service/dao/fawkes"
	ossdao "go-gateway/app/app-svr/fawkes/service/dao/oss"
	gitSvr "go-gateway/app/app-svr/fawkes/service/service/gitlab"
	log "go-gateway/app/app-svr/fawkes/service/tools/logger"
)

// Service struct.
type Service struct {
	c      *conf.Config
	fkDao  *fkdao.Dao
	gitSvr *gitSvr.Service
	ossDao *ossdao.Dao
}

// New new service.
func New(c *conf.Config) (s *Service) {
	s = &Service{
		c:      c,
		fkDao:  fkdao.New(c),
		gitSvr: gitSvr.New(c),
		ossDao: ossdao.New(c),
	}
	return
}

// Ping dao.
func (s *Service) Ping(c context.Context) (err error) {
	if err = s.fkDao.Ping(c); err != nil {
		log.Error("s.dao error(%v)", err)
	}
	return
}

// Close dao.
func (s *Service) Close() {
	s.fkDao.Close()
}

// MakeGitPath to make jobid to gitPath
func MakeGitPath(gitPath string, glJobID int64) (jobURL string) {
	var (
		groupName, projectName string
	)
	if len(gitPath) > 0 {
		if strings.HasPrefix(gitPath, "git@") {
			// git@git.bilibili.co:studio/android/bilibiliStudio.git
			pathComps := strings.Split(gitPath, ":")
			projectNameComp := strings.Split(pathComps[len(pathComps)-1], ".git")[0]
			projectNameComps := strings.Split(projectNameComp, "/")
			groupName = projectNameComps[0]
			i := 0
			projectName = strings.Join(append(projectNameComps[:i], projectNameComps[i+1:]...), "/")
		} else {
			pathComps := strings.Split(gitPath, "/")
			projectName = strings.Split(pathComps[len(pathComps)-1], ".git")[0]
			groupName = pathComps[len(pathComps)-2]
		}
		return conf.Conf.Gitlab.Host + "/" + groupName + "/" + projectName + "/-/jobs/" + strconv.FormatInt(glJobID, 10)
	}
	return
}
