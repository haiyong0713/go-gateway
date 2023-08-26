package bvav

import (
	"go-common/library/log"
	"go-gateway/app/web-svr/esports/interface/model"
	"go-gateway/pkg/idsafe/bvid"
)

// AvToBv .
func AvToBv(contests []*model.Contest) {
	var (
		err error
	)
	for _, v := range contests {
		if v.Aid <= 0 {
			log.Info("listContest aid(%v) equal zero", v.Aid)
			continue
		}
		if v.Bvid, err = bvid.AvToBv(v.Aid); err != nil {
			log.Error("listContest AvToBv(%v)error (%v)", v.Aid, err)
			err = nil
		}
		if v.Collection == 0 {
			continue
		}
		if v.CollectionBvid, err = bvid.AvToBv(v.Collection); err != nil {
			log.Error("listContest AvToBv(%v)error (%v)", v.Aid, err)
			err = nil
		}
	}
}
