package selected

import (
	"context"
	"strconv"

	"go-gateway/app/app-svr/app-feed/admin/model/selected"

	"github.com/pkg/errors"
)

//nolint:bilirailguncheck
func (d *Dao) PubArchiveHonor(c context.Context, archiveHonor *selected.ArchiveHonor) (err error) {
	err = d.archiveHonorPub.Send(c, strconv.FormatInt(archiveHonor.Aid, 10), archiveHonor)
	if err != nil {
		err = errors.WithMessage(err, "dao PubArchiveHonor Send")
		return
	}
	return
}
