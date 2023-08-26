package resource

import (
	"context"
	"time"

	"go-common/library/database/sql"
	"go-common/library/log"
	resourcegrpc "go-gateway/app/app-svr/resource/service/api/v1"
	"go-gateway/app/web-svr/web-show/interface/model/resource"
	bvidTool "go-gateway/pkg/idsafe/bvid"

	resv2grpc "git.bilibili.co/bapis/bapis-go/resource/service/v2"
)

const (
	_selAllResSQL    = `SELECT id,platform,name,parent,counter,position FROM resource WHERE state=0 ORDER BY counter desc`
	_selAllAssignSQL = `SELECT id,name,contract_id,resource_id,pic,litpic,url,atype,weight,rule,agency FROM resource_assignment WHERE stime<? AND etime>? AND state=0 ORDER BY weight,stime desc`
	_selDefBannerSQL = `SELECT id,name,contract_id,resource_id,pic,litpic,url,atype,weight,rule FROM default_one WHERE  state=0`
)

func (dao *Dao) initRes() {
	dao.selAllResStmt = dao.db.Prepared(_selAllResSQL)
	dao.selAllAssignStmt = dao.db.Prepared(_selAllAssignSQL)
	dao.selDefBannerStmt = dao.db.Prepared(_selDefBannerSQL)
}

// Resources get resource infos from db
func (dao *Dao) Resources(c context.Context) (rscs []*resource.Res, err error) {
	rows, err := dao.selAllResStmt.Query(c)
	if err != nil {
		log.Error("dao.selAllResStmt query error (%v)", err)
		return
	}
	defer rows.Close()
	rscs = make([]*resource.Res, 0)
	for rows.Next() {
		rsc := &resource.Res{}
		if err = rows.Scan(&rsc.ID, &rsc.Platform, &rsc.Name, &rsc.Parent, &rsc.Counter, &rsc.Position); err != nil {
			PromError("Resources", "rows.scan err(%v)", err)
			return
		}
		rscs = append(rscs, rsc)
	}
	if err = rows.Err(); err != nil {
		log.Error("rows.Err() error=%+v", err)
		return nil, err
	}
	return
}

// Assignment get assigment from db
func (dao *Dao) Assignment(c context.Context) (asgs []*resource.Assignment, err error) {
	rows, err := dao.selAllAssignStmt.Query(c, time.Now(), time.Now())
	if err != nil {
		log.Error("dao.selAllAssignmentStmt query error (%v)", err)
		return
	}
	defer rows.Close()
	asgs = make([]*resource.Assignment, 0)
	for rows.Next() {
		asg := &resource.Assignment{}
		if err = rows.Scan(&asg.ID, &asg.Name, &asg.ContractID, &asg.ResID, &asg.Pic, &asg.LitPic, &asg.URL, &asg.Atype, &asg.Weight, &asg.Rule, &asg.Agency); err != nil {
			PromError("Assignment", "rows.scan err(%v)", err)
			return
		}
		asgs = append(asgs, asg)
	}
	if err = rows.Err(); err != nil {
		log.Error("rows.Err() error=%+v", err)
		return nil, err
	}
	return
}

// DefaultBanner set
func (dao *Dao) DefaultBanner(c context.Context) (asg *resource.Assignment, err error) {
	row := dao.selDefBannerStmt.QueryRow(c)
	asg = &resource.Assignment{}
	if err = row.Scan(&asg.ID, &asg.Name, &asg.ContractID, &asg.ResID, &asg.Pic, &asg.LitPic, &asg.URL, &asg.Atype, &asg.Weight, &asg.Rule); err != nil {
		if err == sql.ErrNoRows {
			err = nil
		} else {
			PromError("DefaultBanner", "dao.DefaultBanner.QueryRow error(%v)", err)
		}
	}
	return
}

func (d *Dao) FrontPage(c context.Context, resid int64) (*resourcegrpc.FrontPageResp, error) {
	res, err := d.ResourceClient.FrontPage(c, &resourcegrpc.FrontPageReq{ResourceId: resid})
	if err != nil {
		log.Error("FrontPage %v", err)
		return nil, err
	}
	return res, nil
}

func (d *Dao) PageHeader(ctx context.Context, resourceID int64, ip string) (*resv2grpc.FrontPageConfig, error) {
	return d.resv2Client.GetFrontPageConfig(ctx, &resv2grpc.GetFrontPageConfigReq{ResourceId: resourceID, Ip: ip})
}

func (d *Dao) CheckCommonBWList(ctx context.Context, bvid string, aid int64) bool {
	if bvid == "" && aid == 0 {
		return false
	}
	if bvid == "" {
		var err error
		if bvid, err = bvidTool.AvToBv(aid); err != nil {
			log.Error("dao.CheckCommonBWList bvid:%s, aid:%d, err:%v", bvid, aid, err)
			return false
		}
	}
	rep, err := d.resv2Client.CheckCommonBWList(ctx, &resv2grpc.CheckCommonBWListReq{
		Oid:   bvid,
		Token: d.BanResGRPCToken,
	})
	if err != nil {
		log.Error("dao.CheckCommonBWList bvid:%s, token:%s, err:%v", bvid, d.BanResGRPCToken, err)
		return false
	}
	return rep.IsInList
}
