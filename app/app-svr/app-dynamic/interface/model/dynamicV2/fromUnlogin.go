package dynamicV2

import (
	"fmt"
	"strconv"
	"strings"

	"go-common/library/log"

	dyngrpc "git.bilibili.co/bapis/bapis-go/dynamic/service/feed"
)

func (list *DynListRes) FromUnLoginFeed(unlogin *dyngrpc.UnLoginFeedRsp, uid int64) {
	list.HasMore = unlogin.HasMore
	var logs []string
	for _, item := range unlogin.Briefs {
		if item == nil || item.Type == 0 {
			log.Warn("FromUnLoginFeed miss FromUnLoginFeed mid %v, item %+v", uid, item)
			continue
		}
		if item.Type == 1 && item.Origin == nil {
			log.Warn("FromUnLoginFeed miss forward origin nil mid %v, item %+v", uid, item)
			continue
		}
		logs = append(logs, fmt.Sprintf("dynid(%v) type(%v) rid(%v)", item.DynId, item.Type, item.Rid))
		dynTmp := &Dynamic{}
		dynTmp.FromDynamic(item)
		list.Dynamics = append(list.Dynamics, dynTmp)
		list.HistoryOffset = strconv.FormatInt(dynTmp.DynamicID, 10)
	}
	log.Warn("FromUnLoginFeed(new) origin mid(%d) list(%v)", uid, strings.Join(logs, "; "))
}
