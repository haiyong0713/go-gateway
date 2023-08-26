package service

import (
	"context"
	"encoding/json"

	"go-common/library/log"
	pb "go-gateway/app/app-svr/app-gw/management/api"
	"go-gateway/app/app-svr/app-gw/management/internal/model"
)

func (s *HttpService) ListLog(ctx context.Context, req *pb.ListLogReq) (*pb.ListLogReply, error) {
	modelList, err := s.dao.ListLog(ctx, req.Node, req.Gateway, req.Object, req.Pn, req.Ps)
	if err != nil {
		return nil, err
	}
	lists := []*pb.LogItem{}
	content := new(model.Extra)
	for _, item := range modelList.Result {
		if err = json.Unmarshal([]byte(item.ExtraData), &content); err != nil {
			log.Error("Failed to unmarshal extra_data: %+v", err)
			continue
		}
		pblist := &pb.LogItem{
			JobId:    item.Str3,
			Ctime:    content.Ctime,
			Mtime:    content.Mtime,
			State:    content.Result,
			Level:    content.Level,
			Category: content.Category,
			ExtraContent: pb.ExtraContent{
				Detail: content.Detail,
			},
			Sponsor: pb.Sponsor{
				Uid:   item.Uid,
				Uname: item.Uname,
			},
			Entity: pb.Entity{
				Gateway:    item.Str1,
				ObjectType: item.Type,
				Action:     item.Action,
				Identifier: item.Str2,
				Env:        item.Str4,
				Zone:       item.Str5,
			},
		}
		lists = append(lists, pblist)
	}
	reply := &pb.ListLogReply{
		Lists: lists,
		Pages: pb.Page{
			Num:   modelList.Page.Num,
			Size_: modelList.Page.Size,
			Total: modelList.Page.Total,
		},
	}
	return reply, nil
}
