package dynamicV2

import (
	"fmt"
	"strconv"
	"strings"

	"go-common/library/log"

	dyngrpc "git.bilibili.co/bapis/bapis-go/dynamic/service/feed"
)

func (list *DynListRes) FromDynLight(new *dyngrpc.GeneralHistoryRsp, uid int64) {
	list.HistoryOffset = new.HistoryOffset
	list.HasMore = new.HasMore
	var logs []string
	for _, item := range new.Dyns {
		if item.Type == 0 {
			continue
		}
		logs = append(logs, fmt.Sprintf("dynid(%v) type(%v) rid(%v)", item.DynId, item.Type, item.Rid))
		dynTmp := &Dynamic{}
		dynTmp.FromDynamic(item)
		list.Dynamics = append(list.Dynamics, dynTmp)
	}
	log.Warn("FromUnloginLight(new) origin mid(%d) list(%v)", uid, strings.Join(logs, "; "))
}

func (list *DynListRes) FromDynUnLoginLight(new *dyngrpc.UnloginLightRsp, uid int64) {
	var logs []string
	for _, item := range new.Dyns {
		if item.Type == 0 {
			continue
		}
		logs = append(logs, fmt.Sprintf("dynid(%v) type(%v) rid(%v)", item.DynId, item.Type, item.Rid))
		dynTmp := &Dynamic{}
		dynTmp.FromDynamic(item)
		list.Dynamics = append(list.Dynamics, dynTmp)
		list.HistoryOffset = strconv.FormatInt(dynTmp.DynamicID, 10)
	}
	log.Warn("FromUnloginLight(new) origin mid(%d) list(%v)", uid, strings.Join(logs, "; "))
}
