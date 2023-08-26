package common

import (
	"context"
	"fmt"

	"go-common/library/log"
	"go-common/library/sync/errgroup.v2"
	"go-common/library/xstr"

	arcgrpc "git.bilibili.co/bapis/bapis-go/archive/service"
	"go-gateway/app/app-svr/app-feed/admin/model/common"

	taGrpcModel "git.bilibili.co/bapis/bapis-go/community/interface/tag"
)

const (
	_archiveOpen = 0
	_archiveFix  = -6
	_archiveCron = -40
)

var (
	_archiveAuditStatus = map[int]string{
		-1:   "待审",
		-2:   "打回",
		-3:   "网警锁定",
		-4:   "锁定",
		-5:   "锁定",
		-7:   "暂缓审核",
		-8:   "补档审核",
		-9:   "转码中",
		-10:  "延迟发布",
		-11:  "视频源待修",
		-12:  "上传转储失败",
		-13:  "允许评论待审",
		-14:  "临时回收站",
		-15:  "分发中",
		-16:  "转码失败",
		-20:  "稿件创建未提交",
		-30:  "稿件创建已提交",
		-40:  "用户定时发布", // 2021M7: 放开限制 https://www.tapd.bilibili.co/20064511/prong/stories/view/1120064511002272747
		-100: "UP主删除",
	}
)

// ArchiveTypeGrpc .
func (s *Service) ArchiveTypeGrpc() (p map[int32]*arcgrpc.Tp, err error) {
	var (
		typeReply *arcgrpc.TypesReply
	)
	p = make(map[int32]*arcgrpc.Tp)
	if typeReply, err = s.arcClient.Types(ctx, &arcgrpc.NoArgRequest{}); err != nil {
		log.Error("arcRPC loadType Error %v", err)
		return
	}
	res := typeReply.Types
	if len(res) == 0 {
		log.Error("arcRPC loadType Empty")
		return
	}
	for _, value := range res {
		if value.Pid != 0 {
			p[value.ID] = value
		}
	}
	return
}

// ArcTypeString with param ids return partition name string
func (s *Service) ArcTypeString(ids string) (idstr string, err error) {
	idsint64, err := xstr.SplitInts(ids)
	if err != nil {
		return
	}
	idsint32 := []int32{}
	for _, v := range idsint64 {
		idsint32 = append(idsint32, int32(v))
	}
	for _, id := range idsint32 {
		if v, ok := s.ArcType[id]; ok {
			if idstr != "" {
				idstr = idstr + "," + v.Name
			} else {
				idstr = v.Name
			}
		} else {
			if idstr != "" {
				idstr = idstr + "," + "id为" + fmt.Sprintf("%d", id) + "没有找到数据"
			} else {
				idstr = "id为" + fmt.Sprintf("%d", id) + "没有找到数据"
			}
		}
	}
	return
}

// ArcTagString with param ids return tag name string
func (s *Service) ArcTagString(ids string) (res map[int64]string, err error) {
	var (
		tags map[int64]*taGrpcModel.Tag
	)
	idsint64, err := xstr.SplitInts(ids)
	if err != nil {
		return
	}
	tags, err = s.TagGrpc(idsint64)
	if err != nil {
		return
	}
	res = make(map[int64]string, len(tags))
	for _, id := range idsint64 {
		if v, ok := tags[id]; ok {
			res[id] = v.Name
		} else {
			res[id] = "id为" + fmt.Sprintf("%d", id) + "没有找到数据"
		}
	}
	return
}

// ArchivesType .
func (s *Service) ArchivesType(ids []int32) (p map[int32]*arcgrpc.Tp, err error) {
	p = make(map[int32]*arcgrpc.Tp)
	for _, id := range ids {
		if v, ok := s.ArcType[id]; ok {
			p[id] = v
		}
	}
	return
}

// Archives .
func (s *Service) Archives(ids []int64) (archives *arcgrpc.ArcsReply, err error) {
	if archives, err = s.arcClient.Arcs(ctx, &arcgrpc.ArcsRequest{Aids: ids}); err != nil {
		log.Error("common.Archives error %v", err)
		return
	}
	if archives == nil || len(archives.Arcs) == 0 {
		err = fmt.Errorf("错误类型（无效ID %v）", ids)
		return
	}
	return
}

// TagGrpc .
func (s *Service) TagGrpc(ids []int64) (tags map[int64]*taGrpcModel.Tag, err error) {
	var (
		reply *taGrpcModel.TagsReply
	)
	arg := &taGrpcModel.TagsReq{
		Mid:  0,
		Tids: ids,
	}
	if reply, err = s.tagClient.Tags(ctx, arg); err != nil {
		return
	}
	if reply == nil || len(reply.Tags) == 0 {
		err = fmt.Errorf("参数错误，ID为%q的tag找不到", ids)
		return
	}
	tags = reply.Tags
	return
}

// SearchArchiveAudit .
func (s *Service) SearchArchiveAudit(ctx context.Context, id int64) (*common.Archive, error) {
	res, err := s.arcDao.ArchiveAudit(ctx, id)
	if err != nil {
		return nil, err
	}
	if res.State >= _archiveOpen || res.State == _archiveFix || res.State == _archiveCron {
		return res, nil
	}
	v, ok := _archiveAuditStatus[res.State]
	if ok {
		return nil, fmt.Errorf("稿件（%d）状态为%s，无法提交，如需提交，请联系审核，校验或修改稿件状态", id, v)
	}
	return nil, fmt.Errorf("稿件（%d）状态为%d，无法提交，如需提交，请联系审核，校验或修改稿件状态", id, res.State)
}

// SearchArchiveCheck .
func (s *Service) SearchArchiveCheck(ctx context.Context, id int64) (*common.Archive, error) {
	var (
		arc *common.Archive
		err error
	)
	eg := errgroup.WithContext(ctx)
	eg.Go(func(ctx context.Context) error {
		return s.arcDao.ArchiveSearchBan2(ctx, id)
	})
	eg.Go(func(ctx context.Context) error {
		tmpArc, e := s.SearchArchiveAudit(ctx, id)
		if e != nil {
			return e
		}
		arc = tmpArc
		return nil
	})
	err = eg.Wait()
	if err != nil {
		return nil, err
	}
	return arc, nil
}
