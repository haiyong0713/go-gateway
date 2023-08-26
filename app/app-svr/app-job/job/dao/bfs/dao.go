package bfs

import (
	"context"

	"go-common/library/database/bfs"
	"go-common/library/log"
	"go-gateway/app/app-svr/app-job/job/conf"
)

// Dao macross dao.
type Dao struct {
	// conf
	c *conf.Config
	// bfs client
	bfsCli *bfs.BFS
}

// New dao.
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		c: c,
		// bfs
		bfsCli: bfs.New(c.BFS),
	}
	return
}

// Upload to bfs.
func (d *Dao) Upload(c context.Context, bucket, dir, fileName, contentType string, file []byte, wmKey string, wmPaddingX, wmPaddingY uint32, wmScale float64, wmPos string, wmtransparency float64) (url string, err error) {
	if url, err = d.bfsCli.Upload(c, &bfs.Request{
		Bucket:         bucket,
		Dir:            dir,
		Filename:       fileName,
		ContentType:    contentType,
		File:           file,
		WMKey:          wmKey,
		WMPaddingX:     wmPaddingX,
		WMPaddingY:     wmPaddingY,
		WMScale:        wmScale,
		WMPos:          wmPos,
		WMTransparency: wmtransparency,
	}); err != nil {
		log.Error("%v", err)
	}
	return
}
