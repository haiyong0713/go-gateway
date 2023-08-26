package intervention

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"go-common/library/ecode"
	"go-common/library/log"

	arcgrpc "git.bilibili.co/bapis/bapis-go/archive/service"
	"go-gateway/app/app-svr/app-feed/admin/conf"
	"go-gateway/app/app-svr/app-feed/admin/dao/manager"
	"go-gateway/app/app-svr/app-feed/admin/model/common"
	"go-gateway/app/app-svr/app-feed/admin/model/intervention"
	"go-gateway/pkg/idsafe/bvid"
)

var ctx = context.Background()

// Service is search service
type Service struct {
	dao       *manager.Dao
	arcClient arcgrpc.ArchiveClient
}

// New new a search service
func New(c *conf.Config) (s *Service) {
	var (
		err error
	)
	s = &Service{
		dao: manager.New(c),
	}
	if s.arcClient, err = arcgrpc.NewClient(c.ArcClient); err != nil {
		panic(err)
	}
	return
}

type ErrorResponse struct {
	code    ecode.Code
	message string
}

var (
	eExistOnline = ErrorResponse{
		code:    77501,
		message: "该稿件已存在 待上线/待生效/生效中 的干预",
	}
	eStartTimeLate = ErrorResponse{
		code:    77502,
		message: "生效时间不得晚于失效时间",
	}
	eStartTimeEarly = ErrorResponse{
		code:    77503,
		message: "生效时间不得早于当前时间",
	}
	eNotExist = ErrorResponse{
		code:    77504,
		message: "未查找到对应ID干预记录",
	}
	eBvidErr = ErrorResponse{
		code:    77505,
		message: "bvid错误",
	}
)

func fastEcode(errorType ErrorResponse) (status *ecode.Status) {
	status = ecode.Error(errorType.code, errorType.message)
	return
}

// 列表单个item详情
type DetailResItem struct {
	*intervention.Detail
	Bvid       string `json:"bvid"`
	BvidList   string `json:"bvid_list"`
	Status     int    `json:"status"`
	StatusText string `json:"status_text"`
}

// 列表详情返回结果
type DetailPager struct {
	Item []*DetailResItem `json:"item"`
	Page common.Page      `json:"page"`
}

// 创建新干预
func (s *Service) CreateIntervention(request *intervention.Detail, username string) (detail *DetailResItem, err error) {
	if request.StartTime >= request.EndTime {
		err = fastEcode(eStartTimeLate)
		return
	}
	//nolint:staticcheck
	count, err := s.dao.CheckIntervention(request)
	if count >= 1 {
		err = fastEcode(eExistOnline)
		return
	}
	request.CreatedBy = username
	request.List = strings.Replace(request.List, " ", "", -1)

	var archives *arcgrpc.ArcsReply
	if archives, err = s.arcClient.Arcs(ctx, &arcgrpc.ArcsRequest{Aids: []int64{request.Avid}}); err != nil {
		log.Error("s.arcClient.Arcs error %v", err)
		return
	}
	if archives.Arcs[request.Avid] != nil {
		request.Title = archives.Arcs[request.Avid].Title
	}

	var id uint
	if id, err = s.dao.InsertDetail(request); err != nil {
		return
	}

	_ = s.CreateOptLog(&intervention.OptLogDetail{
		Avid:           request.Avid,
		InterventionId: id,
		OpUser:         username,
		OpType:         0,
	})

	fmt.Println(username)
	item, err := s.dao.FindInterventionById(id)
	if err != nil {
		return
	}
	detail = translateAvids(&item)
	return
}

// 编辑已有干预
func (s *Service) EditIntervention(request *intervention.Detail, username string) (detail *DetailResItem, err error) {
	if request.StartTime >= request.EndTime {
		err = fastEcode(eStartTimeLate)
		return
	}
	if request.StartTime <= time.Now().Unix() {
		err = fastEcode(eStartTimeEarly)
		return
	}

	item, err := s.dao.FindInterventionById(request.ID)
	if err != nil {
		err = fastEcode(eNotExist)
		log.Error(err.Error())
		return
	}

	//nolint:staticcheck
	count, err := s.dao.CheckIntervention(request)
	if count >= 1 {
		err = fastEcode(eExistOnline)
		return
	}
	if err = s.dao.EditDetail(request); err != nil {
		return
	}

	//nolint:staticcheck
	before, err := json.Marshal(&intervention.Detail{
		List:      item.List,
		StartTime: item.StartTime,
		EndTime:   item.EndTime,
	})
	//nolint:staticcheck
	after, err := json.Marshal(&intervention.Detail{
		List:      request.List,
		StartTime: request.StartTime,
		EndTime:   request.EndTime,
	})
	_ = s.CreateOptLog(&intervention.OptLogDetail{
		Avid:           request.Avid,
		InterventionId: request.ID,
		OpUser:         username,
		OpType:         2,
		MBefore:        string(before),
		MAfter:         string(after),
	})

	fmt.Println(username)

	item, err = s.dao.FindInterventionById(request.ID)
	if err != nil {
		err = fastEcode(eNotExist)
		log.Error(err.Error())
		return
	}
	detail = translateAvids(&item)
	return
}

// 切换某个干预状态
func (s *Service) ChangeIntervention(request *intervention.Detail, username string) (detail *DetailResItem, err error) {
	var newStatus int64
	if request.OnlineStatus == 1 {
		// 1 表示手动上线
		newStatus = 1
	} else {
		// 2 表示手动下线
		newStatus = 2
	}

	item, err := s.dao.FindInterventionById(request.ID)
	if err != nil {
		err = fastEcode(eNotExist)
		log.Error(err.Error())
		return
	}

	if newStatus == 1 {
		count := 0
		//nolint:staticcheck
		count, err = s.dao.CheckIntervention(&item)
		if count >= 1 {
			err = fastEcode(eExistOnline)
			log.Error(err.Error())
			return
		}
	}

	if err = s.dao.ChangeStatus(&item, newStatus); err != nil {
		log.Error(err.Error())
		return
	}

	_ = s.CreateOptLog(&intervention.OptLogDetail{
		Avid:           item.Avid,
		InterventionId: item.ID,
		OpUser:         username,
		OpType:         1,
		MBefore:        strconv.FormatInt(item.OnlineStatus, 10),
		MAfter:         strconv.FormatInt(newStatus, 10),
	})

	fmt.Println(username)

	detail = translateAvids(&item)
	return
}

// 搜索干预列表
func (s *Service) DetailList(filters *intervention.Detail, pageNum int) (result DetailPager, err error) {
	if filters.Avid == 0 && filters.Bvid != "" {
		filters.Avid, err = strconv.ParseInt(filters.Bvid, 10, 64)
		if err != nil {
			// 不是avid，就当做bvid转换成avid
			filters.Avid, err = bvid.BvToAv(filters.Bvid)
			if err != nil {
				err = fastEcode(eBvidErr)
				return
			}
		}
	}

	list, err := s.dao.List(filters, pageNum)
	if err != nil {
		log.Error("s.dao.List error %v", err)
		return
	}

	ids := make([]int64, len(list))

	items := make([]*DetailResItem, len(list))

	for index := range list {
		ids[index] = list[index].Avid
		items[index] = translateAvids(&list[index])
	}

	var archives *arcgrpc.ArcsReply
	if archives, err = s.arcClient.Arcs(ctx, &arcgrpc.ArcsRequest{Aids: ids}); err != nil {
		log.Error("s.arcClient.Arcs error %v", err)
		return
	}

	for index, value := range list {
		//nolint:gomnd
		if items[index].OnlineStatus == 2 {
			items[index].OnlineStatus = 0
		}
		if archives.Arcs[value.Avid] != nil {
			items[index].Pic = archives.Arcs[value.Avid].Pic
			items[index].Title = archives.Arcs[value.Avid].Title
		}
	}

	count, err := s.dao.ListCount(filters)
	if err != nil {
		log.Error(err.Error())
		return
	}

	result.Page = common.Page{
		Total: count,
		Num:   pageNum,
		Size:  20,
	}
	result.Item = items
	return
}

func translateAvids(item *intervention.Detail) *DetailResItem {
	// 转换稿件bvid
	bid, err := bvid.AvToBv(item.Avid)
	if err != nil {
		log.Error(err.Error())
		bid = "稿件BVID转换出错"
	}

	// 分割干预列表
	aidInList := strings.Split(item.List, ",")
	bidList := make([]string, len(aidInList))

	// 转换干预列表内稿件bvid
	for i := 0; i < len(aidInList); i++ {
		var bid64 string
		aid64, err := strconv.ParseInt(aidInList[i], 10, 64)
		if err != nil {
			log.Error(err.Error())
			//nolint:ineffassign
			bid64 = "干预列表稿件BVID转换出错"
		}
		bid64, err = bvid.AvToBv(aid64)
		if err != nil {
			log.Error(err.Error())
			bid64 = "干预列表稿件BVID转换出错"
		}
		bidList[i] = bid64
	}

	statusText := ""
	status := 0
	now := time.Now().Unix()
	//nolint:gomnd
	switch item.OnlineStatus {
	case 0:
		if now > item.EndTime {
			statusText = "已失效"
		} else {
			statusText = "待上线"
		}
	case 1:
		if now < item.StartTime {
			statusText = "待生效"
		} else if now >= item.StartTime && now <= item.EndTime {
			statusText = "生效中"
			status = 1
		} else if now > item.EndTime {
			statusText = "已失效"
		}
	case 2:
		statusText = "已失效"
	}

	return &DetailResItem{
		Detail:     item,
		Bvid:       bid,
		BvidList:   strings.Join(bidList, ","),
		StatusText: statusText,
		Status:     status,
	}
}

// 操作日志单个item详情
type OpDetailResItem struct {
	intervention.OptLogDetail
	BvId       string `json:"bvid"`
	OpTypeText string `json:"op_type_text"`
}

// 操作日志列表返回结果
type OpDetailPager struct {
	Item []*OpDetailResItem `json:"item"`
	Page common.Page        `json:"page"`
}

// 插入一条操作日志
func (s *Service) CreateOptLog(request *intervention.OptLogDetail) (err error) {
	if err = s.dao.InsertOptLog(request); err != nil {
		log.Error(err.Error())
		return
	}
	return
}

// 查询操作日志
func (s *Service) OpLogList(filters *intervention.OptLogDetail, pageNum int) (result OpDetailPager, err error) {
	list, err := s.dao.OptLogList(filters, pageNum)
	if err != nil {
		log.Error(err.Error())
		return
	}

	items := make([]*OpDetailResItem, len(list))
	for index, item := range list {
		// 转换稿件bvid
		bid, err := bvid.AvToBv(item.Avid)
		if err != nil {
			log.Error(err.Error())
			bid = "稿件BVID转换出错"
		}

		opTypeText := ""
		//nolint:gomnd
		switch item.OpType {
		case 0:
			opTypeText = "新增"
		case 1:
			if item.MAfter == "1" {
				opTypeText = "上线"
			} else {
				opTypeText = "下线"
			}
		case 2:
			opTypeText = "编辑-生效时间/干预列表"
		}

		items[index] = &OpDetailResItem{item, bid, opTypeText}
	}

	count, err := s.dao.OpLogListCount(filters)
	if err != nil {
		log.Error(err.Error())
		return
	}

	result.Page = common.Page{
		Total: count,
		Num:   pageNum,
		Size:  20,
	}
	result.Item = items

	return
}
