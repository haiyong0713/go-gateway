package service

import (
	"context"
	accountGRPC "git.bilibili.co/bapis/bapis-go/account/service"
	"go-common/library/log"
	"go-gateway/app/web-svr/space/admin/model"
)

func (s *Service) GetBannedRcmdList(c context.Context, param *model.UpRcmdBlackListSearchReq) (resp *model.UpRcmdBlackListSearchRep, err error) {
	if result, total, err := s.dao.GetBannedRcmdMids(c, param.Mid, param.Ps, param.Pn); err != nil {
		return resp, err
	} else {
		mids := make([]int64, 0)
		resp = new(model.UpRcmdBlackListSearchRep)
		resp.Page = model.Page{
			Num:   int(param.Pn),
			Size:  int(param.Ps),
			Total: int(total),
		}
		resp.Items = make([]*model.UserInfo, 0)
		if total == 0 {
			return resp, nil
		}
		for _, row := range result {
			mids = append(mids, row.Mid)
			resp.Items = append(resp.Items, &model.UserInfo{Mid: row.Mid, MTime: row.MTime})
		}

		if accountInfo, err := s.accountClient.Infos3(c, &accountGRPC.MidsReq{Mids: mids}); err != nil {
			log.Error("[up_rcmd_banned]GetBannedRcmdList err: %s", err.Error())
			return nil, err
		} else {
			for i := range resp.Items {
				mid := resp.Items[i].Mid
				if u, ok := accountInfo.Infos[mid]; ok {
					resp.Items[i].UName = u.Name
				}
			}
		}
		return resp, nil
	}
}

func (s *Service) AddBannedRcmd(c context.Context, param *model.UpRcmdBlackListCreateReq) (resp *model.UpRcmdBlackListCreateRep, err error) {
	successMids := make([]int64, 0)
	failMids := make([]int64, 0)
	var accountInfo *accountGRPC.InfosReply
	if accountInfo, err = s.accountClient.Infos3(c, &accountGRPC.MidsReq{Mids: param.Mids}); err != nil {
		log.Error("[up_rcmd_banned]GetBannedRcmdList err: %s", err.Error())
		return resp, err
	} else {
		for _, value := range param.Mids {
			mid := value
			if u, ok := accountInfo.Infos[mid]; ok {
				successMids = append(successMids, u.Mid)
			} else {
				failMids = append(failMids, mid)
			}
		}
	}
	if len(successMids) > 0 {
		if err = s.dao.CreateBannedRcmdMids(c, successMids); err != nil {
			return resp, err
		}
	}

	resp = new(model.UpRcmdBlackListCreateRep)
	resp.FailMids = failMids

	return resp, nil
}

func (s *Service) DeleteBannedRcmd(c context.Context, param *model.UpRcmdBlackListDeleteReq) (err error) {
	return s.dao.DeleteBannedRcmdMid(c, param.Mid)
}

//func (s *Service) SearchMidInfo(c context.Context, param *model.UserInfoSearchReq) (resp *model.UserInfoSearchRep, err error) {
//	mids := param.Mids
//	resp = new(model.UserInfoSearchRep)
//	resp.Items = make([]*model.UserInfo, 0)
//	if accountInfo, err := s.accountClient.Infos3(c, &accountGRPC.MidsReq{Mids: mids}); err != nil {
//		log.Error("[up_rcmd_banned]GetBannedRcmdList err: %s", err.Error())
//		return nil, err
//	} else {
//		for _, value := range mids {
//			mid := value
//			if u, ok := accountInfo.Infos[mid]; ok {
//				resp.Items = append(resp.Items, &model.UserInfo{
//					Mid:   mid,
//					UName: u.Name,
//				})
//			}
//		}
//	}
//	return resp, nil
//}
