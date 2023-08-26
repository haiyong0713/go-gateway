package adapters

import (
	"context"
	"encoding/json"
	"go-common/library/database/sql"
	xecode "go-common/library/ecode"
	"go-gateway/app/web-svr/activity/interface/component"
	"go-gateway/app/web-svr/activity/interface/dao/vote"
)

const (
	sql4GetOperPic = `
SELECT data
FROM act_web_data
WHERE id = ?`
)

var OperationPicDS = &operationPicDataSource{}

// operationVideo: 运营数据源(图片)
type operationPic struct {
	Id             int64  `json:"id"`
	PicUrl         string `json:"url"`
	PicDescription string `json:"desc"`
}

func (i *operationPic) GetName() string {
	return i.PicDescription
}

func (i *operationPic) GetId() int64 {
	return i.Id
}

func (i *operationPic) GetSearchField1() string {
	return i.PicDescription
}

func (i *operationPic) GetSearchField2() string {
	return ""
}

func (i *operationPic) GetSearchField3() string {
	return ""
}

type operationPicDataSource struct {
}

type operPicConfig struct {
	PicInfoJson string `json:"pic_info_json"`
}

func (m *operationPicDataSource) ListAllItems(ctx context.Context, sourceId int64) (res []vote.DataSourceItem, err error) {
	res = make([]vote.DataSourceItem, 0)
	var midStr string
	err = component.GlobalDB.QueryRow(ctx, sql4GetOperPic, sourceId).Scan(&midStr)
	if err != nil {
		if err == sql.ErrNoRows {
			err = xecode.Error(xecode.RequestErr, "未找到该数据源ID")
		}
		return
	}
	picConfig := &operPicConfig{}
	err = json.Unmarshal([]byte(midStr), &picConfig)
	if err != nil || picConfig.PicInfoJson == "" {
		err = xecode.Error(xecode.RequestErr, "该数据源配置错误, 请检查")
		return
	}

	picList := make([]*operationPic, 0)
	err = json.Unmarshal([]byte(picConfig.PicInfoJson), &picList)
	if err != nil {
		err = xecode.Error(xecode.RequestErr, "该数据源配置错误, 请检查")
		return
	}
	if len(picList) == 0 {
		err = xecode.Error(xecode.RequestErr, "该数据源配置为空, 请检查")
		return
	}
	existsMap := make(map[int64]struct{}, len(picList))
	for _, p := range picList {
		_, ok := existsMap[p.Id]
		if ok {
			err = xecode.Error(xecode.RequestErr, "该数据源下图片ID重复, 请检查")
			return
		}
		tp := p
		res = append(res, tp)
		existsMap[p.Id] = struct{}{}
	}
	return
}

func (m *operationPicDataSource) NewEmptyItem() vote.DataSourceItem {
	return &operationPic{}
}
