package archive

import (
	"fmt"

	"go-gateway/app/app-svr/app-feed/interface/conf"
	arcgrpc "go-gateway/app/app-svr/archive/service/api"
)

// Dao is archive dao.
type Dao struct {
	//grpc
	rpcClient arcgrpc.ArchiveClient
}

// New new a archive dao.
func New(c *conf.Config) (d *Dao) {
	d = &Dao{}
	//grpc
	var err error
	if d.rpcClient, err = arcgrpc.NewClient(c.ArchiveGRPC); err != nil {
		panic(fmt.Sprintf("rpcClient NewClientt error (%+v)", err))
	}
	return
}
