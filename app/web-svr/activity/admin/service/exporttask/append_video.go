package exporttask

import (
	"context"
	"fmt"
	"go-common/library/log"
	"go-gateway/app/app-svr/archive/service/api"
	"strconv"
)

var videoBatch = 50

type VideoAppend struct {
	Field string
}

func (v *VideoAppend) Append(c context.Context, taskRet []map[string]string) []map[string]string {
	num := len(taskRet)
	for i := 0; i < num; i += videoBatch {
		var tmpRet []map[string]string
		if i+videoBatch <= num {
			tmpRet = taskRet[i : i+videoBatch]
		} else {
			tmpRet = taskRet[i:num]
		}
		req := &api.ArcsRequest{}
		tmpNum := len(tmpRet)
		req.Aids = make([]int64, 0, tmpNum)
		for _, one := range tmpRet {
			vid, _ := strconv.ParseInt(one[v.Field], 10, 64)
			if vid > 0 {
				req.Aids = append(req.Aids, vid)
			}
		}
		if len(req.Aids) == 0 {
			log.Errorc(c, "VideoAppend len(req.Aids) == 0")
			continue
		}
		arcsReply, err := arcClient.Arcs(c, req)
		if err != nil {
			log.Errorc(c, "VideoAppend arcClient.Arcs(c, %v) error(%v)", req.Aids, err)
			continue
		}
		for j := i; j < i+tmpNum; j++ {
			vid, _ := strconv.ParseInt(taskRet[j][v.Field], 10, 64)
			if arc, ok := arcsReply.Arcs[vid]; ok {
				taskRet[j]["arc_like"] = fmt.Sprint(arc.Stat.Like)
				taskRet[j]["view"] = fmt.Sprint(arc.Stat.View)
				taskRet[j]["coin"] = fmt.Sprint(arc.Stat.Coin)
				taskRet[j]["fav"] = fmt.Sprint(arc.Stat.Fav)
				taskRet[j]["share"] = fmt.Sprint(arc.Stat.Share)
				taskRet[j]["title"] = fmt.Sprint(arc.Title)
				taskRet[j]["type_name"] = arc.TypeName
				taskRet[j]["type_id"] = fmt.Sprint(arc.TypeID)
			}
		}
	}
	return taskRet
}
