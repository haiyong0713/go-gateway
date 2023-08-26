package note

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/railgun"
	"io/ioutil"
	"strings"
	"time"

	archive "git.bilibili.co/bapis/bapis-go/archive/service"
	"go-gateway/app/app-svr/hkt-note/model/util"
)

func (s *Service) CrmGroupArcs(ctx context.Context) ([]int64, error) {
	var (
		mids   []int64
		aids   []int64
		latest = time.Now().Unix() - s.c.NoteCfg.BotCrmLastestPubtime
	)
	for _, id := range s.c.NoteCfg.BotCrmGroups {
		groupMids, err := s.dao.GetGroupMember(ctx, id)
		if err != nil {
			return nil, err
		}
		mids = append(mids, groupMids...)
	}
	mids = util.Int64DuplicateRemoval(mids)
	for _, mid := range mids {
		memberAids, err := s.dao.ArcPassed(ctx, mid, latest)
		if ecode.EqualError(ecode.NothingFound, err) {
			continue
		}
		if err != nil {
			return nil, err
		}
		aids = append(aids, memberAids...)
	}

	log.Warn("s.CrmGroupArcs Record crmMids(%v) crmAids(%v)", len(mids), len(aids))
	return util.Int64DuplicateRemoval(aids), nil

}

func (s *Service) HotArcPushBot(ctx context.Context) railgun.MsgPolicy {
	//1.获取crm人群包用户最近发布稿件
	//2.获取热门稿件
	//3.对历史push过的稿件过滤
	//4.对稿件做条件过滤
	//5.写入push记录
	//6.将push稿件写入文件
	//7.将文件上传到企业微信
	//8.bot push文件到企业微信群

	//crm人群包稿件
	crmAids, err := s.CrmGroupArcs(ctx)
	if err != nil {
		return railgun.MsgPolicyAttempts
	}

	//热门稿件
	hotAids, err := s.dao.HotArchives(ctx)
	if err != nil {
		return railgun.MsgPolicyAttempts
	}

	var (
		allAids []int64
		aids    []int64
	)
	allAids = append(allAids, hotAids...)
	allAids = append(allAids, crmAids...)
	allAids = util.Int64DuplicateRemoval(allAids)

	//push记录过滤
	for _, aid := range allAids {
		pushDate, _ := s.dao.GetBotPushRecord(ctx, aid)
		if pushDate != 0 {
			continue
		}
		aids = append(aids, aid)
	}
	if len(aids) == 0 {
		return railgun.MsgPolicyNormal
	}
	arcs, err := s.dao.BatchArchives(ctx, aids)
	if err != nil {
		return railgun.MsgPolicyAttempts
	}
	var pushArcs []*archive.Arc
	//条件过滤
	for _, arc := range arcs {
		if !arc.IsNormal() {
			continue
		}
		if len(s.c.NoteCfg.BotArcTypes) > 0 && !util.Int32ArrayIn(s.c.NoteCfg.BotArcTypes, arc.TypeID) {
			continue
		}
		if arc.Copyright != 1 {
			continue
		}
		if arc.OrderID != 0 {
			continue
		}
		if arc.Duration < 3*60 || arc.Duration > 30*60 {
			continue
		}
		noteCount, err := s.dao.ArcNotesCount(ctx, arc.Aid, 0)
		if err != nil {
			return railgun.MsgPolicyAttempts
		}
		if noteCount > 0 {
			continue
		}
		pushArcs = append(pushArcs, arc)
	}

	//写入push记录
	pushDate := time.Now().Unix()
	for _, arc := range arcs {
		_ = s.dao.SetBotPushRecord(ctx, arc.Aid, pushDate)
	}

	if len(pushArcs) == 0 {
		return railgun.MsgPolicyNormal
	}
	var pushAids []int64
	for _, arc := range pushArcs {
		pushAids = append(pushAids, arc.Aid)
	}

	log.Warn("s.HotArcPushBot Record crmAids(%v) hotAids(%v) checkAids(%v) pushAids(%v)", len(crmAids), len(hotAids), len(aids), len(pushAids))

	//写入文件
	fileName, err := s.BotPushFile(ctx, pushArcs)
	if err != nil {
		return railgun.MsgPolicyAttempts
	}

	//上传文件
	mediaID, err := s.dao.UploadFile(s.c.NoteCfg.BotKey, fileName)
	if err != nil {
		return railgun.MsgPolicyAttempts
	}

	//push文件到企业微信群
	err = s.dao.SendFile(s.c.NoteCfg.BotKey, mediaID)
	if err != nil {
		return railgun.MsgPolicyAttempts
	}

	return railgun.MsgPolicyNormal
}

const (
	BotPushTitle  = "BV号,稿件名称,类型,发布时间"
	BotPushFormat = "%s,%s,%s,%s"
)

var Bom = []byte("\xEF\xBB\xBF")

func (s *Service) BotPushFile(ctx context.Context, arcs []*archive.Arc) (string, error) {
	var (
		err      error
		bv       string
		result   = []string{BotPushTitle}
		buf      = new(bytes.Buffer)
		w        = bufio.NewWriter(buf)
		fileName = fmt.Sprintf("/tmp/hot_note_archive_%s.csv", time.Now().Format("2006-01-02 15:04:05"))
	)
	_, err = w.Write(Bom)
	if err != nil {
		log.Errorc(ctx, "error writing bom to csv:%v", err)
		return "", err
	}
	for _, arc := range arcs {
		bv, _ = util.AvToBv(arc.Aid)
		result = append(result, fmt.Sprintf(BotPushFormat, bv, arc.Title, arc.TypeName, time.Unix(int64(arc.PubDate), 0).Format("2006-01-02 15:04")))
	}
	if _, err = w.WriteString(strings.Join(result, "\n")); err != nil {
		log.Error("error writing record to csv:%v", err)
		return "", err
	}
	w.Flush()
	err = ioutil.WriteFile(fileName, buf.Bytes(), 0600)
	if err != nil {
		log.Error("error writing record to csv:%v", err)
		return "", err
	}
	return fileName, nil

}
