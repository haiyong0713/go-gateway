package springfestival2021

import (
	"context"
	"go-common/library/log"
)

// GetArchiveNums 获取用户投稿数
func (d *Dao) GetArchiveNums(c context.Context, mid int64) (nums int64, err error) {
	nums, err = d.ArchiveNums(c, mid)
	if err != nil {
		log.Errorc(c, "d.GetArchiveNums mid(%d) err(%v)", mid, err)
	}
	if err == nil {
		return nums, nil
	}
	nums, err = d.ArchiveNumsDB(c, mid)
	if err != nil {
		log.Errorc(c, "d.ArchiveNumsDB(c, %d) err(%v)", mid, err)
		return nums, err
	}
	err = d.AddArchiveNums(c, mid, nums)
	if err != nil {
		log.Errorc(c, "d.AddArchiveNums mid(%d)nums(%d) err(%v)", mid, nums, err)
	}
	return nums, nil
}
