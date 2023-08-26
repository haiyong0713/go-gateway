package unicom

import (
	"context"
	"strconv"
	"strings"
	"time"

	log "go-common/library/log"

	"go-common/library/log/infoc.v2"
)

// nolint:gomnd
func (s *Service) InfocIP(c context.Context, mid int64, buvid, mobiApp, device, platform string, build int, localInfo, ip, pip, localOperator, ipOperator string) {
	localInfos := strings.Split(localInfo, ";")
	var manufacturer, model, release, sdkInt string
	if len(localInfos) == 4 {
		manufacturer, model, release, sdkInt = localInfos[0], localInfos[1], localInfos[2], localInfos[3]
	} else {
		log.Error("InfocIP localInfo(%s) len not equal 4", localInfo)
	}
	operatorCompare := localOperator == ipOperator
	event := infoc.NewLogStreamV(s.c.IPLogID,
		log.String(strconv.FormatInt(mid, 10)),
		log.String(buvid),
		log.Int64(time.Now().Unix()),
		log.String(mobiApp),
		log.String(device),
		log.String(platform),
		log.String(strconv.Itoa(build)),
		log.String(manufacturer),
		log.String(model),
		log.String(release),
		log.String(sdkInt),
		log.String(ip),
		log.String(pip),
		log.String(localOperator),
		log.String(ipOperator),
		log.Bool(operatorCompare),
		log.String("x/wall/ip"),
	)
	if err := s.infocV2Log.Info(c, event); err != nil {
		log.Error("InfocIP params(%d,%s,%s,%s,%s,%d,%s,%s,%s,%s,%s) err(%+v)",
			mid, buvid, mobiApp, device, platform, build, localInfo, ip, pip, localOperator, ipOperator, err)
		return
	}
	log.Warn("InfocIP logid(%s) params(%d,%s,%s,%s,%s,%d,%s,%s,%s,%s,%s)",
		s.c.IPLogID, mid, buvid, mobiApp, device, platform, build, localInfo, ip, pip, localOperator, ipOperator)
}
