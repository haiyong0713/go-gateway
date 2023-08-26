package service

import (
	"context"
	"time"

	"go-common/library/ecode"
	"go-common/library/log"
	xtime "go-common/library/time"

	pb "go-gateway/app/app-svr/resource/service/api/v1"
	pb2 "go-gateway/app/app-svr/resource/service/api/v2"

	"go-gateway/app/app-svr/resource/service/model"
)

// Relate Relate card grpc
func (s *Service) Relate(ctx context.Context, req *pb.RelateRequest) (special *pb.SpecialReply, err error) {
	special = &pb.SpecialReply{}
	if req == nil || req.Id == 0 || req.MobiApp == "" || req.Build == 0 {
		err = ecode.RequestErr
		return
	}
	var (
		relateID int64
		ok       bool
		relate   *model.Relate
		versions []*model.Version
	)
	//判断seasonID 是否有配置的相关推荐卡片
	if relateID, ok = s.relatePgcMapCache[req.Id]; !ok {
		err = ecode.NothingFound
		//log.Warn("gRpc.Relate relatePgcMapCache error,req.id(%v)", req.Id)
		return
	}
	//取出seasonID 对应的 相关推荐的卡片数据
	if relate, ok = s.relateCache[relateID]; !ok {
		err = ecode.NothingFound
		//log.Warn("gRpc.Relate relateCache error,req.id(%v),relateID (%v)", req.Id, relateID)
		return
	}
	p := model.Plat(req.MobiApp, req.Device)
	//判断APP是否存在
	if versions, ok = relate.Versions[p]; !ok {
		err = ecode.NothingFound
		//log.Warn("gRpc.Relate relate.Versions error,req.id(%v),plat (%v)", req.Id, p)
		return
	}
	//判断APP版本是否存在
	if len(versions) == 0 {
		err = ecode.NothingFound
		//log.Warn("gRpc.Relate versions error,Versions is zero,req.id(%v)", req.Id)
		return
	}
	//判断版本信息是否匹配
	for _, v := range versions {
		if model.InvalidBuild(int(req.Build), v.Build, v.Condition) {
			err = ecode.NothingFound
			//log.Warn("gRpc.Relate InvalidBuild error,req.id(%v),req.Build (%v)", req.Id, req.Build)
			return
		}
	}
	var specialTmp *pb.SpecialReply
	if specialTmp, ok = s.specialCache[relate.Param]; ok && specialTmp != nil {
		*special = *specialTmp
		special.Position = relate.Position
		special.RecReason = relate.RecReason
	} else {
		special = &pb.SpecialReply{}
		//log.Warn("gRpc.Relate specialCache error,req.id(%v),relate.Param (%v)", req.Id, relate.Param)
	}
	return
}

// loadSpecialCache load special card cache
func (s *Service) loadSpecialCache() {
	offset := 0
	tmpCache := make(map[int64]*pb.SpecialReply, len(s.specialCache))
	for {
		special, nextId, err := s.manager.Specials(context.Background(), offset)
		if err != nil {
			log.Error("日志告警 s.manager.Specials err(%+v)", err)
			return
		}
		for k, v := range special {
			tmpCache[k] = v
		}
		//nolint:gomnd
		if len(special) < 1000 {
			log.Info("loaded specialCache size(%d)", len(tmpCache))
			break
		}
		offset = nextId
		time.Sleep(10 * time.Millisecond)
	}
	s.specialCache = tmpCache
}

func (s *Service) Special(ctx context.Context, req *pb.NoArgRequest) (res *pb.SpecialCardReply, err error) {
	list, err := s.cardDao.SpecialCard(ctx)
	if err != nil {
		log.Error("%+v", err)
		return
	}
	res = &pb.SpecialCardReply{List: list}
	return
}

// 加载特殊卡缓存到内存
func (s *Service) LoadSpecialCardCache() {
	var (
		tmpCache = make(map[int64]*pb2.AppSpecialCard)
		now      = xtime.Time(time.Now().AddDate(-2, 0, 0).Unix())
		offset   int64
		pageSize = 5000
	)

	log.Warn("service.LoadSpecialCardCache Start")
	for {
		specials, nextId, err := s.manager.GetSpecialCard(context.Background(), now, offset, pageSize)
		if err != nil {
			log.Error("service.LoadSpecialCardCache GetSpecialCard now(%+v) offset(%d) size(%d) err(%+v)", now, offset, pageSize, err)
			return
		}
		for k, v := range specials {
			tmpCache[k] = v
		}
		if len(specials) < pageSize {
			break
		}
		offset = nextId
		time.Sleep(20 * time.Millisecond)
	}
	s.specailCardMap = tmpCache
	log.Warn("service.LoadSpecialCardCache success")
}

func (s *Service) GetSpecialCardById(c context.Context, id int64) (specialCard *pb2.AppSpecialCard, err error) {
	if id <= 0 {
		return nil, ecode.Error(ecode.RequestErr, "特殊卡ID错误！")
	}

	// 从内容中获取
	specialCard, ok := s.specailCardMap[id]
	if ok && specialCard != nil {
		return
	}
	// redis缓存中获取
	specialCard, _ = s.manager.GetSpecialFromCache(c, id)
	if specialCard != nil {
		return
	}
	// 从db中查询
	specialCard, err = s.manager.GetSpecialCardById(c, id)
	if err != nil || specialCard == nil {
		log.Error("service.GetSpecialCardById GetSpecialCardById id(%d) err(%+v)", id, err)
		return nil, ecode.Error(ecode.NothingFound, "特殊卡不存在！")
	}
	// 保存到缓存中
	//nolint:errcheck,biligowordcheck
	go s.manager.SetSpecial2Cache(context.Background(), specialCard)

	return
}
