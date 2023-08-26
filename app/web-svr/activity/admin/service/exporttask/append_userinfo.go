package exporttask

import (
	"context"
	"git.bilibili.co/bapis/bapis-go/account/service"
	"go-common/library/log"
	"strconv"
)

var accountBatch = 50

type AccountAppend struct {
	Field string
}

func (a *AccountAppend) Append(c context.Context, taskRet []map[string]string) []map[string]string {
	mapMid := make(map[int64]*api.Card)
	index := make([]int, 0, accountBatch)
	req := &api.MidsReq{
		Mids: make([]int64, 0, accountBatch),
	}
	for i, one := range taskRet {
		uid, _ := strconv.ParseInt(one[a.Field], 10, 64)
		if info, ok := mapMid[uid]; ok {
			taskRet[i]["nickname"] = info.Name
		} else {
			if uid > 0 {
				req.Mids = append(req.Mids, uid)
				index = append(index, i)
			}
			if len(index) >= accountBatch {
				reply, err := accClient.Cards3(c, req)
				if err != nil {
					log.Errorc(c, "AccountAppend accClient.Cards3(c, %v) error(%v)", req.Mids, err)
					continue
				}
				for _, j := range index {
					uid, _ := strconv.ParseInt(taskRet[j][a.Field], 10, 64)
					if info, ok := reply.Cards[uid]; ok {
						mapMid[uid] = info
						taskRet[j]["nickname"] = info.Name
					}
				}
				index = make([]int, 0, accountBatch)
				req = &api.MidsReq{
					Mids: make([]int64, 0, accountBatch),
				}
			}
		}
	}
	if len(index) > 0 {
		reply, err := accClient.Cards3(c, req)
		if err != nil {
			log.Errorc(c, "AccountAppend accClient.Cards3(c, %v) error(%v)", req.Mids, err)
		} else {
			for _, j := range index {
				uid, _ := strconv.ParseInt(taskRet[j][a.Field], 10, 64)
				if info, ok := reply.Cards[uid]; ok {
					taskRet[j]["nickname"] = info.Name
				}
			}
		}
	}
	return taskRet
}
