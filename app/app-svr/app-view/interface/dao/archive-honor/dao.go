package archive_honor

import (
	"context"
	"fmt"

	"go-gateway/app/app-svr/app-view/interface/conf"
	ahApi "go-gateway/app/app-svr/archive-honor/service/api"

	"github.com/pkg/errors"
)

// Dao is archive-honor dao
type Dao struct {
	// grpc
	ahClient ahApi.ArchiveHonorClient
}

// New is
func New(c *conf.Config) (d *Dao) {
	d = &Dao{}
	var err error
	if d.ahClient, err = ahApi.NewClient(c.ArchiveHonorClient); err != nil {
		panic(fmt.Sprintf("archive honor Client not found err(%v)", err))
	}
	return
}

// Honors is
func (d *Dao) Honors(c context.Context, aid, build int64, mobiApp, device string) ([]*ahApi.Honor, error) {
	req := &ahApi.HonorRequest{Aid: aid, Build: build, MobiApp: mobiApp, Device: device}
	rep, err := d.ahClient.Honor(c, req)
	if err != nil {
		err = errors.Wrapf(err, "%v", req)
		return nil, err
	}
	return rep.GetHonor(), nil
}

func (d *Dao) BatchHonors(ctx context.Context, req *ahApi.HonorsRequest) (*ahApi.HonorsReply, error) {
	return d.ahClient.Honors(ctx, req)
}
