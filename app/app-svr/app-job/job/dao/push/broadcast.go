package push

import (
	"context"
	"encoding/json"
	farm "go-farm"
	"strconv"

	"go-common/library/log"
	api "go-gateway/app/app-svr/app-job/job/api"
	"go-gateway/app/app-svr/app-job/job/model/resource"

	bcType "git.bilibili.co/bapis/bapis-go/push/service/broadcast/type"
	bcApi "git.bilibili.co/bapis/bapis-go/push/service/broadcast/v2"

	"github.com/gogo/protobuf/types"
)

func (d *Dao) BroadcastEntry(c context.Context, entryMsg *resource.EntryMsg, pp *resource.PlatLimit, ma, dv string) error {
	var filters []*bcType.LabelFilter
	filters = append(filters, &bcType.LabelFilter{
		Key:       _mobiApp,
		Pattern:   ma,
		MatchKind: bcType.LabelFilter_EQUAL,
	})
	if dv != "" {
		filters = append(filters, &bcType.LabelFilter{
			Key:       _device,
			Pattern:   dv,
			MatchKind: bcType.LabelFilter_EQUAL,
		})
	}
	mk, negate, ok := matchKind(pp.Conditions)
	if ok {
		filters = append(filters, &bcType.LabelFilter{
			Key:       _build,
			Pattern:   strconv.FormatInt(int64(pp.Build), 10),
			MatchKind: mk,
			Negate:    negate, // 是否取反
		})
	}
	pushOpt := &bcType.PushOptions{
		LabelFilters: filters,
	}
	online := &api.TopOnline{
		Icon:     entryMsg.StaticIcon,
		Uri:      entryMsg.Url,
		UniqueId: entryMsg.StateName,
		Interval: d.c.Custom.TopActivityInterval,
		Animate: &api.Animate{
			Svg:  entryMsg.DynamicIcon,
			Loop: entryMsg.LoopCnt,
		},
		Type: _topActivityMng,
		Name: entryMsg.EntryName,
	}
	bs, _ := json.Marshal(online)
	hash := strconv.FormatUint(farm.Hash64(bs), 10)
	em := &api.TopActivityReply{
		Online: online,
		Hash:   hash,
	}
	body, err := types.MarshalAny(em)
	if err != nil {
		return err
	}
	msg := &bcType.Message{
		TargetPath: _bcResourceTargetPath,
		Body:       body,
	}
	req := &bcApi.PushAllReq{
		Opts:  pushOpt,
		Msg:   msg,
		Token: d.c.Broadcast.ResourceToken,
	}
	res, err := d.bcClient.PushAll(c, req)
	if err != nil {
		log.Error("日志告警 broadcast push err(%+v)", err)
		return err
	}
	log.Warn("日志告警 broadcast push success(%d) msg(%+v) req(%+v)", res.GetMsgId(), em, req)
	return nil
}

func matchKind(conditions string) (bcType.LabelFilter_MatchKind, bool, bool) {
	switch conditions {
	case "gt":
		return bcType.LabelFilter_GT, false, true
	case "lt":
		return bcType.LabelFilter_LT, false, true
	case "eq":
		return bcType.LabelFilter_EQUAL, false, true
	case "ne":
		return bcType.LabelFilter_EQUAL, true, true
	default:
		return bcType.LabelFilter_GT, false, false
	}
}
