package dynamic

import (
	"context"

	"go-common/library/log"

	infocV2 "go-common/library/log/infoc.v2"

	dynmdl "go-gateway/app/app-svr/app-dynamic/interface/model/dynamic"
)

// infoc
func (s *Service) infoc(i interface{}) {
	switch v := i.(type) {
	case dynmdl.SVideoInfoc:
		payload := infocV2.NewLogStream(s.c.Infoc.SvideoLogID, v.AID, v.UpID, v.Buvid, v.MID, v.FromSpmid, v.Follow, v.Like, v.CardType, v.CardIndex, v.Offset, v.OType, v.OID)
		if err := s.svideoInfoc.Info(context.Background(), payload); err != nil {
			log.Error("infocproc s.svideoInfoc.Info err(%+v)", err)
		}
	default:
		log.Warn("infocproc can't process the type")
	}
}
