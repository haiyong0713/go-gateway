package unicom

import (
	"context"
	"strconv"

	log "go-common/library/log"
	"go-gateway/app/app-svr/app-wall/interface/model/unicom"
)

func (s *Service) addUserBindState(u *unicom.UserBindInfo) {
	select {
	case s.userBindCh <- u:
	default:
		log.Warn("add user bind state buffer is full")
	}
}

// nolint:bilirailguncheck
func (s *Service) userbindConsumer() {
	for {
		i, ok := <-s.userBindCh
		if !ok {
			return
		}
		var (
			err error
		)
		switch v := i.(type) {
		case *unicom.UserBindInfo:
			if err = s.userbindPub.Send(context.TODO(), strconv.FormatInt(v.MID, 10), v); err != nil {
				log.Error("s.userbindSub.Send error(%v)", err)
				continue
			}
			log.Info("s.userbindSub.Send(%+v) success", v)
		}
	}
}
