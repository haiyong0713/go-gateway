package service

import (
	"context"

	"go-common/library/ecode"
	"go-common/library/log"

	pb "go-gateway/app/app-svr/resource/service/api/v1"
	"go-gateway/app/app-svr/resource/service/model"
)

func (s *Service) ResourceAllNew(c context.Context, req *pb.NoArgRequest) (res *pb.ResourceAllReply, err error) {
	res = new(pb.ResourceAllReply)
	for _, resTmp := range s.ResourceAll(c) {
		if resTmp == nil {
			continue
		}
		re := &pb.ResourceInfo{
			Id:          int32(resTmp.ID),
			Platform:    int32(resTmp.Platform),
			Name:        resTmp.Name,
			Parent:      int32(resTmp.Parent),
			Counter:     int32(resTmp.Counter),
			Position:    int32(resTmp.Position),
			Rule:        resTmp.Rule,
			Size_:       resTmp.Size,
			Preview:     resTmp.Previce,
			Description: resTmp.Desc,
			Mark:        resTmp.Mark,
			Ctime:       resTmp.CTime,
			Mtime:       resTmp.MTime,
			Level:       resTmp.Level,
			Type:        int32(resTmp.Type),
			IsAd:        int32(resTmp.IsAd),
		}
		for _, asgTmp := range resTmp.Assignments {
			if asgTmp == nil {
				continue
			}
			asg := &pb.Assignment{
				Id:             int32(asgTmp.ID),
				Name:           asgTmp.Name,
				ContractId:     asgTmp.ContractID,
				ResourceId:     int32(asgTmp.ResID),
				Pic:            asgTmp.Pic,
				Litpic:         asgTmp.LitPic,
				Url:            asgTmp.URL,
				Rule:           asgTmp.Rule,
				Weight:         int32(asgTmp.Weight),
				Agency:         asgTmp.Agency,
				Price:          asgTmp.Price,
				State:          int32(asgTmp.State),
				Atype:          int32(asgTmp.Atype),
				Username:       asgTmp.Username,
				PlayerCategory: int32(asgTmp.PlayerCategory),
				Stime:          asgTmp.STime,
				Etime:          asgTmp.ETime,
				Ctime:          asgTmp.CTime,
				Mtime:          asgTmp.MTime,
				ActivityId:     asgTmp.ActivityID,
				ActivityStime:  asgTmp.ActivitySTime,
				ActivityEtime:  asgTmp.ActivityETime,
				Category:       int32(asgTmp.Category),
				SubTitle:       asgTmp.SubTitle,
			}
			re.Assignments = append(re.Assignments, asg)
		}
		res.Resources = append(res.Resources, re)
	}
	return
}

func (s *Service) AssignmentAllNew(c context.Context, req *pb.NoArgRequest) (res *pb.AssignmentAllReply, err error) {
	res = new(pb.AssignmentAllReply)
	for _, asgTmp := range s.AssignmentAll(c) {
		if asgTmp == nil {
			continue
		}
		asg := &pb.Assignment{
			Id:             int32(asgTmp.ID),
			Name:           asgTmp.Name,
			ContractId:     asgTmp.ContractID,
			ResourceId:     int32(asgTmp.ResID),
			Pic:            asgTmp.Pic,
			Litpic:         asgTmp.LitPic,
			Url:            asgTmp.URL,
			Rule:           asgTmp.Rule,
			Weight:         int32(asgTmp.Weight),
			Agency:         asgTmp.Agency,
			Price:          asgTmp.Price,
			State:          int32(asgTmp.State),
			Atype:          int32(asgTmp.Atype),
			Username:       asgTmp.Username,
			PlayerCategory: int32(asgTmp.PlayerCategory),
			Stime:          asgTmp.STime,
			Etime:          asgTmp.ETime,
			Ctime:          asgTmp.CTime,
			Mtime:          asgTmp.MTime,
			ActivityId:     asgTmp.ActivityID,
			ActivityStime:  asgTmp.ActivitySTime,
			ActivityEtime:  asgTmp.ActivityETime,
			Category:       int32(asgTmp.Category),
			SubTitle:       asgTmp.SubTitle,
		}
		res.Assignments = append(res.Assignments, asg)
	}
	return
}

func (s *Service) DefBannerNew(c context.Context, req *pb.NoArgRequest) (res *pb.DefBannerReply, err error) {
	res = new(pb.DefBannerReply)
	defbanner := s.DefBanner(c)
	if defbanner == nil {
		return
	}
	res.DefBanner = &pb.Assignment{
		Id:             int32(defbanner.ID),
		Name:           defbanner.Name,
		ContractId:     defbanner.ContractID,
		ResourceId:     int32(defbanner.ResID),
		Pic:            defbanner.Pic,
		Litpic:         defbanner.LitPic,
		Url:            defbanner.URL,
		Rule:           defbanner.Rule,
		Weight:         int32(defbanner.Weight),
		Agency:         defbanner.Agency,
		Price:          defbanner.Price,
		State:          int32(defbanner.State),
		Atype:          int32(defbanner.Atype),
		Username:       defbanner.Username,
		PlayerCategory: int32(defbanner.PlayerCategory),
		Stime:          defbanner.STime,
		Etime:          defbanner.ETime,
		Ctime:          defbanner.CTime,
		Mtime:          defbanner.MTime,
		ActivityId:     defbanner.ActivityID,
		ActivityStime:  defbanner.ActivitySTime,
		ActivityEtime:  defbanner.ActivityETime,
		Category:       int32(defbanner.Category),
		SubTitle:       defbanner.SubTitle,
	}
	return
}

func (s *Service) ResourceNew(c context.Context, req *pb.ResourceRequest) (res *pb.ResourceReply, err error) {
	res = new(pb.ResourceReply)
	resTmp := s.Resource(c, int(req.ResID))
	if resTmp == nil {
		return
	}
	res.Resource = &pb.ResourceInfo{}
	res.Resource.Id = int32(resTmp.ID)
	res.Resource.Platform = int32(resTmp.Platform)
	res.Resource.Name = resTmp.Name
	res.Resource.Parent = int32(resTmp.Parent)
	res.Resource.Counter = int32(resTmp.Counter)
	res.Resource.Position = int32(resTmp.Position)
	res.Resource.Rule = resTmp.Rule
	res.Resource.Size_ = resTmp.Size
	res.Resource.Preview = resTmp.Previce
	res.Resource.Description = resTmp.Desc
	res.Resource.Mark = resTmp.Mark
	res.Resource.Ctime = resTmp.CTime
	res.Resource.Mtime = resTmp.MTime
	res.Resource.Level = resTmp.Level
	res.Resource.Type = int32(resTmp.Type)
	res.Resource.IsAd = int32(resTmp.IsAd)
	for _, asgTmp := range resTmp.Assignments {
		if asgTmp == nil {
			continue
		}
		asg := &pb.Assignment{
			Id:             int32(asgTmp.ID),
			Name:           asgTmp.Name,
			ContractId:     asgTmp.ContractID,
			ResourceId:     int32(asgTmp.ResID),
			Pic:            asgTmp.Pic,
			Litpic:         asgTmp.LitPic,
			Url:            asgTmp.URL,
			Rule:           asgTmp.Rule,
			Weight:         int32(asgTmp.Weight),
			Agency:         asgTmp.Agency,
			Price:          asgTmp.Price,
			State:          int32(asgTmp.State),
			Atype:          int32(asgTmp.Atype),
			Username:       asgTmp.Username,
			PlayerCategory: int32(asgTmp.PlayerCategory),
			Stime:          asgTmp.STime,
			Etime:          asgTmp.ETime,
			Ctime:          asgTmp.CTime,
			Mtime:          asgTmp.MTime,
			ActivityId:     asgTmp.ActivityID,
			ActivityStime:  asgTmp.ActivitySTime,
			ActivityEtime:  asgTmp.ActivityETime,
			Category:       int32(asgTmp.Category),
			SubTitle:       asgTmp.SubTitle,
		}
		res.Resource.Assignments = append(res.Resource.Assignments, asg)
	}
	return
}

func (s *Service) ResourcesNew(c context.Context, req *pb.ResourcesRequest) (res *pb.ResourcesReply, err error) {
	res = new(pb.ResourcesReply)
	resIds := make([]int, 0)
	for _, v := range req.ResIDs {
		resIds = append(resIds, int(v))
	}
	for resID, resTmp := range s.Resources(c, resIds) {
		if resTmp == nil {
			continue
		}
		re := &pb.ResourceInfo{
			Id:          int32(resTmp.ID),
			Platform:    int32(resTmp.Platform),
			Name:        resTmp.Name,
			Parent:      int32(resTmp.Parent),
			Counter:     int32(resTmp.Counter),
			Position:    int32(resTmp.Position),
			Rule:        resTmp.Rule,
			Size_:       resTmp.Size,
			Preview:     resTmp.Previce,
			Description: resTmp.Desc,
			Mark:        resTmp.Mark,
			Ctime:       resTmp.CTime,
			Mtime:       resTmp.MTime,
			Level:       resTmp.Level,
			Type:        int32(resTmp.Type),
			IsAd:        int32(resTmp.IsAd),
		}
		for _, asgTmp := range resTmp.Assignments {
			if asgTmp == nil {
				continue
			}
			asg := &pb.Assignment{
				Id:             int32(asgTmp.ID),
				Name:           asgTmp.Name,
				ContractId:     asgTmp.ContractID,
				ResourceId:     int32(asgTmp.ResID),
				Pic:            asgTmp.Pic,
				Litpic:         asgTmp.LitPic,
				Url:            asgTmp.URL,
				Rule:           asgTmp.Rule,
				Weight:         int32(asgTmp.Weight),
				Agency:         asgTmp.Agency,
				Price:          asgTmp.Price,
				State:          int32(asgTmp.State),
				Atype:          int32(asgTmp.Atype),
				Username:       asgTmp.Username,
				PlayerCategory: int32(asgTmp.PlayerCategory),
				Stime:          asgTmp.STime,
				Etime:          asgTmp.ETime,
				Ctime:          asgTmp.CTime,
				Mtime:          asgTmp.MTime,
				ActivityId:     asgTmp.ActivityID,
				ActivityStime:  asgTmp.ActivitySTime,
				ActivityEtime:  asgTmp.ActivityETime,
				Category:       int32(asgTmp.Category),
				SubTitle:       asgTmp.SubTitle,
			}
			re.Assignments = append(re.Assignments, asg)
		}
		if res.Resources == nil {
			res.Resources = make(map[int32]*pb.ResourceInfo)
		}
		res.Resources[int32(resID)] = re
	}
	return
}

func (s *Service) PasterAPPNew(c context.Context, req *pb.PasterAPPRequest) (res *pb.PasterAPPReply, err error) {
	res = new(pb.PasterAPPReply)
	var resTmp *model.Paster
	if resTmp, err = s.PasterAPP(c, int8(req.Platform), int8(req.AdType), req.Aid, req.TypeID, req.Buvid); err != nil {
		log.Error("%v", err)
		return
	}
	if resTmp == nil {
		return
	}
	res.Aid = resTmp.AID
	res.Cid = resTmp.CID
	res.Duration = resTmp.Duration
	res.Type = int32(resTmp.Type)
	res.AllowJump = int32(resTmp.AllowJump)
	res.Url = resTmp.URL
	return
}

func (s *Service) IndexIconNew(c context.Context, req *pb.NoArgRequest) (res *pb.IndexIconReply, err error) {
	res = new(pb.IndexIconReply)
	for iconType, icons := range s.IndexIcon(c) {
		var re *pb.IndexIconItem
		for _, icon := range icons {
			if icon == nil {
				continue
			}
			re = &pb.IndexIconItem{
				Id:       icon.ID,
				Type:     int32(icon.Type),
				Title:    icon.Title,
				Links:    icon.Links,
				Icon:     icon.Icon,
				Weight:   int32(icon.Weight),
				UserName: icon.UserName,
				Sttime:   icon.StTime,
				Endtime:  icon.EndTime,
				Deltime:  icon.DelTime,
				Ctime:    icon.CTime,
				Mtime:    icon.MTime,
			}
			if res.IndexIcon == nil {
				res.IndexIcon = make(map[string]*pb.IndexIcon)
			}
			var (
				idxIcon *pb.IndexIcon
				ok      bool
			)
			if idxIcon, ok = res.IndexIcon[iconType]; !ok {
				idxIcon = &pb.IndexIcon{}
				res.IndexIcon[iconType] = idxIcon
			}
			idxIcon.IndexIconItem = append(idxIcon.IndexIconItem, re)
		}
	}
	return
}

func (s *Service) PlayerIconNew(c context.Context, req *pb.NoArgRequest) (res *pb.PlayerIconReply, err error) {
	res = new(pb.PlayerIconReply)
	var resTmp *model.PlayerIcon
	if resTmp, err = s.PlayerIcon(c, 0, []int64{}, 0, 0, false, false); err != nil {
		log.Error("%v", err)
		return
	}
	if resTmp == nil {
		return
	}
	res.Url1 = resTmp.URL1
	res.Hash1 = resTmp.Hash1
	res.Url2 = resTmp.URL2
	res.Hash2 = resTmp.Hash2
	res.Ctime = resTmp.CTime
	res.Type = int32(resTmp.Type)
	res.TypeValue = resTmp.TypeValue
	res.Mtime = resTmp.MTime
	return
}
func (s *Service) PlayerIcon2NewV2(c context.Context, req *pb.PlayerIconRequest) (*pb.PlayerIconV2Reply, error) {
	rly, err := s.PlayerIcon2New(c, req)
	if err != nil {
		if err == ecode.NothingFound {
			return &pb.PlayerIconV2Reply{}, nil
		}
		return nil, err
	}
	return &pb.PlayerIconV2Reply{Item: rly}, nil
}

func (s *Service) PlayerIcon2New(c context.Context, req *pb.PlayerIconRequest) (res *pb.PlayerIconReply, err error) {
	res = new(pb.PlayerIconReply)
	var resTmp *model.PlayerIcon
	isUnder604 := model.IsUnderAndroid604(req.MobiApp, req.Build)
	if resTmp, err = s.PlayerIcon(c, req.Aid, req.TagIDs, req.TypeID, req.Mid, req.ShowPlayIcon, isUnder604); err != nil {
		log.Error("%v", err)
		return
	}
	if resTmp == nil {
		return
	}
	res.Url1 = resTmp.URL1
	res.Hash1 = resTmp.Hash1
	res.Url2 = resTmp.URL2
	res.Hash2 = resTmp.Hash2
	res.Ctime = resTmp.CTime
	res.Type = int32(resTmp.Type)
	res.TypeValue = resTmp.TypeValue
	res.Mtime = resTmp.MTime
	res.DragRightPng = resTmp.DragRightPng
	res.MiddlePng = resTmp.MiddlePng
	res.DragLeftPng = resTmp.DragLeftPng
	res.DragData = resTmp.DragData
	res.NodragData = resTmp.NoDragData
	return
}

func (s *Service) CmtboxNew(c context.Context, req *pb.CmtboxRequest) (res *pb.CmtboxReply, err error) {
	res = new(pb.CmtboxReply)
	var resTmp *model.Cmtbox
	if resTmp, err = s.Cmtbox(c, req.Id); err != nil {
		log.Error("%v", err)
		return
	}
	if resTmp == nil {
		return
	}
	res.Id = resTmp.ID
	res.LoadCid = resTmp.LoadCID
	res.Server = resTmp.Server
	res.Port = resTmp.Port
	res.SizeFactor = resTmp.SizeFactor
	res.SpeedFactor = resTmp.SpeedFactor
	res.MaxOnscreen = resTmp.MaxOnscreen
	res.Style = resTmp.Style
	res.StyleParam = resTmp.StyleParam
	res.TopMargin = resTmp.TopMargin
	res.State = resTmp.State
	res.RenqiVisible = resTmp.RenqiVisible
	res.RenqiFontsize = resTmp.RenqiFontsize
	res.RenqiFmt = resTmp.RenqiFmt
	res.RenqiOffset = resTmp.RenqiOffset
	res.RenqiColor = resTmp.RenqiColor
	res.Ctime = resTmp.CTime
	res.Mtime = resTmp.MTime
	return
}

func (s *Service) SideBarsNew(c context.Context, req *pb.NoArgRequest) (res *pb.SideBarsReply, err error) {
	res = new(pb.SideBarsReply)
	resTmp := s.SideBars(c)
	if resTmp == nil {
		return
	}
	for _, sideBars := range resTmp.SideBar {
		res.SideBar = append(res.SideBar, &pb.SideBar{
			Id:           sideBars.ID,
			Tip:          int32(sideBars.Tip),
			Rank:         int32(sideBars.Rank),
			Logo:         sideBars.Logo,
			LogoWhite:    sideBars.LogoWhite,
			Name:         sideBars.Name,
			Param:        sideBars.Param,
			Module:       int32(sideBars.Module),
			Plat:         int32(sideBars.Plat),
			Build:        int32(sideBars.Build),
			Conditions:   sideBars.Conditions,
			OnlineTime:   sideBars.OnlineTime,
			NeedLogin:    int32(sideBars.NeedLogin),
			WhiteUrl:     sideBars.WhiteURL,
			Menu:         int32(sideBars.Menu),
			LogoSelected: sideBars.LogoSelected,
			TabId:        sideBars.TabID,
			RedDotUrl:    sideBars.Red,
			Language:     sideBars.Language,
			GlobalRedDot: int32(sideBars.GlobalRed),
			RedDotLimit:  sideBars.RedLimit,
			Animate:      sideBars.Animate,
			WhiteUrlShow: sideBars.WhiteURLShow,
		})
	}
	for sid, sideBarLimits := range resTmp.Limit {
		for _, sideBarLmit := range sideBarLimits {
			if sideBarLmit == nil {
				continue
			}
			if res.Limit == nil {
				res.Limit = make(map[int64]*pb.SideBarLimit)
			}
			var (
				rl *pb.SideBarLimit
				ok bool
			)
			if rl, ok = res.Limit[sid]; !ok {
				rl = &pb.SideBarLimit{}
				res.Limit[sid] = rl
			}
			rl.SideBarLimitItem = append(rl.SideBarLimitItem, &pb.SideBarLimitItem{Id: sideBarLmit.ID, Build: int32(sideBarLmit.Build), Condition: sideBarLmit.Condition})
		}
	}
	return
}

func (s *Service) AbTestNew(c context.Context, req *pb.AbTestRequest) (res *pb.AbTestReply, err error) {
	res = new(pb.AbTestReply)
	for name, resTmp := range s.AbTest(c, req.Groups, req.Ip) {
		if resTmp == nil {
			continue
		}
		if res.Abtest == nil {
			res.Abtest = make(map[string]*pb.AbTest)
		}
		res.Abtest[name] = &pb.AbTest{
			GroupId:     resTmp.ID,
			GroupName:   resTmp.Name,
			FlowPercent: resTmp.Threshold,
			ParamValues: resTmp.ParamValues,
			Utime:       resTmp.UTime,
		}
	}
	return
}

func (s *Service) PasterCIDNew(c context.Context, req *pb.NoArgRequest) (res *pb.PasterCIDReply, err error) {
	res = new(pb.PasterCIDReply)
	var resTmp map[int64]int64
	if resTmp, err = s.PasterCID(c); err != nil {
		log.Error("%v", err)
		return
	}
	res.Paster = resTmp
	return
}
