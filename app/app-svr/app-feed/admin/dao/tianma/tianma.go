package tianma

import (
	"go-common/library/log"

	"go-gateway/app/app-svr/app-feed/admin/model/tianma"
)

// 查询某一个 file_status 的全部列表
func (d *Dao) SearchPosRecListByStatus(fileStatus int) (resList []*tianma.PosRecItem, err error) {
	err = d.DB.Model(&tianma.PosRecItem{}).Where(&tianma.PosRecItem{
		FileStatus: fileStatus,
	}).Find(&resList).Error
	if err != nil {
		log.Error("PosRecListSearch error %v", err)
	}
	return
}

// 根据file_path_addition查询投放列表 （商业dmp id 或 hdfs地址）
func (d *Dao) SearchPosRecListByTypeAddition(fileStatus int, fileTypeAdditionArr []int) (resList []*tianma.PosRecItem, err error) {
	err = d.DB.Model(&tianma.PosRecItem{}).Where("file_status=?", fileStatus).
		Where("file_type_addition in (?)", fileTypeAdditionArr).
		Find(&resList).Error
	if err != nil {
		log.Error("dao.SearchPosRecListByStateAddition error %v", err)
	}
	return
}

// 更改某一个推荐的 file_status、file_path、file_rows
func (d *Dao) UpdatePosRecItemById(id int64, newItem *tianma.PosRecItem) (err error) {
	err = d.DB.Model(&tianma.PosRecItem{
		Id: id,
	}).Update(&newItem).Error
	if err != nil {
		log.Error("UpdatePosRecItemById error %v", err)
	}
	return
}
