package archive

import (
	"go-gateway/app/app-svr/app-dynamic/interface/conf"
	archivegrpc "go-gateway/app/app-svr/archive/service/api"
)

type Dao struct {
	c *conf.Config
	// grpc
	archiveGRPC archivegrpc.ArchiveClient
}

func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		c: c,
	}
	var err error
	if d.archiveGRPC, err = archivegrpc.NewClient(d.c.ArchiveGRPC); err != nil {
		panic(err)
	}
	return
}
