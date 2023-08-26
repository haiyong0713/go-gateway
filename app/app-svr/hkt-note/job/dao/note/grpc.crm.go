package note

import (
	"context"
	"encoding/json"
	crm "git.bilibili.co/bapis/bapis-go/crm/service/profile-manager"
	"go-common/library/log"
	"time"
)

type GroupMid struct {
	Mid int64 `json:"mid"`
}

func (d *Dao) GetGroupMember(ctx context.Context, groupID int64) ([]int64, error) {
	var (
		err       error
		mids      []int64
		reply     *crm.GroupMembersReply
		page      = int64(1)
		defaultPs = int64(50)
		pageMids  []*GroupMid
	)
	for {
		time.Sleep(20 * time.Millisecond)
		reply, err = d.grpc.crm.GetGroupMemberByPage(ctx, &crm.GetGroupMemberByPageReq{
			GroupId:    groupID,
			MemberType: 1,
			Page:       page,
			Size_:      defaultPs,
			Fields:     []string{""},
		})
		if err != nil {
			log.Errorc(ctx, "d.GetGroupMember(%d) error(%v)", groupID, err)
			return nil, err
		}
		err = json.Unmarshal([]byte(reply.UpsJson), &pageMids)
		if err != nil {
			log.Errorc(ctx, "d.GetGroupMember(%s) unmarshal error(%v)", reply.UpsJson, err)
			return nil, err
		}
		for _, v := range pageMids {
			mids = append(mids, v.Mid)
		}
		if len(pageMids) < int(defaultPs) {
			break
		}
		pageMids = nil
		page++
	}
	return mids, nil
}
