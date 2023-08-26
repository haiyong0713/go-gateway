package exporttask

import (
	"context"
	"fmt"
	"go-gateway/app/web-svr/activity/admin/component/boss"
	"io"
	"time"
)

func (s *Service) saveBoss(c context.Context, filename string, reader io.Reader) (string, error) {
	filename = fmt.Sprintf("exportdata/%s/%s", time.Now().Format("20060102"), filename)
	return boss.Client.UploadObject(c, boss.Bucket, filename, reader)
}
