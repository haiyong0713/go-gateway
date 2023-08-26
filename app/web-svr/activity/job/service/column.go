package service

import (
	"context"
	"fmt"
	"go-common/library/log"
	"go-common/library/net/trace"
	"go-gateway/app/web-svr/activity/job/model/column"
	mdlmail "go-gateway/app/web-svr/activity/job/model/mail"
	"sort"
	"strings"
	"time"
)

func (s *Service) ColumnDataExport() {
	var (
		limit     = int64(50)
		page      = int64(0)
		items     column.Items
		reply     column.Items
		yesterday = time.Now().AddDate(0, 0, -1).Format("2006-01-02")
		sTime     = fmt.Sprintf("%s 00:00:00", yesterday)
		eTime     = fmt.Sprintf("%s 23:59:59", yesterday)
	)
	c := trace.SimpleServerTrace(context.Background(), "ColumnDataExport")
	log.Infoc(c, "ColumnDataExport Start")
	for {
		result, err := s.dao.GetIdWidStateBySQL(c, s.c.Column.Sid, limit, page*limit)
		if err != nil {
			log.Errorc(c, "GetIdWidStateBySQL Err:%v", err)
			return
		}
		page++
		// 收集lid
		var lids []int64
		for _, v := range result {
			if v.State == 1 {
				lids = append(lids, v.ID)
			}
		}

		log.Infoc(c, "s.dao.GetIdWidStateBySQL collection lids %v", lids)

		if len(lids) == 0 {
			return
		}
		// 获取总点赞数
		totalLikes, err := s.dao.BatchLikeActSum(c, lids)
		if err != nil {
			log.Errorc(c, "BatchLikeActSum Err:%v", err)
			return
		}

		log.Infoc(c, "s.dao.BatchLikeActSum totalLikes %v", totalLikes)

		// 赋值
		for _, v := range result {
			if v.State == 1 {
				if likes, ok := totalLikes[v.ID]; ok {
					items = append(items, &column.Item{Id: v.ID, Wid: v.Wid, TotalLikes: likes})
				} else {
					// 如果没有用户点赞 那么明细表拿不到数据 要赋值为0
					items = append(items, &column.Item{Id: v.ID, Wid: v.Wid, TotalLikes: int64(0)})
				}
			}
		}

		log.Infoc(c, "batch items len %v detail %v", len(totalLikes), totalLikes)

		if int64(len(result)) < limit {
			break
		}
	}

	if len(items) == 0 {
		log.Infoc(c, "len(items) == 0 exit")
		return
	}

	// 排序
	sort.Sort(items)

	for _, value := range items {
		// 请求当日点赞数
		var lids []int64
		lids = append(lids, value.Id)
		likesSet, err := s.dao.BatchLikeActSumRangeTime(c, lids, sTime, eTime)
		if err != nil {
			log.Errorc(c, "BatchLikeActSumRangeTime lid:%v Err:%v", lids, err)
			return
		}

		log.Infoc(c, "s.dao.BatchLikeActSumRangeTime lid %v", lids)

		if v, ok := likesSet[value.Id]; ok {
			value.Likes = v
		}

		// 获取标题详情
		var wids []int64
		wids = append(wids, value.Wid)
		artsRes, err := s.arts(c, wids, 1)
		if err != nil {
			log.Errorc(c, "arts aids:%v Err:%v", value.Wid, err)
			return
		}

		log.Infoc(c, "get s.arts wid %v result %v", value.Wid, artsRes)

		art, ok := artsRes[value.Wid]
		if !ok {
			log.Errorc(c, "get art info from arts Err aids:%v Err:%v", value.Wid, err)
			continue
		}

		if !art.IsNormal() {
			continue
		}

		value.Title = art.Title

		reply = append(reply, value)
	}

	splitNum := 200
	if len(reply) < 200 {
		splitNum = len(reply)
	}

	// 获取前200条数据 循环查询今天点赞数和稿件标题信息
	data := reply[:splitNum]

	// 邮件
	if err := s.sendEmailWithCSVFile(c, data); err != nil {
		log.Errorc(c, "sendEmailWithCSVFile Err:%v", err)
	}

	return
}

func (s *Service) sendEmailWithCSVFile(c context.Context, data column.Items) error {
	categoryHeader := []string{"标题", "昨日推荐次数", "累计推荐次数", "CV号"}
	mailData := [][]string{}
	for _, v := range data {
		rows := []string{}
		rows = append(rows, v.Title, fmt.Sprint(v.Likes), fmt.Sprint(v.TotalLikes), fmt.Sprintf("cv%v", v.Wid))
		mailData = append(mailData, rows)
	}
	fileName := fmt.Sprintf("%s_%s.csv", time.Now().Format("20060102"), "专栏热搜活动昨天Top200")
	err := s.columnCreateCsvAndSend(c, "./data/", fileName, fileName, categoryHeader, mailData)
	if err != nil {
		return err
	}
	return nil
}

func (s *Service) columnCreateCsvAndSend(c context.Context, filePath string, fileName string, subject string, categoryHeader []string, data [][]string) error {
	base := &mdlmail.Base{
		Host:    s.c.Mail.Host,
		Port:    s.c.Mail.Port,
		Address: s.c.Mail.Address,
		Pwd:     s.c.Mail.Pwd,
		Name:    s.c.Mail.Name,
	}
	return s.activityCreateCsvAndSend(c, filePath, fileName, subject, base, s.columnBuildReceivers(), []*mdlmail.Address{}, []*mdlmail.Address{}, categoryHeader, data)
}

func (s *Service) columnBuildReceivers() []*mdlmail.Address {
	var mailReceivers []*mdlmail.Address

	receivers := strings.Split(s.c.Column.EmailReceivers, ",")
	for _, v := range receivers {
		user := &mdlmail.Address{
			Address: v,
			Name:    "",
		}
		mailReceivers = append(mailReceivers, user)
	}
	return mailReceivers
}
