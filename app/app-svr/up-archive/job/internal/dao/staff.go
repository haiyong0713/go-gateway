package dao

import (
	"context"

	videoUpOpen "git.bilibili.co/bapis/bapis-go/videoup/open/service"
	"go-common/library/ecode"

	"go-common/library/log"
)

func (d *dao) RawStaffAids(ctx context.Context, mid int64) ([]int64, error) {
	info, err := d.videoUpOpenClient.GetArchiveStaffAid(ctx, &videoUpOpen.GetArchiveStaffAidReq{StaffMid: mid})
	if err != nil {
		log.Error("日志告警 db to grpc GetArchiveStaffAid mid(%d) error(%+v)", mid, err)
		if ecode.EqualError(ecode.NothingFound, err) {
			return nil, nil
		}
		return nil, err
	}
	return info.Aid, nil
}

func (d *dao) RawStaffMids(ctx context.Context, aid int64) ([]int64, error) {
	info, err := d.videoUpOpenClient.GetArchiveStaffMid(ctx, &videoUpOpen.GetArchiveStaffMidReq{Aid: aid})
	if err != nil {
		log.Error("日志告警 db to grpc GetArchiveStaffMid aid(%d) error(%+v)", aid, err)
		if ecode.EqualError(ecode.NothingFound, err) {
			return nil, nil
		}
		return nil, err
	}
	return info.StaffMid, nil
}
