package service

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"go-common/library/ecode"
	"go-common/library/log"

	xecode "go-gateway/app/web-svr/native-page/ecode"
	natGRPC "go-gateway/app/web-svr/native-page/interface/api"
	natmdl "go-gateway/app/web-svr/native-page/job/internal/model"

	"gopkg.in/gomail.v2"
)

func (s *Service) NewTopicPage(c context.Context) {
	log.Info("NewTopicPage start")
	if !s.dao.GetCfg().Mail.Switch { //只有线上环境需要发送邮件
		return
	}
	// 获取运营发起话题活动的creators
	page, creators, err := s.dao.NewTopicPage(c)
	if err != nil {
		log.Error("NewTopicPage get page data error(%v)", err)
		return
	}
	//给creators发送邮件
	base := &natmdl.Base{
		Host:    s.dao.GetCfg().Mail.Host,
		Port:    s.dao.GetCfg().Mail.Port,
		Address: s.dao.GetCfg().Mail.Address,
		Pwd:     s.dao.GetCfg().Mail.Pwd,
		Name:    s.dao.GetCfg().Mail.Name,
	}
	var toAddress []*natmdl.Address
	for v := range creators {
		if v == "" || v == "system" || v == "admin" {
			continue
		}
		toAddress = append(toAddress, &natmdl.Address{
			Address: fmt.Sprintf("%s@bilibili.com", v),
			Name:    v,
		})
	}
	if len(toAddress) == 0 {
		log.Info("NewTopicPage toAddress is empty")
		return
	}
	if len(page) == 0 {
		log.Info("NewTopicPage page is empty")
		return
	}
	subject := fmt.Sprintf("【%d月%d日】NA活动创建周知", time.Now().Month(), time.Now().Day())
	content := "Dear all，<br/>"
	content += "现将24小时内创建的活动周知如下：<br/>"
	yDay := time.Now().AddDate(0, 0, -1).Format("2006年01月02日")
	nowDay := time.Now().Format("2006年01月02日")
	content += fmt.Sprintf("创建时间范围：%s19点 ~ %s19点<br/>", yDay, nowDay)
	content += "NA活动列表：<br/>"
	content += `<style>
table,
td {
    border: 1px solid #000000;
    border-spacing: 1px;
}
thead,
tfoot {
    background-color: #E6E6E6;
    color: #000000;
    font-weight: bold;
}
</style>`
	content += "<table><thead><tr><td>活动页id</td><td>话题活动名</td><td>当前状态</td><td>上线时间</td><td>创建人</td><td>活动页链接</td></tr></thead><tbody>"
	for _, v := range page {
		if v == nil {
			continue
		}
		var state string
		switch v.State {
		case natGRPC.WaitForOnline:
			state = "待上线"
		case natGRPC.OnlineState:
			state = "已上线"
		case natGRPC.OfflineState:
			state = "已下线"
		default:
			state = "未知"
		}
		var stime string
		if int64(v.Stime) < 1 {
			stime = "0000年00月00日 00:00:00"
		} else {
			stime = v.Stime.Time().Format("2006年01月02日 15:04:05")
		}
		url := fmt.Sprintf("<a href=\"https://www.bilibili.com/blackboard/dynamic/%d\"> https://www.bilibili.com/blackboard/dynamic/%d</a>", v.ID, v.ID)
		content += fmt.Sprintf("<tr><td>%d</td><td>%s</td><td>%s</td><td>%s</td><td>%s</td><td>%s</td></tr>", v.ID, v.Title, state, stime, v.Creator, url)
	}
	content += "</tbody></table>"
	// setnx 防多发
	_, locked, err := s.dao.Lock(c, "nat_job_send_mail", s.dao.GetCfg().Expire.SendMailLockExpire)
	if err != nil {
		log.Error("NewTopicPage setnx error(%v)", err)
		return
	}
	if !locked {
		log.Error("NewTopicPage Fail to get nat_job_send_mail lock, lock has been taken")
		return
	}
	maxNum := s.dao.GetCfg().Mail.MaxNum
	if maxNum <= 0 { //防止配置文件出错，出现死循环
		maxNum = 5
	}
	for i := 0; i < len(toAddress); i += maxNum {
		var subIDs []*natmdl.Address
		if i+maxNum > len(toAddress) {
			subIDs = toAddress[i:]
		} else {
			subIDs = toAddress[i : i+maxNum]
		}
		mail := &natmdl.Mail{
			ToAddresses: subIDs,
			Subject:     subject,
			Body:        content,
			Type:        natmdl.TypeTextHTML,
		}
		//邮件发送失败，跳过处理
		if err = s.SendMail(c, mail, base, nil); err != nil {
			log.Error("NewTopicPage SendMail error(%v)", err)
			continue
		}
	}
	log.Info("NewTopicPage send success,creator(%d),page(%d)", len(creators), len(page))
}

// SendMail send mail
func (s *Service) SendMail(c context.Context, m *natmdl.Mail, base *natmdl.Base, attach *natmdl.Attach) (err error) {
	var (
		toUsers  []string
		ccUsers  []string
		bccUsers []string
		msg      = gomail.NewMessage()
	)
	msg.SetAddressHeader("From", base.Address, base.Name) // 发件人
	for _, ads := range m.ToAddresses {
		toUsers = append(toUsers, msg.FormatAddress(ads.Address, ads.Name))
	}

	for _, ads := range m.CcAddresses {
		ccUsers = append(ccUsers, msg.FormatAddress(ads.Address, ads.Name))
	}

	for _, ads := range m.BccAddresses {
		bccUsers = append(bccUsers, msg.FormatAddress(ads.Address, ads.Name))
	}
	msg.SetHeader("To", toUsers...)
	msg.SetHeader("Subject", m.Subject) // 主题

	if len(ccUsers) > 0 {
		msg.SetHeader("Cc", ccUsers...)
	}
	if len(bccUsers) > 0 {
		msg.SetHeader("Bcc", bccUsers...)
	}

	if m.Type == natmdl.TypeTextHTML {
		msg.SetBody("text/html", m.Body)
	} else {
		msg.SetBody("text/plain", m.Body)
	}
	// 附件处理
	if attach != nil {
		tmpSavePath := filepath.Join(os.TempDir(), "mail_tmp")
		err = os.MkdirAll(tmpSavePath, 0755)
		if err != nil {
			log.Error("os.MkdirAll error(%v)", err)
			return
		}
		destFilePath := filepath.Join(tmpSavePath, attach.Name)
		destFile, cErr := os.Create(destFilePath)
		if cErr != nil {
			log.Error("os.Create(%s) error(%v)", destFilePath, cErr)
			return cErr
		}
		defer os.RemoveAll(tmpSavePath)
		orginFile, err := os.Open(attach.File)
		if err != nil {
			log.Error("os.Open(%s) error(%v)", attach.File, err)
			return err
		}
		defer orginFile.Close()

		_, err = io.Copy(destFile, orginFile)
		if err != nil {
			log.Error("io.Copy() error(%v)", err)
			return err
		}
		msg.Attach(destFilePath)

	}
	d := gomail.NewDialer(
		base.Host,
		base.Port,
		base.Address,
		base.Pwd,
	)
	if err = d.DialAndSend(msg); err != nil {
		log.Error("Send mail Fail(%v) diff(%s)", msg, err)
		return
	}
	return
}

func (s *Service) OfflinePage(c context.Context) {
	log.Warn("OfflinePage start")
	list, err := s.dao.OfflinePage(c)
	if err != nil {
		log.Error("OfflinePage servie error(%v)", err)
		return
	}
	if len(list) > 0 {
		s.unbindUpSpace(c, list)
	}
	log.Info("OfflinePage success")
}

func (s *Service) unbindUpSpace(c context.Context, pages []*natGRPC.NativePage) {
	mid2PID := make(map[int64]int64, len(pages))
	for _, page := range pages {
		if page.RelatedUid == 0 || page.FromType != natmdl.PageFromUid {
			continue
		}
		mid2PID[page.RelatedUid] = page.ID
	}
	if len(mid2PID) == 0 {
		return
	}
	for mid, pageID := range mid2PID {
		pass := false
		mid := mid
		pageID := pageID
		success, err := func() (bool, error) {
			var err error
			defer func() {
				if err != nil {
					return
				}
				_ = s.dao.ResetUserSpace(c, mid, pageID, natGRPC.USpaceOfflineNormal)
			}()
			if err = s.dao.SpaceOffline(c, mid, pageID, natmdl.TabTypeUpAct); err != nil {
				// up暂未配置空间，直接跳过
				if err == ecode.NothingFound || err == xecode.UpBindOtherPage {
					pass = true
				}
				return false, err
			}
			success, err := s.dao.UpActivityTab(c, mid, 0, "", pageID)
			if err != nil {
				return false, err
			}
			return success, nil
		}()
		if pass {
			continue
		}
		if err != nil {
			log.Error("日志告警 Native活动过期通知空间tab解绑失败, mid=%+v pageID=%+v error=%+v", mid, pageID, err)
			continue
		}
		if !success {
			log.Error("日志告警 Native活动过期通知空间tab解绑失败, mid=%+v pageID=%+v", mid, pageID)
			continue
		}
		log.Info("Success to offline space, mid=%+v pageID=%+v", mid, pageID)
	}
}
