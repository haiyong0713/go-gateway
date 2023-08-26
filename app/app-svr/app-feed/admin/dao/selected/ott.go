package selected

import (
	"context"
	"github.com/pkg/errors"
	"go-gateway/app/app-svr/app-feed/admin/model/selected"
	"strconv"
)

//nolint:bilirailguncheck
func (d *Dao) PubOTT(c context.Context, msg *selected.OTTSeriesMsg) (err error) {
	err = d.ottSeriesPub.Send(c, strconv.FormatInt(msg.Number, 10), msg)
	if err != nil {
		err = errors.WithMessage(err, "dao PubOTT Send")
		return
	}
	return
}
