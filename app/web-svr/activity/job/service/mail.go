package service

import (
	"context"
	"encoding/csv"
	"io"
	"os"
	"path/filepath"

	"go-common/library/log"
	"go-gateway/app/web-svr/activity/job/model/mail"
	mdlmail "go-gateway/app/web-svr/activity/job/model/mail"

	"github.com/pkg/errors"
	gomail "gopkg.in/gomail.v2"
)

// activityCreateCsvAndSend 创建csv文件并发送
func (s *Service) activityCreateCsvAndSend(c context.Context, filePath, fileName string, subject string, base *mdlmail.Base, toAddress, ccAddress, bccAddress []*mdlmail.Address, categoryHeader []string, data [][]string) error {
	err := os.MkdirAll(filePath, 0755)
	if err != nil {
		log.Error("os.MkdirAll error(%v)", err)
		return err
	}
	f, err := os.Create(filePath + fileName)
	if err != nil {
		err = errors.Wrapf(err, "s.createCsv")
		log.Error("s.createCsv: error(%v) fileName %v", err, fileName)
	}
	defer os.RemoveAll(filePath)
	defer f.Close()
	f.WriteString("\xEF\xBB\xBF")
	w := csv.NewWriter(f)
	w.Write(categoryHeader)
	w.WriteAll(data) //写入数据
	err = s.mailFileConfig(c, base, toAddress, ccAddress, bccAddress, subject, mdlmail.TypeTextHTML, &mdlmail.Attach{Name: fileName, File: filePath + fileName})
	if err != nil {
		log.Error("s.mailFile: error(%v) fileName %v", err, fileName)

	}
	return nil
}

// mailFile 邮件发送
func (s *Service) mailFileConfig(c context.Context, base *mdlmail.Base, toAddress, ccAddress, bccAddress []*mdlmail.Address, subject string, mailType mdlmail.Type, attach *mdlmail.Attach) error {
	mail := &mdlmail.Mail{
		ToAddresses:  toAddress,
		CcAddresses:  ccAddress,
		BccAddresses: bccAddress,
		Subject:      subject,
		Type:         mailType,
	}
	return s.SendMail(c, mail, base, attach)
}

// SendMail send mail
func (s *Service) SendMail(c context.Context, m *mail.Mail, base *mail.Base, attach *mail.Attach) (err error) {
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

	if m.Type == mail.TypeTextHTML {
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

func (s *Service) SendTextMail(c context.Context, to []*mdlmail.Address, subject, content string) error {
	return s.SendMail(c, &mdlmail.Mail{
		ToAddresses: to,
		Subject:     subject,
		Type:        mdlmail.TypeTextPlain,
		Body:        content,
	}, &mdlmail.Base{
		Host:    s.c.Mail.Host,
		Port:    s.c.Mail.Port,
		Address: s.c.Mail.Address,
		Pwd:     s.c.Mail.Pwd,
		Name:    s.c.Mail.Name,
	}, nil)
}
