package favorite

import (
	"context"
	api "git.bilibili.co/bapis/bapis-go/community/service/favorite"
)

func (d *Dao) AddFolder(c context.Context, req *api.AddFolderReq) (fid int64, err error) {
	addFolderReply, err := d.favClient.AddFolder(c, req)
	if addFolderReply != nil {
		return addFolderReply.Fid, err
	}
	return
}

// 清空播单，然后添加新的数据
func (d *Dao) MultiReplace(c context.Context, req *api.MultiReplaceReq) (err error) {
	_, err = d.favClient.MultiReplace(c, req)
	return
}
