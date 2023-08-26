package article

import (
	"context"
	replyAPI "git.bilibili.co/bapis/bapis-go/community/interface/reply"
	"go-common/library/log"
)

//func (d *Dao) Arcs(ctx context.Context, aids []int64) (arcsReply *archiveAPI.ArcsReply, err error) {
//	arcsReply = &archiveAPI.ArcsReply{}
//	if len(aids) <= 0 {
//		return arcsReply, nil
//	}
//	if arcsReply, err = d.arcClient.Arcs(ctx, &archiveAPI.ArcsRequest{Aids: aids}); err != nil {
//		log.Error("d.Arc aids:%+v d.grpc.ArchiveClient.Arcs err:%+v", aids, err)
//		return arcsReply, err
//	}
//	return arcsReply, nil
//}

func (d *Dao) AddReplyOperation(ctx context.Context, req *replyAPI.AddOperationReq) (resp *replyAPI.AddOperationResp, err error) {
	if resp, err = d.replyClient.AddOperation(ctx, req); err != nil {
		log.Error("d.AddReplyOperation req:%+v err:%+v", req, err)
		return &replyAPI.AddOperationResp{}, err
	}
	return resp, nil
}

func (d *Dao) OfflineReplyOperation(ctx context.Context, opid int64) (resp *replyAPI.OfflineOperationResp, err error) {
	req := &replyAPI.OfflineOperationReq{
		Id: opid,
	}
	if resp, err = d.replyClient.OfflineOperation(ctx, req); err != nil {
		log.Error("d.OfflineReplyOperation req:%+v err:%+v", req, err)
		return &replyAPI.OfflineOperationResp{}, err
	}
	return resp, nil
}
