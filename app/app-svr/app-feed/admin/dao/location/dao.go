package location

import (
	"context"
	"go-common/library/log"
	"go-gateway/app/app-svr/app-feed/admin/conf"
	"strconv"

	locgrpc "git.bilibili.co/bapis/bapis-go/community/service/location"
	locadmingrpc "git.bilibili.co/bapis/bapis-go/platform/admin/location"

	"github.com/pkg/errors"
)

// Dao is location dao.
type Dao struct {
	locGRPC      locgrpc.LocationClient
	locAdminGRPC locadmingrpc.PolicyClient
}

// New new a location dao.
func New(c *conf.Config) (d *Dao) {
	d = &Dao{}
	var err error
	if d.locGRPC, err = locgrpc.NewClient(c.LocationClient); err != nil {
		panic(err)
	}
	if d.locAdminGRPC, err = locadmingrpc.NewClient(c.LocationAdminClient); err != nil {
		panic(err)
	}
	return
}

// AddPolicy is
func (d *Dao) AddPolicy(c context.Context, areaIDs []int64) (pid int64, err error) {
	var (
		arg   = &locgrpc.AddPolicyReq{AreaIds: areaIDs, PlayAuth: locgrpc.Status_Forbidden, DownAuth: locgrpc.StatusDown_ForbiddenDown}
		reply *locgrpc.AddPolicyReply
	)
	if reply, err = d.locGRPC.AddPolicy(c, arg); err != nil {
		err = errors.Wrapf(err, "arg(%+v)", arg)
		return
	}
	if reply != nil {
		pid = reply.PolicyId
	}
	return
}

// PolicyInfo is
func (d *Dao) PolicyInfo(c context.Context, pid int64) (areaIDs []int64, err error) {
	var (
		arg   = &locgrpc.PolicyInfoReq{PolicyId: pid}
		reply *locgrpc.PolicyInfo
	)
	if reply, err = d.locGRPC.PolicyInfo(c, arg); err != nil {
		err = errors.Wrapf(err, "d.locGRPC.PolicyInfo arg(%+v)", arg)
		return
	}
	if reply != nil {
		areaIDs = reply.AreaIds
	}
	return
}

// PolicyInfo is
func (d *Dao) PolicyInfos(c context.Context, pids []int64) (res map[int64][]int64, err error) {
	reply, err := d.locGRPC.PolicyInfos(c, &locgrpc.PolicyInfosReq{PolicyIds: pids})
	if err != nil {
		err = errors.Wrapf(err, "d.locGRPC.PolicyInfos arg(%+v)", &locgrpc.PolicyInfosReq{PolicyIds: pids})
		return
	}
	res = make(map[int64][]int64)
	if reply != nil {
		for pid, v := range reply.Infos {
			res[pid] = v.AreaIds
		}
	}
	return
}

// AddGroupWithItems 新建策略组及其策略项
func (d *Dao) AddGroupWithItems(c context.Context, policyGroupName string, policyGroupType int32, policyGroupRemark string, areaIDs []int64, playAuth int32, downAuth int32, businessSource string, username string) (pid int64, err error) {
	var reply *locadmingrpc.PolicyGroupWithItemsAddReply
	toAddPolicyItem := &locadmingrpc.PolicyItemParams{
		Id:       0,
		AreaIds:  areaIDs,
		DownAuth: downAuth,
		PlayAuth: playAuth,
	}
	arg := &locadmingrpc.PolicyGroupWithItemsAddReq{
		Name:           policyGroupName,
		Type:           locadmingrpc.PolicyGroupType(policyGroupType),
		Remark:         policyGroupRemark,
		Items:          []*locadmingrpc.PolicyItemParams{toAddPolicyItem},
		BusinessSource: businessSource,
		Username:       username,
	}
	if reply, err = d.locAdminGRPC.AddGroupWithItems(c, arg); err != nil {
		err = errors.Wrapf(err, "arg(%+v)", arg)
		return
	}
	if reply != nil {
		pid = reply.Id
	}
	return
}

// DeleteGroup 删除策略组及其策略项
func (d *Dao) DeleteGroup(c context.Context, policyGroupID int64, businessSource string, username string) (err error) {
	if policyGroupID == 0 {
		return
	}
	arg := &locadmingrpc.PolicyGroupDeleteReq{
		Ids:            strconv.FormatInt(policyGroupID, 10),
		BusinessSource: businessSource,
		Username:       username,
	}
	if _, err = d.locAdminGRPC.DeleteGroup(c, arg); err != nil {
		log.Error("feed-admin.Dao.location.DeleteGroup Error (%v)", err)
		return
	}
	return
}

func (d *Dao) ListGroup(ctx context.Context, groupType locadmingrpc.PolicyGroupType, pn int64, ps int64) (res []*locadmingrpc.PolicyGroupInfo, err error) {
	req := &locadmingrpc.PolicyGroupListReq{
		Type:  groupType,
		Pn:    pn,
		Ps:    ps,
		State: locadmingrpc.OK,
	}
	var rawRes *locadmingrpc.PolicyGroupListReply
	if rawRes, err = d.locAdminGRPC.ListGroup(ctx, req); err != nil {
		err = errors.Wrapf(err, "GetGroups.PolicyGroupList")
		return
	}
	res = rawRes.Items
	return
}
