package job

import (
	"context"

	jobmdl "go-gateway/app/web-svr/web-show/interface/model/job"
)

// Jobs get job infos
func (s *Service) Jobs(c context.Context) (js []*jobmdl.Job) {
	js = s.cache
	return
}
