package helper

import (
	"context"
	"strconv"
	"strings"

	"go-common/library/log"
	"go-gateway/pkg/idsafe/bvid"
)

func BvidsToAid(ctx context.Context, bvids []string) (avIds map[string]int64, err error) {
	avIds = make(map[string]int64, len(bvids))
	for _, id := range bvids {
		var avid int64
		if strings.HasPrefix(id, "BV1") {
			if avid, err = bvid.BvToAv(id); err != nil || avid == 0 {
				log.Errorc(ctx, "helper bvid.BvToAv() id(%s) error(%+v)", id, err)
				return
			}
		} else {
			if id == "" {
				continue
			}
			if avid, err = strconv.ParseInt(id, 10, 64); err != nil || avid == 0 {
				log.Errorc(ctx, "helper strconv.ParseInt() id(%s) error(%+v)", id, err)
				return
			}
		}
		avIds[id] = avid
	}
	return
}
