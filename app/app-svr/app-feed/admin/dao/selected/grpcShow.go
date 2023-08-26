package selected

import (
	"context"

	v1 "go-gateway/app/app-svr/app-show/interface/api"

	"github.com/pkg/errors"
)

// RefreshSeries def.
func (d *Dao) RefreshSeries(c context.Context, sType string) (err error) {
	_, err = d.showGrpc.RefreshSeriesList(c, &v1.RefreshSeriesListReq{Type: sType})
	if err != nil {
		return errors.WithMessagef(err, "Dao RefreshSeries showGrpc.RefreshSeriesList")
	}
	_, err = d.showGrpcSH004.RefreshSeriesList(c, &v1.RefreshSeriesListReq{Type: sType})
	if err != nil {
		return errors.WithMessagef(err, "Dao RefreshSeries showGrpcSH004.RefreshSeriesList")
	}
	return
}

// RefreshSingleSerie def
func (d *Dao) RefreshSingleSerie(c context.Context, sType string, number int64) (err error) {
	_, err = d.showGrpc.RefreshSerie(c, &v1.RefreshSerieReq{Type: sType, Number: number})
	if err != nil {
		return errors.WithMessagef(err, "Dao RefreshSingleSerie showGrpc.RefreshSerie")
	}
	_, err = d.showGrpcSH004.RefreshSerie(c, &v1.RefreshSerieReq{Type: sType, Number: number})
	if err != nil {
		return errors.WithMessagef(err, "Dao RefreshSingleSerie showGrpcSH004.RefreshSerie")
	}
	return
}
