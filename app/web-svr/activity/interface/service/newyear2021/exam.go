package newyear2021

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"go-gateway/app/web-svr/activity/ecode"
	"go-gateway/app/web-svr/activity/interface/component"
	dao "go-gateway/app/web-svr/activity/interface/dao/newyear2021"
	model "go-gateway/app/web-svr/activity/interface/model/newyear2021"
	"go-gateway/app/web-svr/activity/interface/tool"

	"github.com/Shopify/sarama"
)

const (
	bizNameOfExamPub = "bnj2021_exam_pub"

	liveStatusOfNotStart = "not_start"
	liveStatusOfStarted  = "started"
)

var (
	examBank       []*model.BnjExamItem
	originExamBank []*model.BnjExamItem
)

func init() {
	examBank = make([]*model.BnjExamItem, 0)
	originExamBank = make([]*model.BnjExamItem, 0)
	currentUnix = time.Now().Unix()
	go updateExamBankInRealTime()
}

func updateExamBankInRealTime() {
	ticker := time.NewTicker(time.Second)
	defer func() {
		ticker.Stop()
	}()

	for {
		select {
		case <-ticker.C:
			UpdateExamBank(model.DeepCopyExamBank(originExamBank))
		}
	}
}

func updateExamBankByFilename(filename string) {
	if filename == "" {
		return
	}

	if d, err := readExamBank(filename); err == nil {
		originExamBank = model.DeepCopyExamBank(d)
		examBank = model.DeepCopyExamBank(d)
		UpdateExamBank(examBank)
	}
}

func readExamBank(filename string) (list []*model.BnjExamItem, err error) {
	list = make([]*model.BnjExamItem, 0)
	var bs []byte
	bs, err = readByteSliceByFilename(filename)
	if err == nil {
		err = json.Unmarshal(bs, &list)
	}

	return
}

func UpdateExamBank(list []*model.BnjExamItem) {
	if len(list) == 0 {
		return
	}

	stats, _ := dao.FetchExamStats(context.Background())
	for _, v := range list {
		for _, opt := range v.Options {
			key := fmt.Sprintf("%v_%v", v.ID, opt.ID)
			if d, ok := stats[key]; ok {
				opt.Count = d
			}
		}

		v.Rebuild(currentUnix)
	}

	examBank = model.DeepCopyExamBank(list)
}

func (s *Service) LiveStatus(ctx context.Context) (m map[string]interface{}, err error) {
	m = make(map[string]interface{}, 0)
	{
		m["status"] = liveStatusOfNotStart
	}

	if currentUnix >= BnjStrategyInfo.LiveStartTime {
		m["status"] = liveStatusOfStarted
	}

	return
}

func (s *Service) ExamDetail(ctx context.Context, mid int64) (resp *model.BnjExamResponse, err error) {
	var commitM map[string]int64
	resp = new(model.BnjExamResponse)
	list := model.DeepCopyExamBank(examBank)
	if mid > 0 {
		commitM, _ = dao.FetchUserCommitDetail(ctx, mid)
		for _, v := range list {
			if currentUnix < v.StartTime {
				continue
			}

			if d, ok := commitM[v.IDStr]; ok {
				v.UserOpt = d
			}
		}
	}

	resp.Bank = list
	resp.Sleep = 10

	return
}

func (s *Service) CommitUserAnswer(ctx context.Context, mid int64, report *model.RiskManagementReportInfoOfExam) (err error) {
	if !canUserCommitAnswer(report.OrderID, report.UserAnswer) {
		err = ecode.BNJExamInvalidCommit

		return
	}

	var affect bool
	affect, err = dao.CommitUserOption(ctx, mid, report.OrderID, report.UserAnswer)
	if err == nil {
		if !affect {
			err = ecode.BNJExamDuplicatedCommit

			return
		}

		examPub := new(model.ExamPub)
		{
			examPub.MID = report.MID
			examPub.ItemID = report.OrderID
			examPub.OptID = report.UserAnswer
			examPub.LogTime = time.Now().Unix()
		}
		// pub into ClickHouse consumer
		pubExamLog2CH(examPub)
		// report risk
		_ = component.Report2RiskManagement(ctx, component.RiskManagementScene4Exam, "", report)
	}

	if err != nil {
		err = ecode.BNJTooManyUser
	}

	return
}

func pubExamLog2CH(examLog *model.ExamPub) {
	if BnjExamProducer.Topic == "" {
		return
	}

	bs, _ := json.Marshal(examLog)
	msg := &sarama.ProducerMessage{
		Topic:    BnjExamProducer.Topic,
		Value:    sarama.ByteEncoder(bs),
		Metadata: "bnj2021_exam",
	}

	_, _, pubErr := ExamProducer.SendMessage(msg)
	if pubErr != nil {
		tool.IncrCommonBizStatus(bizNameOfExamPub, tool.StatusOfFailed)
	}
}

func canUserCommitAnswer(itemID, optID int64) bool {
	for _, v := range examBank {
		if v.ID == itemID {
			if currentUnix >= v.StartTime && currentUnix < v.EndTime {
				for _, d := range v.Options {
					if d.ID == optID {
						return true
					}
				}
			}

			break
		}
	}

	return false
}
