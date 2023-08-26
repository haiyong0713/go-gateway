package prediction

import (
	"context"
	"fmt"

	premdl "go-gateway/app/web-svr/activity/interface/model/prediction"
)

func predictionKey(id int64) string {
	return fmt.Sprintf("go_pre_l_%d", id)
}

func predItemKey(id int64) string {
	return fmt.Sprintf("go_p_it_l_%d", id)
}

//go:generate kratos tool btsgen
type _bts interface {
	// bts:-sync=true
	Predictions(c context.Context, ids []int64) (map[int64]*premdl.Prediction, error)
	// bts:-sync=true
	PredItems(c context.Context, ids []int64) (map[int64]*premdl.PredictionItem, error)
}

//go:generate kratos tool mcgen
type _mc interface {
	// mc: -key=predictionKey
	CachePredictions(c context.Context, ids []int64) (res map[int64]*premdl.Prediction, err error)
	// mc: -key=predictionKey -expire=d.mcPerpetualExpire -encode=pb
	AddCachePredictions(c context.Context, val map[int64]*premdl.Prediction) error
	// mc: -key=predictionKey
	DelCachePredictions(c context.Context, ids []int64) error
	// mc: -key=predItemKey
	CachePredItems(c context.Context, ids []int64) (res map[int64]*premdl.PredictionItem, err error)
	// mc: -key=predItemKey -expire=d.mcPerpetualExpire -encode=pb
	AddCachePredItems(c context.Context, val map[int64]*premdl.PredictionItem) error
	// mc: -key=predItemKey
	DelCachePredItems(c context.Context, ids []int64) error
}
